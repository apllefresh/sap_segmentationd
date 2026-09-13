package logx

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

const logDir = "log"
const logFileName = "segmentation_import.log"

func Init() error {
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return fmt.Errorf("create log dir: %w", err)
	}

	path := filepath.Join(logDir, logFileName)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}

	log.SetOutput(io.MultiWriter(os.Stdout, f))
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	return nil
}
