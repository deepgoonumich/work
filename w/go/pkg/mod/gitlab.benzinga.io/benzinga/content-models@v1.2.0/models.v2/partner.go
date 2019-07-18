package models

import "time"

// PartnerMeta contains information for a partner.
type PartnerMeta struct {
	ID        string
	VersionID string

	UpdatedAt   time.Time
	PublishedAt time.Time

	Resource  string
	Copyright string
	Contact   string
}
