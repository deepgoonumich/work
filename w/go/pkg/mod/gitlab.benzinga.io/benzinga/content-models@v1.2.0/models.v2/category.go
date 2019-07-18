package models

// TaxonomyType specifies a category type.
type TaxonomyType string

// Valid TaxonomyType values.
var (
	TaxonomySource  TaxonomyType = "source"
	TaxonomyChannel              = "category"
	TaxonomyTag                  = "tag"
)

// TaxonomyID specifies a taxonomical categorization.
type TaxonomyID struct {
	ID   string
	Type TaxonomyType
}
