package models

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tinylib/msgp/msgp"
	"gopkg.in/mgo.v2/bson"
)

func TestContentMeta(t *testing.T) {
	AssertRoundTripSafe(t, &Content{ID: bson.NewObjectId()})

	AssertRoundTripSafe(t, &Content{
		ID: bson.NewObjectId(),
		Meta: Meta{
			SectorV2: &SectorMeta{
				SIC:         []SICSector{{IndustryCode: 1}},
				NAICS:       []NAICSSector{{IndustryCode: 1}},
				Morningstar: []MorningstarSector{{IndustryCode: 1}},
			},
			PartnerTaxonomy: &PartnerTaxonomyMeta{
				Taxonomies: []PartnerTaxonomy{
					{Symbol: "TEST"},
				},
			},
			Ext: map[string]interface{}{
				"Test1": float64(64),
			},
		},
	})

	assert.Panics(t, func() {
		meta := Meta{
			Ext: map[string]interface{}{
				"SectorV2": map[string]interface{}{
					"This": "ShouldPanic",
				},
			},
		}

		meta.EncodeMsg(msgp.NewWriter(ioutil.Discard))
	})
}
