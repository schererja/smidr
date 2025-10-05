package source

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// Downloader handles downloading source tarballs and files
type Downloader struct {
	downloadDir string
	logger      Logger
	client      *http.Client
}

// NewDownloader creates a new source downloader
func NewDownloader(downloadDir string, logger Logger) *Downloader {
	return &Downloader{
		downloadDir: downloadDir,
		logger:      logger,
		client:      &http.Client{},
	}
}

// DownloadFile downloads a file from a URL to the download directory
func (d *Downloader) DownloadFile(url string) (string, error) {
	if err := os.MkdirAll(d.downloadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create download directory: %w", err)
	}

	// Generate filename from URL
	filename := filepath.Base(url)
	destPath := filepath.Join(d.downloadDir, filename)

	// Check if already downloaded
	if _, err := os.Stat(destPath); err == nil {
		d.logger.Info("File %s already cached", filename)
		return destPath, nil
	}

	d.logger.Info("Downloading %s", url)

	// Download file
	resp, err := d.client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download %s: HTTP %d", url, resp.StatusCode)
	}

	// Create destination file
	out, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	// Copy data
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		os.Remove(destPath) // Clean up partial download
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	d.logger.Info("Successfully downloaded %s", filename)
	return destPath, nil
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
