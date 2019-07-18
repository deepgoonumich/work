package ravenpack

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"gitlab.benzinga.io/benzinga/content-models/models"
	"gitlab.benzinga.io/benzinga/ftp-engine/config"
	"gitlab.benzinga.io/benzinga/ftp-engine/process"
	"gitlab.benzinga.io/benzinga/ftp-engine/rstore"
	"gitlab.benzinga.io/benzinga/reference-service/reference"
)

const (
	timeFormat = "Mon, 02 Jan 06 15:04:05 -700"
	xmlHeader  = `<?xml version="1.0" encoding="utf-8" ?>`
)

type Processor struct {
	cfg     *config.Config
	rClient *rstore.Client
	log     *zap.Logger
}

func NewRavenpackProcessor(cfg *config.Config, r *rstore.Client, log *zap.Logger) *Processor {
	return &Processor{cfg, r, log}
}

func (p *Processor) Convert(e *models.Event) (*process.Output, error) {

	nodeID := strconv.Itoa(e.Content.NodeID)
	timestamp := e.Content.CreatedAt.Format(timeFormat)

	author := func() string {
		if e.Content.Author == "" {
			return "Benzinga"
		}
		return e.Content.Author
	}()

	contentType := func() process.ContentType {
		if t, ok := process.ContentTypeMappings[e.Content.Type]; ok {
			return t
		}
		return ""
	}()

	base := func() string {
		if contentType == process.PressRelease {
			return "https://www.benzinga.com/export/feed/ravenpack_pr1/" + nodeID + ".xml"
		}
		if contentType == process.Story {
			return "https://www.benzinga.com/export/feed/ravenpack_realtime1/" + nodeID + ".xml"
		}
		return ""
	}()

	rXML := XML{
		Version: "2.0",
		Base:    base,
		Dc:      "http://purl.org/dc/elements/1.1/",
		Bz:      "http://www.benzinga.com/feed-ns-bz/1.0/",
		Channel: Channel{
			Title:    "Benzinga",
			Language: "en",
			Item: Item{
				Title:       e.Content.Title,
				Link:        `https://www.benzinga.com/node/` + nodeID,
				Description: process.RewriteBodyTickerPaths("https://benzinga.com", e.Content.Body),
				PubDate:     timestamp,
				Creator:     author,
				GUID: ItemGUID{
					IsPermaLink: "false",
					Text:        nodeID + " at " + "http://benzinga.com",
				},
				ID:           nodeID,
				RevisionID:   strconv.Itoa(e.Content.VersionID),
				RevisionDate: timestamp,
				Type: ItemType{
					Text: contentType.String(),
				},
				Tickers: p.getTickers(e),
			},
		},
	}

	// Handle Categories
	if contentType == process.PressRelease {
		rXML.Channel.Item.Categories = []ItemCategory{
			ItemCategory{
				Domain: "https://www.benzinga.com/press-releases",
				Text:   "Press Releases",
			},
		}
	} else {
		rXML.Channel.Item.Categories = getCategories(e)
	}

	rXML.Channel.Item.Categories = append(rXML.Channel.Item.Categories, ItemCategory{
		Domain: "publisher",
		Text:   "Benzinga",
	})

	if strings.EqualFold(e.Content.Type, "story") && e.Content.PartnerURL == "" {
		rXML.Channel.Item.Type.FirstRun = "1"
	} else {
		rXML.Channel.Item.Type.FirstRun = "0"
	}

	if e.Content.IsBzPost {
		rXML.Channel.Item.Type.Bz = "1"
	} else {
		rXML.Channel.Item.Type.Bz = "0"
	}

	if e.Content.IsBzProPost {
		rXML.Channel.Item.Type.Pro = "1"
	} else {
		rXML.Channel.Item.Type.Pro = "0"
	}

	rXMLBytes, err := xml.MarshalIndent(&rXML, "", " ")
	if err != nil {
		return nil, err
	}

	filename, err := newFilename(e)
	if err != nil {
		return nil, err
	}

	data := bytes.NewBufferString(xmlHeader + "\n")
	if _, err := data.Write(rXMLBytes); err != nil {
		return nil, err
	}

	output := &process.Output{
		Filename: filename,
		Data:     data,
	}

	return output.CalculateChecksumSize(), nil
}

func (p *Processor) getTickers(e *models.Event) (tickers []ItemTicker) {

	for i := 0; i < len(e.Content.Tickers); i++ {

		symbolSplit := strings.Split(e.Content.Tickers[i].Name, ":")

		ctx := context.TODO()

		var tickerData *reference.Instrument
		if len(symbolSplit) > 1 {
			res, err := p.rClient.GetSymbolExchange(ctx, symbolSplit[1], symbolSplit[0])
			if err != nil {
				p.log.Error("Get Ticker by Symbol-Exchange Error", zap.Error(err))
			}
			tickerData = res
		} else {
			res, err := p.rClient.GetSymbolCurrency(ctx, symbolSplit[0], "USD") // default to USD for tickers without an exchange
			if err != nil {
				p.log.Error("Get Ticker by Symbol-Currency Error", zap.Error(err))
			}
			tickerData = res
		}

		t := ItemTicker{
			Text:      e.Content.Tickers[i].Name,
			Sentiment: "0",
		}

		if tickerData != nil {
			t.ISIN = tickerData.ISIN
			t.Exchange = tickerData.Exchange
		}

		if e.Content.Tickers[i].Primary {
			t.Primary = "1"
		} else {
			t.Primary = "0"
		}

		tickers = append(tickers, t)
	}

	return tickers
}

func getCategories(e *models.Event) (categories []ItemCategory) {

	for i := 0; i < len(e.Content.Channels); i++ {
		cat := ItemCategory{
			Domain: fmt.Sprintf("https://www.benzinga.com/taxonomy/term/%v", e.Content.Channels[i].ID),
			Text:   e.Content.Channels[i].Name,
		}
		categories = append(categories, cat)
	}

	for i := 0; i < len(e.Content.Tickers); i++ {
		cat := ItemCategory{
			Domain: "stock-symbol",
			Text:   e.Content.Tickers[i].Name,
		}
		stockCat := ItemCategory{
			Domain: "https://www.benzinga.com/stock/" + strings.ToLower(e.Content.Tickers[i].Name),
			Text:   e.Content.Tickers[i].Name,
		}
		categories = append(categories, cat, stockCat)
	}

	return categories
}

func newFilename(e *models.Event) (string, error) {
	var filename strings.Builder
	if _, err := filename.WriteString("benzinga_"); err != nil {
		return "", err
	}
	if _, err := filename.WriteString(strconv.Itoa(e.Content.NodeID)); err != nil {
		return "", err
	}
	if _, err := filename.WriteString("_" + strconv.FormatInt(e.Content.UpdatedAt.Unix(), 10)); err != nil {
		return "", err
	}
	if _, err := filename.WriteString("_rss2.xml"); err != nil {
		return "", err
	}
	// ex. benzinga_12345_rss2.xml
	return filename.String(), nil
}
