package model

type Segmentation struct {
	ID           int64  `db:"id" json:"-"`
	AddressSapID string `db:"address_sap_id" json:"address_sap_id"`
	AdrSegment   string `db:"adr_segment" json:"adr_segment"`
	SegmentID    int64  `db:"segment_id" json:"segment_id"`
}
