package main

import (
	"context"
	"fmt"
	"os"

	"sap_segmentation/internal/config"
	"sap_segmentation/internal/erp"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("load config: %v\n", err)
		os.Exit(1)
	}

	client := erp.NewClient(cfg)
	if err := client.RunAll(context.Background()); err != nil {
		fmt.Printf("erp client: %v\n", err)
		os.Exit(1)
	}
}
