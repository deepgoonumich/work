//msgp:shim bson.ObjectId as:[]byte using:IDToBytes/IDFromBytes

package models

// PartnerMeta contains the content metadata.
type PartnerMeta struct {
	ID         string `json:"ID" bson:"ID"`
	RevisionID string `json:"revision_id" bson:"revision_id"`

	Updated   Time `json:"updated" bson:"updated"`
	Published Time `json:"published" bson:"published"`

	Resource   string   `json:"resource" bson:"resource"`
	Copyright  string   `json:"copyright" bson:"copyright"`
	Contact    string   `json:"contact" bson:"contact"`
	Taxonomies []string `json:"taxonomies,omitempty" bson:"taxonomies,omitempty"`
}

// PartnerTaxonomyMeta specifies partner taxonomies.
type PartnerTaxonomyMeta struct {
	Taxonomies []PartnerTaxonomy
}

// PartnerTaxonomy represents any partner specific taxonomy
type PartnerTaxonomy struct {
	Symbol   string `json:"symbol" bson:"symbol"`
	Name     string `json:"name" bson:"name"`
	Exchange string `json:"exchange" bson:"exchange"`

	ISIN  string `json:"isin" bson:"isin"`
	CIK   string `json:"cik" bson:"cik"`
	CUSIP string `json:"cusip" bson:"cusip"`

	Order string `json:"order" bson:"order"`
}

// SectorMeta contains sector data.
type SectorMeta struct {
	SIC         []SICSector         `json:",omitempty"`
	NAICS       []NAICSSector       `json:",omitempty"`
	Morningstar []MorningstarSector `json:",omitempty"`
}

// SICSector contains SIC sector data.
type SICSector struct {
	// 4-digit SIC industry code
	IndustryCode int
	Industry     string

	// 3-digit SIC industry group code (first 3 digits)
	IndustryGroup int

	// 2-digit SIC major group code (first 2 digits)
	MajorGroup int

	// Top-level divison description.
	Division string
}

// NAICSSector contains NAICS sector data.
type NAICSSector struct {
	// 6-digit NAICS national industry code
	NationalIndustryCode int
	NationalIndustry     string

	// 5-digit NAICS industry code
	IndustryCode int
	Industry     string

	// 4-digit NAICS industry group code (first 4 digits)
	IndustryGroupCode int
	IndustryGroup     string

	// 3-digit NAICS subsector code (first 3 digits)
	SubSectorCode int
	SubSector     string

	// 2-digit NAICS sector code (first 2 digits)
	SectorCode int
	Sector     string
}

// MorningstarSector contains MorningstarSector sector data.
type MorningstarSector struct {
	// 8-digit Morningstar Industry code.
	IndustryCode int
	Industry     string

	// 5-digit industry group code. (first 5 digits)
	IndustryGroupCode int
	IndustryGroup     string

	// 3-digit sector code. (first 3 digits)
	SectorCode int
	Sector     string

	// 1-digit super sector code. (first digit)
	SuperSectorCode int
	SuperSector     string
}

// SECMeta contains metadata for SEC documents.
type SECMeta struct {
	AccessionNumber string
}
