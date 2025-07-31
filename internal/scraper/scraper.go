package scraper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const (
	apiURL = "https://api.allanime.day/api"
)

type GraphQLQuery struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

type SearchResult struct {
	Data struct {
		Shows struct {
			Edges []struct {
				ID                string `json:"_id"`
				Name              string `json:"name"`
				AvailableEpisodes struct {
					Sub int `json:"sub"`
					Dub int `json:"dub"`
				} `json:"availableEpisodes"`
			} `json:"edges"`
		} `json:"shows"`
	} `json:"data"`
}

func buildSearchQuery(query string) GraphQLQuery {
	return GraphQLQuery{
		Query: `query($search: SearchInput, $limit: Int, $page: Int, $translationType: VaildTranslationTypeEnumType, $countryOrigin: VaildCountryOriginEnumType) {
			shows(
				search: $search
				limit: $limit
				page: $page
				translationType: $translationType
				countryOrigin: $countryOrigin
			) {
				edges {
					_id
					name
					availableEpisodes
					__typename
				}
			}
		}`,
		Variables: map[string]any{
			"search": map[string]any{
				"allowAdult":   false,
				"allowUnknown": false,
				"query":        query,
			},
			"limit":           40,
			"page":            1,
			"translationType": "sub",
			"countryOrigin":   "ALL",
		},
	}
}

func buildTrendingQuery() GraphQLQuery {
	return GraphQLQuery{
		Query: `query($limit: Int, $page: Int, $translationType: VaildTranslationTypeEnumType, $countryOrigin: VaildCountryOriginEnumType) {
			shows(
				limit: $limit
				page: $page
				translationType: $translationType
				countryOrigin: $countryOrigin
			) {
				edges {
					_id
					name
					availableEpisodes
					__typename
				}
			}
		}`,
		Variables: map[string]any{
			"limit":           20,
			"page":            1,
			"translationType": "sub",
			"countryOrigin":   "JP",
		},
	}
}

func buildPopularQuery() GraphQLQuery {
	return GraphQLQuery{
		Query: `query($limit: Int, $page: Int, $translationType: VaildTranslationTypeEnumType, $countryOrigin: VaildCountryOriginEnumType) {
			shows(
				limit: $limit
				page: $page
				translationType: $translationType
				countryOrigin: $countryOrigin
			) {
				edges {
					_id
					name
					availableEpisodes
					__typename
				}
			}
		}`,
		Variables: map[string]any{
			"limit":           20,
			"page":            3,
			"translationType": "sub",
			"countryOrigin":   "JP",
		},
	}
}

func executeQuery(gqlQuery GraphQLQuery) ([]Anime, error) {
	initCaches()

	cacheKey := generateCacheKey(gqlQuery.Query, gqlQuery.Variables)

	if cached, found := searchCache.Get(cacheKey); found {
		return cached.([]Anime), nil
	}

	var animes []Anime

	err := executeWithFallback(func(baseURL string) error {
		jsonData, err := json.Marshal(gqlQuery)
		if err != nil {
			return err
		}

		req, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return err
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/121.0")
		req.Header.Set("Referer", "https://allanime.to")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		var searchResult SearchResult
		if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
			return err
		}

		animes = nil
		for _, edge := range searchResult.Data.Shows.Edges {
			animes = append(animes, Anime{
				Title:    edge.Name,
				URL:      fmt.Sprintf("https://allanime.to/anime/%s", edge.ID),
				Episodes: fmt.Sprintf("%d", edge.AvailableEpisodes.Sub),
			})
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	searchCache.Set(cacheKey, animes)
	return animes, nil
}
func Search(query string) ([]Anime, error) {
	return executeQuery(buildSearchQuery(query))
}

func GetTrending() ([]Anime, error) {
	return executeQuery(buildTrendingQuery())
}

func GetPopular() ([]Anime, error) {
	return executeQuery(buildPopularQuery())
}

func DownloadEpisode(showID, episode, outputPath string) error {
	videoURL, err := GetVideoURL(showID, episode)
	if err != nil {
		return fmt.Errorf("failed to get video URL: %w", err)
	}

	if videoURL == "" {
		return fmt.Errorf("no video URL found for episode %s", episode)
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create download directory: %w", err)
	}

	resp, err := http.Get(videoURL)
	if err != nil {
		return fmt.Errorf("failed to download video: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download video: status %d", resp.StatusCode)
	}

	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write video data: %w", err)
	}

	return nil
}
