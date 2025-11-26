package client

import (
	"context"
	"fmt"
	pb "mangahub/proto/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GRPCClient handles gRPC client functionality for the mangahubCLI application.
type GRPCClient struct {
	conn   *grpc.ClientConn
	client pb.MangaServiceClient
}

// NewGRPCClient creates a new gRPC client and establishes a connection
func NewGRPCClient(address string) (*GRPCClient, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	client := pb.NewMangaServiceClient(conn)

	return &GRPCClient{
		conn:   conn,
		client: client,
	}, nil
}

// Close closes the gRPC connection
func (c *GRPCClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// GetManga retrieves manga details by ID
func (c *GRPCClient) GetManga(ctx context.Context, mangaID int64) (*pb.Manga, error) {
	req := &pb.GetMangaRequest{
		MangaId: mangaID,
	}

	resp, err := c.client.GetManga(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get manga: %w", err)
	}

	return resp.GetManga(), nil
}

// SearchManga searches for manga by query string
func (c *GRPCClient) SearchManga(ctx context.Context, query string, limit, offset int32) ([]*pb.Manga, int64, error) {
	req := &pb.SearchRequest{
		Query:  query,
		Limit:  limit,
		Offset: offset,
	}

	resp, err := c.client.SearchManga(ctx, req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search manga: %w", err)
	}

	return resp.GetMangas(), resp.GetTotalCount(), nil
}

// UpdateProgress updates user's reading progress
func (c *GRPCClient) UpdateProgress(ctx context.Context, userID string, mangaID int64, mangaTitle string, chapter int32, status string) error {
	req := &pb.UpdateProgressRequest{
		UserId:     userID,
		MangaId:    mangaID,
		MangaTitle: mangaTitle,
		Chapter:    chapter,
		Status:     status,
	}

	resp, err := c.client.UpdateProgress(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to update progress: %w", err)
	}

	if !resp.GetSuccess() {
		return fmt.Errorf("update failed: %s", resp.GetMessage())
	}

	return nil
}
