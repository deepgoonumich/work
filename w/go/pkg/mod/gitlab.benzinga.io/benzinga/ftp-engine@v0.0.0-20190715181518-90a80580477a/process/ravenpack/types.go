package ravenpack

import "encoding/xml"

type XML struct {
	XMLName xml.Name `xml:"rss"`
	Text    string   `xml:",chardata"`
	Version string   `xml:"version,attr"`
	Base    string   `xml:"base,attr"`
	Dc      string   `xml:"dc,attr"`
	Bz      string   `xml:"bz,attr"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Text        string `xml:",chardata"`
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Language    string `xml:"language"`
	Item        Item   `xml:"item"`
}

type Item struct {
	Text         string         `xml:",chardata"`
	Title        string         `xml:"title"`
	Link         string         `xml:"link"`
	Description  string         `xml:"description"`
	PubDate      string         `xml:"pubDate"`
	Creator      string         `xml:"dc:creator"`
	GUID         ItemGUID       `xml:"guid"`
	Categories   []ItemCategory `xml:"category"`
	ID           string         `xml:"bz:id"`
	RevisionID   string         `xml:"bz:revisionid"`
	RevisionDate string         `xml:"bz:revisiondate"`
	Type         ItemType       `xml:"bz:type"`
	Tickers      []ItemTicker   `xml:"bz:ticker"`
}

type ItemCategory struct {
	Text   string `xml:",chardata"`
	Domain string `xml:"domain,attr"`
}

type ItemType struct {
	Text     string `xml:",chardata"`
	Bz       string `xml:"bz,attr,omitempty"`
	Pro      string `xml:"pro,attr,omitempty"`
	FirstRun string `xml:"firstrun,attr,omitempty"`
}

type ItemTicker struct {
	Text      string `xml:",chardata"`
	Primary   string `xml:"primary,attr"`
	ISIN      string `xml:"isin,attr"`
	Exchange  string `xml:"exchange,attr"`
	Sentiment string `xml:"sentiment,attr"`
}

type ItemGUID struct {
	Text        string `xml:",chardata"`
	IsPermaLink string `xml:"isPermaLink,attr"`
}
