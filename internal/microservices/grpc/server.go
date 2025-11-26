package grpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"

	pb "mangahub/proto/pb"

	models "mangahub/internal/microservices/http-api/models"
	rp "mangahub/internal/microservices/http-api/repository"
)

type MangaServiceServer struct { // internal servuce for manga operations internally(microservice GRPC server)
	pb.UnimplementedMangaServiceServer
	mangaRepo    *rp.MangaRepo
	progressRepo rp.ProgressRepository
}

func NewMangaServiceServer(
	mangaRepo *rp.MangaRepo,
	progressRepo rp.ProgressRepository,
) *MangaServiceServer {
	return &MangaServiceServer{
		mangaRepo:    mangaRepo,
		progressRepo: progressRepo,
	}
}

func modelToProto(m *models.Manga) *pb.Manga {
	if m == nil {
		return nil
	}
	// map fields and handle pointer values
	desc := ""
	if m.Description != nil {
		desc = *m.Description
	}
	var authors []string
	if m.Author != nil && *m.Author != "" {
		authors = []string{*m.Author}
	}
	var genres []string
	for _, g := range m.Genres {
		// assume models.Genre has Name field
		genres = append(genres, g.Name)
	}
	cover := ""
	if m.CoverURL != nil {
		cover = *m.CoverURL
	}
	var chapters int32
	if m.TotalChapters != nil {
		chapters = int32(*m.TotalChapters)
	}
	return &pb.Manga{
		Id:            int64(m.ID),
		Title:         m.Title,
		Description:   desc,
		Authors:       authors,
		Genres:        genres,
		CoverUrl:      cover,
		ChaptersCount: chapters,
	}
}

// GetManga implements MangaService.GetManga
func (s *MangaServiceServer) GetManga(ctx context.Context, req *pb.GetMangaRequest) (*pb.GetMangaResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("empty request")
	}
	mangaID := req.GetMangaId()
	manga, err := s.mangaRepo.GetByID(ctx, mangaID)
	if err != nil {
		return nil, err
	}
	return &pb.GetMangaResponse{
		Manga: modelToProto(manga),
	}, nil
}

// SearchManga implements MangaService.SearchManga
func (s *MangaServiceServer) SearchManga(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("empty request")
	}
	limit := int(req.GetLimit())
	offset := int(req.GetOffset())
	if limit <= 0 {
		limit = 20
	}
	// repository exposes SearchByTitle; use it and paginate results here
	all, err := s.mangaRepo.SearchByTitle(ctx, req.GetQuery())
	if err != nil {
		return nil, err
	}
	total := int64(len(all))
	start := offset
	if start > len(all) {
		start = len(all)
	}
	end := start + limit
	if end > len(all) {
		end = len(all)
	}
	page := all[start:end]

	resp := &pb.SearchResponse{
		TotalCount: total,
	}
	for _, m := range page {
		// each m is models.Manga (not pointer) from SearchByTitle — take address
		pm := modelToProto(&m)
		resp.Mangas = append(resp.Mangas, pm)
	}
	return resp, nil
}

// UpdateProgress implements MangaService.UpdateProgress
func (s *MangaServiceServer) UpdateProgress(ctx context.Context, req *pb.UpdateProgressRequest) (*pb.UpdateProgressResponse, error) {
	err := s.progressRepo.UpdateProgress(ctx, &models.UserProgress{
		UserID:         req.GetUserId(),
		MangaID:        req.GetMangaId(),
		CurrentChapter: int(req.GetChapter()),
		Status:         req.GetStatus(),
		UpdatedAt:      time.Now().UTC(),
	})
	if err != nil {
		return &pb.UpdateProgressResponse{
			Success: false,
			Message: fmt.Sprintf("failed to update: %v", err),
		}, nil
	}

	return &pb.UpdateProgressResponse{
		Success: true,
		Message: "Progress updated successfully",
	}, nil
}

// StartGRPCServer starts the gRPC server
func StartGRPCServer(addr string, mangaRepo *rp.MangaRepo, progressRepo rp.ProgressRepository) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()
	srv := NewMangaServiceServer(mangaRepo, progressRepo)
	pb.RegisterMangaServiceServer(grpcServer, srv)
	log.Printf("gRPC listening on %s", addr)
	return grpcServer.Serve(lis)
}
