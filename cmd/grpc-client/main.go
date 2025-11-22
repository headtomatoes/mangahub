package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "mangahub/proto/pb"
)

func main() {
	// Parse command line flags
	serverAddr := flag.String("server", "localhost:8083", "gRPC server address")
	action := flag.String("action", "get", "Action to perform: get, search")
	mangaID := flag.Int64("id", 1, "Manga ID for get action")
	query := flag.String("query", "", "Search query for search action")
	limit := flag.Int("limit", 10, "Limit for search results")
	offset := flag.Int("offset", 0, "Offset for search results")
	flag.Parse()

	// Connect to gRPC server
	conn, err := grpc.NewClient(*serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewMangaServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	switch *action {
	case "get":
		testGetManga(ctx, client, *mangaID)
	case "search":
		testSearchManga(ctx, client, *query, int32(*limit), int32(*offset))
	default:
		log.Fatalf("Unknown action: %s (use 'get' or 'search')", *action)
	}
}

// testGetManga tests UC-014: Retrieve Manga via gRPC
func testGetManga(ctx context.Context, client pb.MangaServiceClient, mangaID int64) {
	fmt.Println("=== UC-014: Testing GetManga ===")
	fmt.Printf("Requesting manga ID: %d\n\n", mangaID)

	// Step 1: Client service calls GetManga gRPC method
	req := &pb.GetMangaRequest{
		MangaId: mangaID,
	}

	// Steps 2-5: Server receives, queries DB, constructs response, returns data
	resp, err := client.GetManga(ctx, req)
	if err != nil {
		log.Fatalf("GetManga failed: %v", err)
	}

	// Verify and display result
	if resp.Manga == nil {
		log.Fatal("Response manga is nil")
	}

	fmt.Println("Successfully retrieved manga!")
	fmt.Println("\n📖 Manga Details:")
	fmt.Printf("  ID:           %d\n", resp.Manga.Id)
	fmt.Printf("  Title:        %s\n", resp.Manga.Title)
	fmt.Printf("  Description:  %s\n", resp.Manga.Description)
	fmt.Printf("  Authors:      %v\n", resp.Manga.Authors)
	fmt.Printf("  Genres:       %v\n", resp.Manga.Genres)
	fmt.Printf("  Cover URL:    %s\n", resp.Manga.CoverUrl)
	fmt.Printf("  Chapters:     %d\n", resp.Manga.ChaptersCount)
	fmt.Println("\nUC-014: PASSED - All 5 steps completed successfully")
}

// testSearchManga tests UC-015: Search Manga via gRPC
func testSearchManga(ctx context.Context, client pb.MangaServiceClient, query string, limit, offset int32) {
	fmt.Println("=== UC-015: Testing SearchManga ===")
	fmt.Printf("Search Query: '%s'\n", query)
	fmt.Printf("Limit: %d, Offset: %d\n\n", limit, offset)

	// Step 1: Client calls SearchManga with search criteria
	req := &pb.SearchRequest{
		Query:  query,
		Limit:  limit,
		Offset: offset,
	}

	// Steps 2-4: Server processes params, executes query, constructs response
	resp, err := client.SearchManga(ctx, req)
	if err != nil {
		log.Fatalf("SearchManga failed: %v", err)
	}

	// Verify and display results
	fmt.Printf("Successfully searched manga!\n")
	fmt.Printf("Total Results: %d\n", resp.TotalCount)
	fmt.Printf("Page Results: %d\n\n", len(resp.Mangas))

	if len(resp.Mangas) == 0 {
		fmt.Println("No results found")
		fmt.Println("UC-015: PASSED - Search executed successfully (empty result)")
		return
	}

	fmt.Println("📚 Search Results:")
	for i, manga := range resp.Mangas {
		fmt.Printf("\n%d. %s (ID: %d)\n", i+1, manga.Title, manga.Id)
		fmt.Printf("   Authors: %v\n", manga.Authors)
		fmt.Printf("   Genres:  %v\n", manga.Genres)
		if manga.Description != "" {
			desc := manga.Description
			if len(desc) > 100 {
				desc = desc[:100] + "..."
			}
			fmt.Printf("   Description: %s\n", desc)
		}
		fmt.Printf("   Chapters: %d\n", manga.ChaptersCount)
	}

	fmt.Println("\nUC-015: PASSED - All 4 steps completed successfully")

	// Additional validation
	if int32(len(resp.Mangas)) > limit {
		log.Printf("Warning: Returned %d results, exceeds limit of %d", len(resp.Mangas), limit)
	}
}
