package client

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	v1 "github.com/schererja/smidr/api/proto"
)

// Client wraps the gRPC client for the Smidr daemon
type Client struct {
	conn   *grpc.ClientConn
	client v1.SmidrClient
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
		conn:   conn,
		client: v1.NewSmidrClient(conn),
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
func (c *Client) StartBuild(ctx context.Context, configPath, target, customer string, forceClean, forceImage bool) (*v1.BuildStatus, error) {
	req := &v1.StartBuildRequest{
		Config:     configPath,
		Target:     target,
		Customer:   customer,
		ForceClean: forceClean,
		ForceImage: forceImage,
	}

	return c.client.StartBuild(ctx, req)
}

// GetBuildStatus retrieves the status of a build
func (c *Client) GetBuildStatus(ctx context.Context, buildID string) (*v1.BuildStatus, error) {
	req := &v1.BuildStatusRequest{
		BuildId: buildID,
	}

	return c.client.GetBuildStatus(ctx, req)
}

// StreamLogs streams logs from a build
func (c *Client) StreamLogs(ctx context.Context, buildID string, follow bool) (v1.Smidr_StreamLogsClient, error) {
	req := &v1.StreamLogsRequest{
		BuildId: buildID,
		Follow:  follow,
	}

	return c.client.StreamLogs(ctx, req)
}

// ListBuilds lists all builds
func (c *Client) ListBuilds(ctx context.Context, states []v1.BuildState, limit int32) (*v1.BuildsList, error) {
	req := &v1.ListBuildsRequest{
		States: states,
		Limit:  limit,
	}

	return c.client.ListBuilds(ctx, req)
}

// CancelBuild cancels a running build
func (c *Client) CancelBuild(ctx context.Context, buildID string) (*v1.CancelResult, error) {
	req := &v1.CancelBuildRequest{
		BuildId: buildID,
	}

	return c.client.CancelBuild(ctx, req)
}

// ListArtifacts lists artifacts from a completed build
func (c *Client) ListArtifacts(ctx context.Context, buildID string) (*v1.ArtifactsList, error) {
	req := &v1.ListArtifactsRequest{
		BuildId: buildID,
	}

	return c.client.ListArtifacts(ctx, req)
}
