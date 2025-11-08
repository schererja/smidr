package bitbake

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/schererja/smidr/internal/config"
	"github.com/schererja/smidr/internal/container"
	"github.com/schererja/smidr/pkg/logger"
)

type mockContainerManager struct {
	execCalls []struct {
		cmd      []string
		duration time.Duration
	}
	returnResult container.ExecResult
	returnErr    error
}

func (m *mockContainerManager) Exec(ctx context.Context, containerID string, cmd []string, timeout time.Duration) (container.ExecResult, error) {
	m.execCalls = append(m.execCalls, struct {
		cmd      []string
		duration time.Duration
	}{cmd, timeout})
	return m.returnResult, m.returnErr
}

func (m *mockContainerManager) ExecStream(ctx context.Context, containerID string, cmd []string, timeout time.Duration) (container.ExecResult, error) {
	return m.returnResult, m.returnErr
}

func (m *mockContainerManager) CopyFromContainer(ctx context.Context, containerID, containerPath, hostPath string) error {
	return nil
}

func (m *mockContainerManager) CreateContainer(ctx context.Context, cfg container.ContainerConfig) (string, error) {
	return "mock-container-id", nil
}

func (m *mockContainerManager) ImageExists(ctx context.Context, imageName string) bool {
	return true
}

func (m *mockContainerManager) PullImage(ctx context.Context, image string) error {
	return nil
}

func (m *mockContainerManager) RemoveContainer(ctx context.Context, containerID string, force bool) error {
	return nil
}

func (m *mockContainerManager) StartContainer(ctx context.Context, containerID string) error {
	return nil
}

func (m *mockContainerManager) StopContainer(ctx context.Context, containerID string, timeout time.Duration) error {
	return nil
}

func TestNewBuildExecutor(t *testing.T) {
	log := logger.NewLogger()
	cfg := &config.Config{}
	mgr := &mockContainerManager{}
	be := NewBuildExecutor(cfg, mgr, "cid", "/tmp/workspace", log)
	if be.config != cfg || be.containerMgr != mgr || be.containerID != "cid" || be.workspaceDir != "/tmp/workspace" {
		t.Errorf("NewBuildExecutor did not set fields correctly")
	}
}

func TestBuildLogWriter_WriteLog(t *testing.T) {
	var plain, jsonl testWriter
	w := &BuildLogWriter{PlainWriter: &plain, JSONLWriter: &jsonl}
	w.WriteLog("stdout", "hello world")
	if plain.written == 0 || jsonl.written == 0 {
		t.Errorf("WriteLog did not write to both writers")
	}
}

type testWriter struct{ written int }

func (w *testWriter) Write(p []byte) (int, error) {
	w.written += len(p)
	return len(p), nil
}

func TestBuildExecutor_setupBuildEnvironment_error(t *testing.T) {
	log := logger.NewLogger()

	mgr := &mockContainerManager{returnErr: errors.New("fail")}
	be := NewBuildExecutor(&config.Config{}, mgr, "cid", "/tmp", log)
	err := be.setupBuildEnvironment(context.Background())
	if err == nil {
		t.Errorf("expected error from setupBuildEnvironment, got nil")
	}
}

func TestBuildExecutor_sourceEnvironment_success(t *testing.T) {
	log := logger.NewLogger()
	mgr := &mockContainerManager{
		returnResult: container.ExecResult{ExitCode: 0},
	}
	be := NewBuildExecutor(&config.Config{}, mgr, "cid", "/tmp", log)
	err := be.sourceEnvironment(context.Background())
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestBuildExecutor_sourceEnvironment_failure(t *testing.T) {
	log := logger.NewLogger()

	mgr := &mockContainerManager{
		returnResult: container.ExecResult{ExitCode: 1, Stderr: []byte("fail")},
	}
	be := NewBuildExecutor(&config.Config{}, mgr, "cid", "/tmp", log)
	err := be.sourceEnvironment(context.Background())
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestBuildExecutor_generateLocalConfContent_defaults(t *testing.T) {
	log := logger.NewLogger()

	be := NewBuildExecutor(&config.Config{}, nil, "cid", "/tmp", log)
	conf := be.generateLocalConfContent()
	if !strings.Contains(conf, "BB_NUMBER_THREADS") || !strings.Contains(conf, "PARALLEL_MAKE") {
		t.Errorf("local.conf content missing expected settings: %s", conf)
	}
}

func TestBuildExecutor_generateLocalConfContent_Advanced(t *testing.T) {
	log := logger.NewLogger()

	cfg := &config.Config{}
	cfg.Advanced.SStateMirrors = "file://.* file:///cache/sstate/PATH"
	cfg.Advanced.PreMirrors = "https?://.*/.* http://mirror.local/"
	cfg.Advanced.NoNetwork = true
	cfg.Advanced.FetchPremirrorOnly = true
	be := NewBuildExecutor(cfg, nil, "cid", "/tmp", log)
	conf := be.generateLocalConfContent()
	if !strings.Contains(conf, "SSTATE_MIRRORS = \"file://.* file:///cache/sstate/PATH\"") {
		t.Fatalf("expected SSTATE_MIRRORS from config, got:\n%s", conf)
	}
	if !strings.Contains(conf, "PREMIRRORS = \"") {
		t.Fatalf("expected PREMIRRORS in conf, got:\n%s", conf)
	}
	if !strings.Contains(conf, "BB_NO_NETWORK = \"1\"") {
		t.Fatalf("expected BB_NO_NETWORK in conf, got:\n%s", conf)
	}
	if !strings.Contains(conf, "BB_FETCH_PREMIRRORONLY = \"1\"") {
		t.Fatalf("expected BB_FETCH_PREMIRRORONLY in conf, got:\n%s", conf)
	}
}

type fakeLogWriter struct {
	lines []string
}

func (f *fakeLogWriter) WriteLog(stream, line string) {
	f.lines = append(f.lines, stream+":"+line)
}

func TestBuildExecutor_executeBitbake_fetchFail(t *testing.T) {
	mgr := &mockContainerManager{
		returnResult: container.ExecResult{ExitCode: 1, Stdout: []byte("fetch fail"), Stderr: []byte("err")},
		returnErr:    errors.New("fetch error"),
	}
	cfg := &config.Config{Build: config.BuildConfig{Image: "core-image-minimal"}}
	log := logger.NewLogger()
	be := NewBuildExecutor(cfg, mgr, "cid", "/tmp", log)
	logWriter := &BuildLogWriter{PlainWriter: nil, JSONLWriter: nil}
	res, err := be.executeBitbake(context.Background(), logWriter)
	if err == nil || res.Success {
		t.Errorf("expected fetch error, got success")
	}
}

func TestBuildExecutor_generateBBLayersConfContent(t *testing.T) {
	// Mock container manager that simulates finding layer.conf files
	mgr := &mockContainerManager{
		returnResult: container.ExecResult{
			ExitCode: 0,
			Stdout:   []byte("/home/builder/layers/poky/meta/conf/layer.conf\n/home/builder/layers/poky/meta-poky/conf/layer.conf"),
		},
	}
	cfg := &config.Config{
		Layers: []config.Layer{
			{Name: "poky"},
		},
		YoctoSeries: "scarthgap",
	}
	log := logger.NewLogger()
	be := NewBuildExecutor(cfg, mgr, "cid", "/tmp", log)
	content := be.generateBBLayersConfContent()
	if !strings.Contains(content, "BBLAYERS") {
		t.Errorf("bblayers.conf content missing BBLAYERS: %s", content)
	}
}

func TestBuildExecutor_ExecuteBuild_setupFail(t *testing.T) {
	mgr := &mockContainerManager{
		returnErr: errors.New("setup failed"),
	}
	cfg := &config.Config{Build: config.BuildConfig{Image: "core-image-minimal"}}
	log := logger.NewLogger()
	be := NewBuildExecutor(cfg, mgr, "cid", "/tmp", log)
	res, err := be.ExecuteBuild(context.Background(), nil)
	if err == nil {
		t.Errorf("expected error from ExecuteBuild, got nil")
	}
	if res.Success {
		t.Errorf("expected build to fail")
	}
}
