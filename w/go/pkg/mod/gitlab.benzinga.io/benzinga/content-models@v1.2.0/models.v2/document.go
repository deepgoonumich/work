package models

import (
	"bytes"
	"regexp"
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"
)

// SentimentLevel specifies how positive or negative a story is.
type SentimentLevel int

// InstrumentMeta contains the information for an Instrument.
type InstrumentMeta struct {
	ISIN   string
	Symbol string
	Price  string
	Volume int
}

// DocumentMeta contains the information associated with a Document.
type DocumentMeta struct {
	PublishedAt *time.Time
	UpdatedAt   *time.Time
	Cursor      string

	Title   string
	URL     string
	Summary string

	Sentiment   SentimentLevel
	Categories  []TaxonomyID
	Instruments []InstrumentMeta
	Partner     *PartnerMeta
	Sectors     *SectorMeta
}

// DocumentID contains identification information for a Document.
type DocumentID struct {
	UUID string
	NID  string
	VID  string
}

// DocumentIndex contains keyword index information for a Document.
type DocumentIndex struct {
	Keywords string
}

// DocumentContent contains the full body of a Document.
type DocumentContent struct {
	ContentType string
	ContentBody string
}

// Document is a full document object.
type Document struct {
	DocumentID
	DocumentMeta
	DocumentContent
}

// NewDocument creates a new Document.
func NewDocument(title string) *Document {
	uuid := bson.NewObjectId().String()
	doc := new(Document)
	doc.Title = title
	doc.UUID = uuid
	return doc
}

// SetHTML sets the body of a Document
func (doc *Document) SetHTML(body string) *Document {
	doc.ContentType = "text/html"
	doc.ContentBody = body
	return doc
}

// AddInstrument adds an instrument to the document.
func (doc *Document) AddInstrument(isin string) *Document {
	// TODO
	return doc
}

var wordre = regexp.MustCompile(`\w+`)

// ToIndexed gets an IndexedDocument for the Document.
func (doc *Document) ToIndexed() *IndexedDocument {
	indexed := new(IndexedDocument)
	indexed.DocumentID = doc.DocumentID
	indexed.DocumentMeta = doc.DocumentMeta

	// Map all of the unique words in the document.
	wordmap := map[string]struct{}{}
	for _, word := range wordre.FindAllString(doc.ContentBody, -1) {
		wordmap[strings.ToLower(word)] = struct{}{}
	}

	// Concat all of the words together into a giant field.
	var buffer bytes.Buffer
	for word := range wordmap {
		buffer.WriteString(word)
		buffer.WriteRune(' ')
	}

	indexed.DocumentIndex.Keywords = buffer.String()

	return indexed
}

// IndexedDocument is an indexed document.
type IndexedDocument struct {
	DocumentID
	DocumentMeta
	DocumentIndex
}

// ToResult gets an IndexedDocument for the Document.
func (doc *IndexedDocument) ToResult() *ResultDocument {
	result := new(ResultDocument)
	result.DocumentID = doc.DocumentID
	result.DocumentMeta = doc.DocumentMeta
	return result
}

// ResultDocument is a document result.
type ResultDocument struct {
	DocumentID
	DocumentMeta
}
