package grpc

import (
	"context"
	pb "mangahub/proto"

	rp "mangahub/internal/microservices/http-api/repository"
	// "github.com/headtomatoes/mangahub/proto/pb"
)

type MangaServiceServer struct { // internal servuce for manga operations internally(microservice GRPC server)
	// pb.UnimplementedMangaServiceServer
	mangaRepo    rp.MangaRepo
	progressRepo rp.ProgressRepository
}

func NewMangaServiceServer(
	mangaRepo rp.MangaRepo,
	progressRepo rp.ProgressRepository,
) *MangaServiceServer {
	return &MangaServiceServer{
		mangaRepo:    mangaRepo,
		progressRepo: progressRepo,
	}
}

// GetManga implements MangaService.GetManga
func (s *MangaServiceServer) GetManga(ctx context.Context, req *pb.GetMangaRequest) (*pb.GetMangaResponse, error) {
	// Fetch manga from database
	return &pb.GetMangaResponse{}, nil
}

// SearchManga implements MangaService.SearchManga
func (s *MangaServiceServer) SearchManga(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	// Search manga in database
	return &pb.SearchResponse{}, nil
}

// UpdateProgress implements MangaService.UpdateProgress
func (s *MangaServiceServer) UpdateProgress(ctx context.Context, req *pb.UpdateProgressRequest) (*pb.UpdateProgressResponse, error) {
	// Update progress in database
	return &pb.UpdateProgressResponse{}, nil
}

// StartGRPCServer starts the gRPC server
func StartGRPCServer(addr string, mangaRepo *rp.MangaRepo, progressRepo rp.ProgressRepository) error {
	//template function to start gRPC server
	return nil
}
