package importer

import (
	"context"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"

	"sap_segmentation/internal/config"
	"sap_segmentation/internal/erp"
	"sap_segmentation/model"
)

func Run(ctx context.Context, cfg config.Config, db *sqlx.DB) error {
	client := erp.NewClient(cfg)

	var total int
	err := client.RunAll(ctx, func(segments []model.Segmentation) error {
		if err := model.UpsertBatch(ctx, db, segments); err != nil {
			log.Printf("database upsert error: %v", err)
			return fmt.Errorf("save batch: %w", err)
		}
		total += len(segments)
		log.Printf("saved batch: %d rows (total: %d)", len(segments), total)
		return nil
	})
	if err != nil {
		log.Printf("import error: %v", err)
		return err
	}

	log.Printf("import finished, total rows: %d", total)
	return nil
}
