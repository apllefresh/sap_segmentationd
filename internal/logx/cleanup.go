package logx

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

func CleanupOld(maxAgeDays int) {
	if maxAgeDays < 0 {
		return
	}

	entries, err := os.ReadDir(logDir)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		log.Printf("log cleanup: read dir: %v", err)
		return
	}

	cutoff := time.Now().Add(-time.Duration(maxAgeDays) * 24 * time.Hour)

	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		path := filepath.Join(logDir, e.Name())
		info, err := os.Stat(path)
		if err != nil {
			log.Printf("log cleanup: stat %s: %v", path, err)
			continue
		}

		if info.ModTime().Before(cutoff) {
			if err := os.Remove(path); err != nil {
				log.Printf("log cleanup: remove %s: %v", path, err)
				continue
			}
			log.Printf("log cleanup: removed old file %s", path)
		}
	}
}
