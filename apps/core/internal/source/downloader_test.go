package source

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

type testLogger struct{}

func (l *testLogger) Info(format string, args ...interface{})  {}
func (l *testLogger) Error(format string, args ...interface{}) {}
func (l *testLogger) Debug(format string, args ...interface{}) {}

func TestDownloader_DownloadFileWithMirrors_SuccessOnFirst(t *testing.T) {
	dir := t.TempDir()
	d := &Downloader{downloadDir: dir, logger: &testLogger{}, client: http.DefaultClient}

	// Start a test server that always succeeds
	ts := httpTestServerWithContent("hello world", http.StatusOK)
	defer ts.Close()

	url := ts.URL + "/file.txt"
	path, err := d.DownloadFileWithMirrors([]string{url}, 2)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	data, _ := os.ReadFile(path)
	if string(data) != "hello world" {
		t.Errorf("unexpected file content: %s", string(data))
	}
}

func TestDownloader_DownloadFileWithMirrors_FallbackToMirror(t *testing.T) {
	dir := t.TempDir()
	d := &Downloader{downloadDir: dir, logger: &testLogger{}, client: http.DefaultClient}

	bad := "http://127.0.0.1:0/doesnotexist"
	ts := httpTestServerWithContent("mirror ok", http.StatusOK)
	defer ts.Close()
	good := ts.URL + "/mirror.txt"

	path, err := d.DownloadFileWithMirrors([]string{bad, good}, 2)
	if err != nil {
		t.Fatalf("expected fallback to mirror, got error: %v", err)
	}
	data, _ := os.ReadFile(path)
	if string(data) != "mirror ok" {
		t.Errorf("unexpected file content: %s", string(data))
	}
}

func TestDownloader_DownloadFileWithMirrors_AllFail(t *testing.T) {
	dir := t.TempDir()
	d := &Downloader{downloadDir: dir, logger: &testLogger{}, client: http.DefaultClient}

	bad1 := "http://127.0.0.1:0/one"
	bad2 := "http://127.0.0.1:0/two"
	_, err := d.DownloadFileWithMirrors([]string{bad1, bad2}, 2)
	if err == nil || !strings.Contains(err.Error(), "all mirrors failed") {
		t.Errorf("expected all mirrors to fail, got: %v", err)
	}
}

// Helper: HTTP test server using httptest
func httpTestServerWithContent(content string, status int) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		io.WriteString(w, content)
	}))
	return ts
}

func TestNewDownloader(t *testing.T) {
	dir := t.TempDir()
	logger := &testLogger{}
	d := NewDownloader(dir, logger)
	if d == nil || d.downloadDir != dir || d.logger != logger {
		t.Errorf("NewDownloader did not initialize correctly")
	}
}

func TestDownloader_DownloadFile(t *testing.T) {
	dir := t.TempDir()
	d := NewDownloader(dir, &testLogger{})
	ts := httpTestServerWithContent("test", http.StatusOK)
	defer ts.Close()
	path, err := d.DownloadFile(ts.URL + "/file.txt")
	if err != nil {
		t.Fatalf("DownloadFile failed: %v", err)
	}
	data, _ := os.ReadFile(path)
	if string(data) != "test" {
		t.Errorf("unexpected content: %s", string(data))
	}
}

func TestDownloader_VerifyChecksum_success(t *testing.T) {
	dir := t.TempDir()
	d := NewDownloader(dir, &testLogger{})
	filePath := dir + "/test.txt"
	os.WriteFile(filePath, []byte("hello"), 0644)
	// SHA256 of "hello"
	expectedChecksum := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	err := d.VerifyChecksum(filePath, expectedChecksum)
	if err != nil {
		t.Errorf("VerifyChecksum failed: %v", err)
	}
}

func TestDownloader_VerifyChecksum_mismatch(t *testing.T) {
	dir := t.TempDir()
	d := NewDownloader(dir, &testLogger{})
	filePath := dir + "/test.txt"
	os.WriteFile(filePath, []byte("hello"), 0644)
	err := d.VerifyChecksum(filePath, "wrongchecksum")
	if err == nil || !strings.Contains(err.Error(), "checksum mismatch") {
		t.Errorf("expected checksum mismatch error, got: %v", err)
	}
}

func TestDownloader_GetDownloadSize(t *testing.T) {
	dir := t.TempDir()
	d := NewDownloader(dir, &testLogger{})
	os.WriteFile(dir+"/file1.txt", []byte("12345"), 0644)
	os.WriteFile(dir+"/file2.txt", []byte("67890"), 0644)
	size, err := d.GetDownloadSize()
	if err != nil {
		t.Fatalf("GetDownloadSize failed: %v", err)
	}
	if size != 10 {
		t.Errorf("expected size 10, got %d", size)
	}
}

func TestDownloader_EvictOldCache(t *testing.T) {
	dir := t.TempDir()
	d := NewDownloader(dir, &testLogger{})
	// Create a file with old metadata
	filePath := dir + "/old.txt"
	os.WriteFile(filePath, []byte("old content"), 0644)
	// Evict with 0 TTL should remove everything with metadata
	// Since we don't have metadata, this should not remove the file
	err := d.EvictOldCache(0)
	if err != nil {
		t.Errorf("EvictOldCache failed: %v", err)
	}
	// File should still exist (no metadata to evict)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("file was evicted without metadata")
	}
}
