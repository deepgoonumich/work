package models

// SimpleQuery is a simple query object for basic queries.
type SimpleQuery struct {
	// Range
	EarliestDate, LatestDate Time // DEPRECATED

	Before      Time
	Limit, Skip int

	// Cursor
	LastDate Time // DEPRECATED
	LastNID  int  // DEPRECATED

	// Query
	Channels    []int
	ContentType []string
	ExcludeBody bool
	Keywords    []string
	PartnerID   string
	Sectors     []string
	Sentiments  []int
	Symbols     []string
	Published   *bool
	IsBzPost    *bool
	IsBzProPost *bool
}

// SimpleQueryer provides a simple interface for searching content.
type SimpleQueryer interface {
	QuerySimple(SimpleQuery) ([]Content, error)
}
