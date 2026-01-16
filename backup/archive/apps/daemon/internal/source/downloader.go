package source

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/schererja/smidr/pkg/logger"
)

// Downloader handles downloading source tarballs and files
type Downloader struct {
	downloadDir string
	logger      *logger.Logger
	client      *http.Client
}

// NewDownloader creates a new source downloader
func NewDownloader(downloadDir string, logger *logger.Logger) *Downloader {
	return &Downloader{
		downloadDir: downloadDir,
		logger:      logger,
		client:      &http.Client{},
	}
}

// DownloadFile downloads a file from a URL to the download directory
func (d *Downloader) DownloadFile(url string) (string, error) {
	return d.DownloadFileWithMirrors([]string{url}, 3)
}

// EvictOldCache removes downloads not accessed within ttl
func (d *Downloader) EvictOldCache(ttl time.Duration) error {
	entries, err := os.ReadDir(d.downloadDir)
	if err != nil {
		return err
	}
	now := time.Now()
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		filePath := filepath.Join(d.downloadDir, entry.Name())
		meta, err := readCacheMeta(filePath)
		if err != nil {
			// If no meta, skip eviction (could be in use or legacy)
			continue
		}
		if now.Sub(meta.LastAccess) > ttl {
			d.logger.Info("Evicting download", slog.String("file", entry.Name()), slog.String("lastAccess", meta.LastAccess.Format(time.RFC3339)))
			_ = os.Remove(filePath)
			_ = os.Remove(filePath + ".smidr_meta.json")
		}
	}
	return nil
}

// VerifyChecksum verifies the SHA256 checksum of a file
func (d *Downloader) VerifyChecksum(filePath string, expectedChecksum string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return fmt.Errorf("failed to compute checksum: %w", err)
	}

	actualChecksum := fmt.Sprintf("%x", hash.Sum(nil))
	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil
}

// GetDownloadSize returns the total size of downloads in bytes
func (d *Downloader) GetDownloadSize() (int64, error) {
	var size int64

	err := filepath.Walk(d.downloadDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

// DownloadFileWithMirrors downloads a file from a list of URLs (mirrors) with retries and backoff.
func (d *Downloader) DownloadFileWithMirrors(urls []string, maxRetries int) (string, error) {
	var lastErr error
	for _, url := range urls {
		for attempt := range maxRetries {
			destPath, err := d.downloadFileOnce(url)
			if err == nil {
				return destPath, nil
			}
			lastErr = err
			backoff := time.Duration(1<<attempt) * 100 * time.Millisecond
			d.logger.Error("Retrying after error", err, slog.String("url", url), slog.Int("attempt", attempt+1), slog.Duration("backoff", backoff))
			time.Sleep(backoff)
		}
	}
	return "", fmt.Errorf("all mirrors failed: %w", lastErr)
}

// downloadFileOnce attempts to download a file from a single URL once.
func (d *Downloader) downloadFileOnce(url string) (string, error) {
	filename := filepath.Base(url)
	destPath := filepath.Join(d.downloadDir, filename)

	// Check if file exists and is fresh
	if _, err := os.Stat(destPath); err == nil {
		meta, err := readCacheMeta(destPath + ".smidr_meta.json")
		if err == nil {
			// TODO: Switch to using proper setter for last access time
			meta.LastAccess = time.Now()
			_ = writeCacheMeta(destPath + ".smidr_meta.json")
		}
		d.logger.Info("Cache hit for %s", slog.String("file", filename))
		return destPath, nil
	}

	resp, err := d.client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download %s: HTTP %d", url, resp.StatusCode)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		os.Remove(destPath)
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	d.logger.Info("Successfully downloaded %s", slog.String("file", filename))
	_ = writeCacheMeta(destPath + ".smidr_meta.json")
	return destPath, nil
}
