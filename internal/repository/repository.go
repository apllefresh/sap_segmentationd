package repository

import (
	"context"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"

	"sap_segmentation/internal/api"
	"sap_segmentation/internal/config"
	"sap_segmentation/model"
)

const upsertSQL = `
INSERT INTO segmentation (address_sap_id, adr_segment, segment_id)
VALUES ($1, $2, $3)
ON CONFLICT (address_sap_id) DO UPDATE SET
    adr_segment = EXCLUDED.adr_segment,
    segment_id  = EXCLUDED.segment_id`

func UpsertBatch(ctx context.Context, db *sqlx.DB, rows []model.Segmentation) error {
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

func Run(ctx context.Context, cfg config.Config, db *sqlx.DB) error {
	client := api.NewClient(cfg)

	var total int
	err := client.RunAll(ctx, func(segments []model.Segmentation) error {
		if err := UpsertBatch(ctx, db, segments); err != nil {
			log.Printf("database upsert error: %v", err)
			return fmt.Errorf("save batch: %w", err)
		}
		total += len(segments)
		log.Printf("saved batch: %d rows (total: %d)", len(segments), total)
		return nil
	})
	if err != nil {
		log.Printf("erp import error: %v", err)
		return err
	}

	log.Printf("import finished, total rows: %d", total)
	return nil
}
