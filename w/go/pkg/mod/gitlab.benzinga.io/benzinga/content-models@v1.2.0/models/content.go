//msgp:shim bson.ObjectId as:[]byte using:IDToBytes/IDFromBytes
//msgp:ignore Node

package models

import (
	"bufio"
	"bytes"
	"strconv"

	"github.com/pierrec/lz4"
	"github.com/tinylib/msgp/msgp"
	"gopkg.in/mgo.v2/bson"
)

// Vocab represents the type of taxonomy a category is.
type Vocab int

// These are the various valid types of taxonomy.
const (
	VocabChannel    Vocab = 1
	VocabEntity     Vocab = 3
	VocabTicker     Vocab = 2
	VocabMutualFund Vocab = 8
	VocabArticle    Vocab = 9
	VocabCareers    Vocab = 10
	VocabFutures    Vocab = 11
)

// Category contains the structure for a taxonomy object.
type Category struct {
	ID          int            `json:"tid" bson:"tid"`
	Vocab       Vocab          `json:"vid" bson:"vid"`
	Name        string         `json:"name" bson:"name"`
	Description string         `json:"description" bson:"description"`
	Primary     bool           `json:"primary" bson:"primary"`
	Price       string         `json:"price,omitempty" bson:"price,omitempty"`
	Volume      int            `json:"volume,omitempty" bson:"volume,omitempty"`
	Sectors     map[string]int `json:"sectors,omitempty" bson:"sectors,omitempty"`
}

// NodeQuote contains the structure for a stock quote at the time of a story.
type NodeQuote struct {
	ID     string `json:"tid" bson:"tid"`
	Price  string `json:"price" bson:"price"`
	Volume string `json:"volume" bson:"volume"`
}

// Node contains the basic content node information.
type Node struct {
	ID        PHPObjID `json:"_id" bson:"_id"`
	EventID   PHPObjID `json:"eid,omitempty" bson:"eid,omitempty"`
	NodeID    int      `json:"nid" bson:"nid"`
	UserID    int      `json:"uid" bson:"uid"`
	VersionID int      `json:"vid" bson:"vid"`
	Type      string   `json:"type" bson:"type"`
	Status    int      `json:"status" bson:"status"`

	CreatedAt float64 `json:"created" bson:"created"`
	UpdatedAt float64 `json:"changed" bson:"changed"`

	Title      string `json:"title" bson:"title"`
	Body       string `json:"body" bson:"body"`
	Author     string `json:"name" bson:"name"`
	PartnerURL string `json:"syndicate_url" bson:"syndicate_url"`
	Path       string `json:"path" bson:"path"`

	Taxonomy       []Category  `json:"taxonomy" bson:"taxonomy"`
	PrimaryTickers []DrupalInt `json:"field_primay_tickers_custom" bson:"field_primay_tickers_custom"`
	Price          []NodeQuote `json:"price" bson:"price"`

	Image  []FieldImage `json:"field_image" bson:"field_image"`
	Assets []Asset      `json:"assets" bson:"assets"`

	IsBzPost        int `json:"is_bz_post" bson:"is_bz_post"`
	IsBzProPost     int `json:"is_bzpro_post" bson:"is_bzpro_post"`
	DoNotDistribute int `json:"is_dnd" bson:"is_dnd"`

	Sentiment Sentiment `json:"field_rate_bull_bear" bson:"field_rate_bull_bear"`

	Meta Meta `json:"meta" bson:"meta"`
}

// Quote contains a stock quote.
type Quote struct {
	Price  string
	Volume int
}

// Content contains the processed node structure.
type Content struct {
	ID        bson.ObjectId `json:"ID" bson:"ID"`
	EventID   bson.ObjectId `json:"EventID,omitempty" bson:"EventID,omitempty"`
	NodeID    int           `json:"NodeID" bson:"NodeID"`
	UserID    int           `json:"UserID" bson:"UserID"`
	VersionID int           `json:"VersionID" bson:"VersionID"`
	Type      string        `json:"Type" bson:"Type"`
	Published bool          `json:"Published" bson:"Published"`

	CreatedAt Time `json:"CreatedAt" bson:"CreatedAt"`
	UpdatedAt Time `json:"UpdatedAt" bson:"UpdatedAt"`

	Title      string  `json:"Title" bson:"Title"`
	Body       string  `json:"Body" bson:"Body"`
	Author     string  `json:"name" bson:"name"`
	Assets     []Asset `json:"assets,omitempty" bson:"assets,omitempty"`
	PartnerURL string  `json:"PartnerURL" bson:"PartnerURL"`
	TeaserText string  `json:"TeaserText" bson:"TeaserText"`

	Tags     []Category       `json:"Tags,omitempty" bson:"Tags,omitempty"`
	Tickers  []Category       `json:"Tickers,omitempty" bson:"Tickers,omitempty"`
	Futures  []Category       `json:"Futures,omitempty" bson:"Futures,omitempty"`
	Channels []Category       `json:"Channels,omitempty" bson:"Channels,omitempty"`
	Quotes   map[string]Quote `json:"Quotes,omitempty" bson:"Quotes,omitempty"`

	IsBzPost        bool `json:"IsBzPost" bson:"IsBzPost"`
	IsBzProPost     bool `json:"IsBzProPost" bson:"IsBzProPost"`
	DoNotDistribute bool `json:"DoNotDistribute" bson:"DoNotDistribute"`
	Sentiment       int  `json:"Sentiment" bson:"Sentiment"`

	Meta Meta `json:"Meta" bson:"Meta"`
}

// AssetType ...
type AssetType string

var (
	// ImageAsset ...
	ImageAsset AssetType = "image"
	// VideoAsset ...
	VideoAsset AssetType = "video"
)

// Asset ...
type Asset struct {
	Type      AssetType `json:"type" bson:"type"`
	Title     string    `json:"title" bson:"title"`
	MIME      string    `json:"mime" bson:"filemime"`
	Primary   bool      `json:"primary" bson:"primary"`
	Copyright string    `json:"copyright" bson:"copyright"`
	URL       string    `json:"url" bson:"url"`

	Attributes *AssetAttributes `json:"attributes" bson:"attributes"`
}

// AssetAttributes ...
type AssetAttributes struct {
	FID             string           `json:"fid" bson:"fid"`
	Filename        string           `json:"filename" bson:"filename"`
	Filepath        string           `json:"filepath" bson:"filepath"`
	Filesize        int64            `json:"filesize" bson:"filesize"`
	ImageAttributes *ImageAttributes `json:"image_attributes,omitempty" bson:"image_attributes,omitempty"`
}

// ImageAttributes ...
type ImageAttributes struct {
	Resolution struct {
		Height int `json:"height" bson:"height"`
		Width  int `json:"width" bson:"width"`
		DPI    int `json:"dpi" bson:"dpi"`
	} `json:"resolution" bson:"resolution"`
	AltTitle string `json:"alt_title" bson:"alt_title"`
}

// FieldImage ...
type FieldImage struct {
	UploadIdentifier string `json:"UPLOAD_IDENTIFIER" bson:"UPLOAD_IDENTIFIER"`
	Fid              string `json:"fid" bson:"fid"`
	Data             struct {
		Alt   string `json:"alt" bson:"alt"`
		Title string `json:"title" bson:"title"`
	} `json:"data" bson:"data"`
	List      string      `json:"list" bson:"list"`
	UID       string      `json:"uid" bson:"uid"`
	Filename  string      `json:"filename" bson:"filename"`
	Filepath  string      `json:"filepath" bson:"filepath"`
	Filemime  string      `json:"filemime" bson:"filemime"`
	Filesize  int64       `json:"filesize,string" bson:"filesize"`
	Status    interface{} `json:"status" bson:"status"`
	Timestamp interface{} `json:"timestamp" bson:"timestamp"`
	Alt       string      `json:"alt" bson:"alt"`
	Title     string      `json:"title" bson:"title"`
}

// AsAsset ...
func (f FieldImage) AsAsset() Asset {
	a := Asset{
		Type:  ImageAsset,
		Title: f.Title,
		MIME:  f.Filemime,
		Attributes: &AssetAttributes{
			FID:      f.Fid,
			Filepath: f.Filepath,
			Filename: f.Filename,
			Filesize: f.Filesize,
			ImageAttributes: &ImageAttributes{
				AltTitle: f.Alt,
			},
		},
	}

	return a
}

// AsContent returns a content structure for a node.
func (n *Node) AsContent() Content {
	tickerMap := map[int]string{}
	tickerPrimaryMap := map[int]bool{}

	content := Content{
		ID:              bson.ObjectId(n.ID),
		EventID:         bson.ObjectId(n.EventID),
		NodeID:          n.NodeID,
		UserID:          n.UserID,
		VersionID:       n.VersionID,
		Type:            n.Type,
		Published:       n.Status == 1,
		CreatedAt:       TimeFromFloat(n.CreatedAt),
		UpdatedAt:       TimeFromFloat(n.UpdatedAt),
		Title:           n.Title,
		Body:            n.Body,
		Author:          n.Author,
		PartnerURL:      n.PartnerURL,
		IsBzPost:        n.IsBzPost == 1,
		IsBzProPost:     n.IsBzProPost == 1,
		DoNotDistribute: n.DoNotDistribute == 1,
		Sentiment:       int(n.Sentiment),
		Meta:            n.Meta,
	}

	// Use First Field Image as primary asset
	// at most one field image should exist
	// TODO: May want to check type/size in future
	if len(n.Image) > 0 {
		primaryAsset := n.Image[0].AsAsset()
		primaryAsset.Primary = true
		content.Assets = append(content.Assets, primaryAsset)
	}
	if len(n.Assets) > 1 {
		content.Assets = append(content.Assets, n.Assets...)
	}

	for _, primary := range n.PrimaryTickers {
		tickerPrimaryMap[int(primary)] = true
	}

	for _, category := range n.Taxonomy {
		switch category.Vocab {
		case VocabTicker:
			_, category.Primary = tickerPrimaryMap[category.ID]
			content.Tickers = append(content.Tickers, category)
			tickerMap[category.ID] = category.Name
		case VocabChannel:
			content.Channels = append(content.Channels, category)
		case VocabFutures:
			content.Futures = append(content.Futures, category)
		default:
			content.Tags = append(content.Tags, category)
		}
	}

	if len(n.Price) > 0 {
		content.Quotes = map[string]Quote{}
	}

	// Normalize quotes.
	for _, quote := range n.Price {
		tid, _ := strconv.Atoi(quote.ID)
		vol, _ := strconv.Atoi(quote.Volume)
		symbol := tickerMap[tid]
		if symbol == "" {
			continue
		}

		content.Quotes[symbol] = Quote{
			Price:  quote.Price,
			Volume: vol,
		}
	}

	return content
}

// AsNode returns node data for a content structure.
func (c *Content) AsNode() Node {
	boolToInt := func(b bool) int {
		if b {
			return 1
		}
		return 0
	}

	taxonomy := []Category{}
	taxonomy = append(taxonomy, c.Tags...)
	taxonomy = append(taxonomy, c.Tickers...)
	taxonomy = append(taxonomy, c.Channels...)
	taxonomy = append(taxonomy, c.Futures...)

	primaryTickers := []DrupalInt{}
	for _, ticker := range c.Tickers {
		if ticker.Primary {
			primaryTickers = append(primaryTickers, DrupalInt(ticker.ID))
		}
	}

	return Node{
		ID:              PHPObjID(c.ID),
		EventID:         PHPObjID(c.EventID),
		NodeID:          c.NodeID,
		UserID:          c.UserID,
		VersionID:       c.VersionID,
		Type:            c.Type,
		Status:          boolToInt(c.Published),
		CreatedAt:       c.CreatedAt.Float(),
		UpdatedAt:       c.UpdatedAt.Float(),
		Title:           c.Title,
		Body:            c.Body,
		Author:          c.Author,
		Assets:          c.Assets,
		PartnerURL:      c.PartnerURL,
		Taxonomy:        taxonomy,
		PrimaryTickers:  primaryTickers,
		IsBzPost:        boolToInt(c.IsBzPost),
		IsBzProPost:     boolToInt(c.IsBzProPost),
		DoNotDistribute: boolToInt(c.DoNotDistribute),
		Sentiment:       Sentiment(c.Sentiment),
		Meta:            c.Meta,
	}
}

// Compress returns an Lz4-compressed MessagePack representation of the
// content object.
func (c *Content) Compress() ([]byte, error) {
	out := bytes.Buffer{}

	lz := lz4.NewWriter(&out)
	lz.Header.HighCompression = true

	// lz4.Writer is unbuffered, so ensure we buffer the data for maximum
	// compression. Otherwise, we can potentially end up with many frames.
	w := bufio.NewWriterSize(lz, 65536)
	err := msgp.Encode(w, c)
	if err != nil {
		return []byte{}, err
	}

	err = w.Flush()
	if err != nil {
		return []byte{}, err
	}

	return out.Bytes(), err
}

// Decompress returns a Content object from an Lz4-compressed
// MessagePack representation of a content object.
func (c *Content) Decompress(data []byte) error {
	lz := lz4.NewReader(bytes.NewBuffer(data))

	return msgp.Decode(lz, c)
}
