package model

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

const upsertSQL = `
INSERT INTO segmentation (address_sap_id, adr_segment, segment_id)
VALUES ($1, $2, $3)
ON CONFLICT (address_sap_id) DO UPDATE SET
    adr_segment = EXCLUDED.adr_segment,
    segment_id  = EXCLUDED.segment_id`

type Segmentation struct {
	ID           int64  `db:"id" json:"-"`
	AddressSapID string `db:"address_sap_id" json:"address_sap_id"`
	AdrSegment   string `db:"adr_segment" json:"adr_segment"`
	SegmentID    int64  `db:"segment_id" json:"segment_id"`
}

func UpsertBatch(ctx context.Context, db *sqlx.DB, rows []Segmentation) error {
	if len(rows) == 0 {
		return nil
	}

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PreparexContext(ctx, upsertSQL)
	if err != nil {
		return fmt.Errorf("prepare upsert: %w", err)
	}
	defer stmt.Close()

	for _, row := range rows {
		if _, err := stmt.ExecContext(ctx, row.AddressSapID, row.AdrSegment, row.SegmentID); err != nil {
			return fmt.Errorf("upsert address_sap_id=%q: %w", row.AddressSapID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}
