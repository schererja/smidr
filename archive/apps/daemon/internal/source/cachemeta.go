package source

import (
	"encoding/json"
	"os"
	"time"
)

// CacheMeta is metadata for cache entries (repos or downloads)
type CacheMeta struct {
	LastAccess time.Time `json:"last_access"`
}

// writeCacheMeta writes last-access metadata to a file (repo or download)
func writeCacheMeta(metaPath string) error {
	meta := CacheMeta{LastAccess: time.Now()}
	data, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	return os.WriteFile(metaPath, data, 0644)
}

// readCacheMeta reads last-access metadata from a file (repo or download)
func readCacheMeta(metaPath string) (CacheMeta, error) {
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return CacheMeta{}, err
	}
	var meta CacheMeta
	err = json.Unmarshal(data, &meta)
	return meta, err
}
