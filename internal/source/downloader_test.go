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
