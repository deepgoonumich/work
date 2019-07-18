package models

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
	"github.com/tinylib/msgp/msgp"
	"gopkg.in/mgo.v2/bson"
)

func TestPHPObjID(t *testing.T) {
	node := Node{}
	data, err := ioutil.ReadFile("testdata/content0001.json")
	require.NoError(t, err)

	// Test unmarshaling (with real data.)
	err = json.Unmarshal(data, &node)
	require.NoError(t, err)
	assert.Equal(t, PHPObjID(bson.ObjectIdHex("50de2127c3be26ca32000000")), node.ID)

	// Test marshaling.
	obj := struct{ ID PHPObjID }{node.ID}
	data, err = json.Marshal(obj)
	require.NoError(t, err)
	assert.Equal(t, `{"ID":{"$id":"50de2127c3be26ca32000000"}}`, string(data))
}

func TestAsContent(t *testing.T) {

	t.Run("Test File content0001.json", func(t *testing.T) {
		node := Node{}
		data, err := ioutil.ReadFile("testdata/content0001.json")
		require.NoError(t, err)
		json.Unmarshal(data, &node)
		node.Meta = Meta{
			SectorV2: &SectorMeta{
				SIC: []SICSector{
					{
						IndustryCode: 1234,
					},
				},
			},
			Ext: map[string]interface{}{
				"Test": map[string]interface{}{
					"TestField": 1,
				},
			},
		}

		// Make sure the taxonomy is working.
		content := node.AsContent()
		assert.Len(t, content.Tickers, 6)
		assert.Equal(t, content.Channels[0].Name, "News")
		assert.Len(t, content.Futures, 1)
		assert.Len(t, content.Tags, 4)
		assert.Equal(t, content.Sentiment, 2)
		assert.Equal(t, content.PartnerURL, "https://www.google.com/search?q=how+to+internet")
		assert.Equal(t, content.Meta, node.Meta)
		assert.Equal(t, ImageAsset, content.Assets[0].Type)

		for _, ticker := range content.Tickers {
			if ticker.ID == 10394 || ticker.ID == 14502 {
				assert.True(t, ticker.Primary)
			}
		}

		assert.Equal(t, "Scott Rubin", content.Author)
	})
	t.Run("Test File node_data.json", func(t *testing.T) {
		node := Node{}
		data, err := ioutil.ReadFile("testdata/node_data.json")
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(data, &node), "Unmarshal should be successful")

		content := node.AsContent()
		assert.NotEmpty(t, content.Assets)
		require.NotZero(t, len(content.Assets))
		assert.NotEmpty(t, content.Assets[0].MIME)

	})
}

func TestAsNode(t *testing.T) {
	content := Content{
		Title:  "Title",
		Body:   "Body",
		Author: "Author",
		Tickers: []Category{
			{ID: 14502, Vocab: 2, Name: "F", Description: "Ford Motor Credit Company", Primary: true},
		},
		Channels: []Category{
			{ID: 57, Vocab: 1, Name: "News"},
		},
		PartnerURL: "test",
		IsBzPost:   true,
		Meta: Meta{
			SectorV2: &SectorMeta{
				SIC: []SICSector{
					{
						IndustryCode: 1234,
					},
				},
			},
			Ext: map[string]interface{}{
				"Test": map[string]interface{}{
					"TestField": 1,
				},
			},
		},
		Assets: []Asset{
			Asset{
				Type:    ImageAsset,
				Primary: true,
			},
		},
	}

	node := content.AsNode()

	assert.Equal(t, "Title", node.Title)
	assert.Equal(t, "Body", node.Body)
	assert.Equal(t, "Author", node.Author)
	assert.Len(t, node.Taxonomy, 2)
	assert.Equal(t, node.PartnerURL, "test")
	assert.Equal(t, node.IsBzPost, 1)
	assert.Len(t, node.PrimaryTickers, 1)
	assert.Equal(t, int(node.PrimaryTickers[0]), 14502)
	assert.Equal(t, node.Meta, content.Meta)
	assert.NotEmpty(t, node.Assets)
}

func TestCompress(t *testing.T) {
	content1 := Content{
		Title:     "Title",
		Body:      "Body",
		Author:    "Author",
		CreatedAt: Time{time.Now().In(time.Local)},
		UpdatedAt: Time{time.Now().In(time.Local)},
		Tickers: []Category{
			{ID: 14502, Vocab: 2, Name: "F", Description: "Ford Motor Credit Company"},
		},
		Channels: []Category{
			{ID: 57, Vocab: 1, Name: "News"},
		},
		PartnerURL: "test",
		Assets: []Asset{
			Asset{
				Type:    ImageAsset,
				Primary: true,
			},
		},
	}

	data, err := content1.Compress()
	require.NoError(t, err)

	content2 := Content{}
	err = content2.Decompress(data)
	require.NoError(t, err)

	assert.Equal(t, content1.Title, content2.Title)
	assert.Equal(t, content1.Body, content2.Body)
	assert.Equal(t, content1.Author, content2.Author)
	assert.Equal(t, content1.CreatedAt, content2.CreatedAt)
	assert.Equal(t, content1.UpdatedAt, content2.UpdatedAt)
	assert.Equal(t, content1.Tickers[0].ID, content2.Tickers[0].ID)
	assert.Equal(t, content1.Tickers[0].Vocab, content2.Tickers[0].Vocab)
	assert.Equal(t, content1.Tickers[0].Name, content2.Tickers[0].Name)
	assert.Equal(t, content1.Tickers[0].Description, content2.Tickers[0].Description)
	assert.Equal(t, content1.Channels[0].ID, content2.Channels[0].ID)
	assert.Equal(t, content1.Channels[0].Vocab, content2.Channels[0].Vocab)
	assert.Equal(t, content1.Channels[0].Name, content2.Channels[0].Name)
	assert.Equal(t, content1.PartnerURL, content2.PartnerURL)
	assert.Equal(t, content1.Assets, content2.Assets)
}

func TestMsgpBroken(t *testing.T) {
	content := Content{ID: "TEST1", EventID: "TEST2"}
	data, err := content.Compress()
	require.NoError(t, err)

	content = Content{}
	content.Decompress(data)
	assert.Equal(t, "TEST1", string(content.ID))
	assert.Equal(t, "TEST2", string(content.EventID))
}

func TestMsgpSizerBroken(t *testing.T) {
	content := Content{}
	baseSize := content.Msgsize()
	content.ID = "TEST1"
	test1Size := content.Msgsize()
	content.EventID = "TEST2"
	test2Size := content.Msgsize()

	assert.Equal(t, baseSize+5, test1Size)
	assert.Equal(t, test1Size+5, test2Size)
}

func TestMsgpCoderBroken(t *testing.T) {
	buffer := &bytes.Buffer{}
	content := Content{}
	msgp.Encode(buffer, &Content{ID: "TEST1", EventID: "TEST2"})
	msgp.Decode(buffer, &content)
	assert.Equal(t, "TEST1", string(content.ID))
	assert.Equal(t, "TEST2", string(content.EventID))
}
