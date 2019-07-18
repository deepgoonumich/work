package models

import (
	"testing"
)

// Static assertions about generated code.
var (
	_ = IsMsgpable(&PartnerMeta{})
	_ = IsMsgpable(&SectorMeta{})
	_ = IsMsgpable(&SECMeta{})
)

func TestPartnerMeta(t *testing.T) {
	AssertRoundTripSafe(t, &PartnerMeta{
		ID: "test",
	})

	AssertRoundTripSafe(t, &SectorMeta{
		SIC:         []SICSector{{IndustryCode: 1234}},
		NAICS:       []NAICSSector{{IndustryCode: 123456}},
		Morningstar: []MorningstarSector{{IndustryCode: 12345678}},
	})

	AssertRoundTripSafe(t, &SECMeta{
		AccessionNumber: "1",
	})
}
