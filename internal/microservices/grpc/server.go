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
	search "mangahub/internal/search"
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
		Source:        "local",
		SourceUrl:     "",
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
	query := req.GetQuery()
	limit := int(req.GetLimit())
	offset := int(req.GetOffset())
	if limit <= 0 || limit > 20 {
		limit = 20 // hard cap
	}

	// 1) Search local DB (pagination applied)
	localAll, err := s.mangaRepo.SearchByTitle(ctx, query)
	if err != nil {
		return nil, err
	}
	totalLocal := len(localAll)
	start := offset
	if start > totalLocal {
		start = totalLocal
	}
	end := start + limit
	if end > totalLocal {
		end = totalLocal
	}
	localPage := localAll[start:end]

	// Convert local to proto now
	var localPB []*pb.Manga
	for _, m := range localPage {
		pm := modelToProto(&m)
		localPB = append(localPB, pm)
	}

	// 2) Always fetch external to ensure links are included
	externals := search.FetchExternalSources(ctx, query, limit)

	// 3) Merge results with simple policy to ensure external visibility
	// - Take up to half of the limit from local first
	// - Fill the rest from externals
	// - If still not full, backfill with remaining local
	half := limit / 2
	if half == 0 {
		half = 1
	}

	seen := make(map[string]struct{})
	var out []*pb.Manga

	// helper to push unique items
	push := func(items []*pb.Manga) {
		for _, it := range items {
			if len(out) >= limit {
				return
			}
			key := fmt.Sprintf("%s|%s|%d", it.GetSource(), it.GetTitle(), it.GetId())
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, it)
		}
	}

	// Take up to half from local
	if len(localPB) > half {
		push(localPB[:half])
	} else {
		push(localPB)
	}
	// Fill with externals
	push(externals)
	// Backfill with remaining local
	if len(out) < limit {
		if len(localPB) > half {
			push(localPB[half:])
		}
	}

	resp := &pb.SearchResponse{
		Mangas:     out,
		TotalCount: int64(len(out)),
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
