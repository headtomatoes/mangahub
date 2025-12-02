package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	pb "mangahub/proto/pb"
)

// ExternalManga is a simplified representation unified across providers.
type ExternalManga struct {
	Title         string
	Description   string
	Authors       []string
	Genres        []string
	CoverURL      string
	ChaptersCount int32
	Source        string
	SourceURL     string
}

// FetchExternalSources queries AniList, Kitsu and MangaDex concurrently and returns up to `limit` results.
// It is best-effort: failures from one source do not abort others.
func FetchExternalSources(ctx context.Context, query string, limit int) []*pb.Manga {
	if limit <= 0 {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, 6*time.Second)
	defer cancel()

	type result struct {
		items []ExternalManga
	}
	ch := make(chan result, 3)

	// Launch goroutines
	go func() { ch <- result{items: searchAniList(ctx, query, limit)} }()
	go func() { ch <- result{items: searchKitsu(ctx, query, limit)} }()
	go func() { ch <- result{items: searchMangaDex(ctx, query, limit)} }()

	var merged []ExternalManga
outer:
	for i := 0; i < 3; i++ {
		select {
		case r := <-ch:
			merged = append(merged, r.items...)
		case <-ctx.Done():
			// timeout; stop early
			break outer
		}
	}
	// Deduplicate by Title+Source (simple) and cap to limit
	seen := make(map[string]struct{})
	var out []*pb.Manga
	for _, em := range merged {
		if len(out) >= limit {
			break
		}
		key := em.Source + "|" + em.Title
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, &pb.Manga{
			Title:         em.Title,
			Description:   em.Description,
			Authors:       em.Authors,
			Genres:        em.Genres,
			CoverUrl:      em.CoverURL,
			ChaptersCount: em.ChaptersCount,
			Source:        em.Source,
			SourceUrl:     em.SourceURL,
		})
	}
	return out
}

// ------------- AniList -------------
// Minimal GraphQL query for AniList manga (type=MANGA).
const anilistEndpoint = "https://graphql.anilist.co"

func searchAniList(ctx context.Context, query string, limit int) []ExternalManga {
	gqlQuery := `query ($search: String, $perPage: Int) { Page(page: 1, perPage: $perPage) { media(search: $search, type: MANGA) { id title { romaji english native } description(asHtml: false) genres } } }`
	payload := map[string]any{
		"query":     gqlQuery,
		"variables": map[string]any{"search": query, "perPage": min(limit, 10)}, // cap per provider
	}
	b, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, anilistEndpoint, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	var parsed struct {
		Data struct {
			Page struct {
				Media []struct {
					ID    int `json:"id"`
					Title struct {
						Romaji  string `json:"romaji"`
						English string `json:"english"`
						Native  string `json:"native"`
					} `json:"title"`
					Description string   `json:"description"`
					Genres      []string `json:"genres"`
				} `json:"media"`
			} `json:"Page"`
		} `json:"data"`
	}
	if json.Unmarshal(data, &parsed) != nil {
		return nil
	}
	var out []ExternalManga
	for _, m := range parsed.Data.Page.Media {
		title := firstNonEmpty(m.Title.English, m.Title.Romaji, m.Title.Native)
		out = append(out, ExternalManga{
			Title:       title,
			Description: stripHTML(m.Description),
			Genres:      m.Genres,
			Source:      "anilist",
			SourceURL:   fmt.Sprintf("https://anilist.co/manga/%d", m.ID),
		})
	}
	return out
}

// ------------- Kitsu -------------
const kitsuEndpoint = "https://kitsu.io/api/edge/manga"

func searchKitsu(ctx context.Context, query string, limit int) []ExternalManga {
	url := fmt.Sprintf("%s?filter[text]=%s&page[limit]=%d", kitsuEndpoint, urlQueryEscape(query), min(limit, 10))
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	var parsed struct {
		Data []struct {
			ID         string `json:"id"`
			Attributes struct {
				CanonicalTitle string `json:"canonicalTitle"`
				Titles         struct {
					EnJp string `json:"en_jp"`
					En   string `json:"en"`
					JaJp string `json:"ja_jp"`
				} `json:"titles"`
				Synopsis string   `json:"synopsis"`
				Genres   []string `json:"genres"`
			} `json:"attributes"`
		} `json:"data"`
	}
	if json.Unmarshal(body, &parsed) != nil {
		return nil
	}
	var out []ExternalManga
	for _, d := range parsed.Data {
		title := firstNonEmpty(d.Attributes.CanonicalTitle, d.Attributes.Titles.En, d.Attributes.Titles.EnJp, d.Attributes.Titles.JaJp)
		out = append(out, ExternalManga{
			Title:       title,
			Description: d.Attributes.Synopsis,
			Source:      "kitsu",
			SourceURL:   fmt.Sprintf("https://kitsu.io/manga/%s", d.ID),
		})
	}
	return out
}

// ------------- MangaDex -------------
const mangadexEndpoint = "https://api.mangadex.org/manga"

func searchMangaDex(ctx context.Context, query string, limit int) []ExternalManga {
	url := fmt.Sprintf("%s?title=%s&limit=%d&includes[]=cover_art", mangadexEndpoint, urlQueryEscape(query), min(limit, 10))
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	var parsed struct {
		Data []struct {
			ID         string `json:"id"`
			Attributes struct {
				Title       map[string]string `json:"title"`
				Description map[string]string `json:"description"`
			} `json:"attributes"`
			Relationships []struct {
				Type       string `json:"type"`
				Attributes struct {
					FileName string `json:"fileName"`
				} `json:"attributes"`
			} `json:"relationships"`
		} `json:"data"`
	}
	if json.Unmarshal(body, &parsed) != nil {
		return nil
	}
	var out []ExternalManga
	for _, d := range parsed.Data {
		// pick English or first title
		title := d.Attributes.Title["en"]
		if title == "" {
			for _, v := range d.Attributes.Title {
				title = v
				break
			}
		}
		desc := d.Attributes.Description["en"]
		if desc == "" {
			for _, v := range d.Attributes.Description {
				desc = v
				break
			}
		}
		coverURL := ""
		for _, rel := range d.Relationships {
			if rel.Type == "cover_art" {
				coverURL = fmt.Sprintf("https://uploads.mangadex.org/covers/%s/%s", d.ID, rel.Attributes.FileName)
				break
			}
		}
		out = append(out, ExternalManga{
			Title:       title,
			Description: desc,
			CoverURL:    coverURL,
			Source:      "mangadex",
			SourceURL:   fmt.Sprintf("https://mangadex.org/title/%s", d.ID),
		})
	}
	return out
}

// ---------- helpers ----------
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
func stripHTML(s string) string { // naive removal of '<' ... '>' blocks
	out := make([]rune, 0, len(s))
	inTag := false
	for _, r := range s {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			out = append(out, r)
		}
	}
	return string(out)
}
func urlQueryEscape(q string) string { return url.QueryEscape(q) }
