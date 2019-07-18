package models

// SectorMeta contains associated sector data.
type SectorMeta struct {
	// SIC contains 4-digit SIC sector codes.
	SIC []string

	// NAICS contains 6-digit NAICS sector codes.
	NAICS []string

	// Morningstar contains 8-digit Morningstar sector codes.
	Morningstar []string
}
