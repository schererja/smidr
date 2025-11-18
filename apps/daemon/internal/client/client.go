package client

import (
	"context"
	"fmt"
	"time"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	v1 "github.com/schererja/smidr/pkg/smidr-sdk/v1"
)

// Client wraps the gRPC clients for the Smidr daemon
type Client struct {
	conn           *grpc.ClientConn
	buildClient    v1.BuildServiceClient
	artifactClient v1.ArtifactServiceClient
	logClient      v1.LogServiceClient
}

// NewClient creates a new client connected to the daemon at the given address
func NewClient(address string) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to daemon at %s: %w", address, err)
	}

	return &Client{
		conn:           conn,
		buildClient:    v1.NewBuildServiceClient(conn),
		artifactClient: v1.NewArtifactServiceClient(conn),
		logClient:      v1.NewLogServiceClient(conn),
	}, nil
}

// Close closes the connection to the daemon
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// StartBuild starts a new build on the daemon
func (c *Client) StartBuild(ctx context.Context, configPath, target, customer string, forceClean, forceImageRebuild bool) (*v1.BuildStatusResponse, error) {
	req := &v1.StartBuildRequest{
		Config:            configPath,
		Target:            target,
		Customer:          customer,
		ForceClean:        forceClean,
		ForceImageRebuild: forceImageRebuild,
	}

	return c.buildClient.StartBuild(ctx, req)
}

// GetBuildStatus retrieves the status of a build
func (c *Client) GetBuildStatus(ctx context.Context, buildID string) (*v1.BuildStatusResponse, error) {
	req := &v1.BuildStatusRequest{
		BuildIdentifier: &v1.BuildIdentifier{
			BuildId: buildID,
		},
	}

	return c.buildClient.GetBuildStatus(ctx, req)
}

// StreamLogs streams logs from a build
func (c *Client) StreamLogs(ctx context.Context, buildID string, follow bool) (grpc.ServerStreamingClient[v1.LogEntry], error) {
	req := &v1.StreamBuildLogsRequest{
		BuildIdentifier: &v1.BuildIdentifier{
			BuildId: buildID,
		},
		Follow: follow,
	}

	return c.logClient.StreamBuildLogs(ctx, req)
}

// ListBuilds lists all builds
func (c *Client) ListBuilds(ctx context.Context, states []v1.BuildState, pageSize int32) (*v1.ListBuildsResponse, error) {
	req := &v1.ListBuildsRequest{
		StateFilter: states,
		PageSize:    pageSize,
	}

	return c.buildClient.ListBuilds(ctx, req)
}

// CancelBuild cancels a running build
func (c *Client) CancelBuild(ctx context.Context, buildID string) (*v1.CancelBuildResponse, error) {
	req := &v1.CancelBuildRequest{
		BuildIdentifier: &v1.BuildIdentifier{
			BuildId: buildID,
		},
	}

	return c.buildClient.CancelBuild(ctx, req)
}

// ListArtifacts lists artifacts from a completed build
func (c *Client) ListArtifacts(ctx context.Context, buildID string) (*v1.ListArtifactsResponse, error) {
	req := &v1.ListArtifactsRequest{
		BuildIdentifier: &v1.BuildIdentifier{
			BuildId: buildID,
		},
	}

	return c.artifactClient.ListArtifacts(ctx, req)
}
