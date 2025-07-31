package scraper

import (
	"context"
	"fmt"
	"time"

	"github.com/keircn/karu/pkg/errors"
	"github.com/keircn/karu/pkg/graphql"
	"github.com/keircn/karu/pkg/http"
)

const apiURL = "https://api.allanime.day/api"

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

type Client struct {
	httpClient   *http.Client
	queryBuilder *graphql.QueryBuilder
}

func NewClient() *Client {
	httpClient := http.NewClient(
		http.WithTimeout(10*time.Second),
		http.WithUserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/121.0"),
		http.WithReferer("https://allanime.to"),
	)

	queryBuilder := graphql.NewQueryBuilder(apiURL, httpClient)

	return &Client{
		httpClient:   httpClient,
		queryBuilder: queryBuilder,
	}
}

func (c *Client) executeShowsQuery(ctx context.Context, qb *graphql.QueryBuilder) ([]Anime, error) {
	initCaches()

	query := qb.Build()
	cacheKey := generateCacheKey(query.Query, query.Variables)

	if cached, found := searchCache.Get(cacheKey); found {
		return cached.([]Anime), nil
	}

	var result SearchResult
	err := executeWithFallback(func(baseURL string) error {
		qb := graphql.NewQueryBuilder(baseURL, c.httpClient).
			SetQuery(query.Query)

		for key, value := range query.Variables {
			qb.AddVariable(key, value)
		}

		return qb.Execute(ctx, &result)
	})

	if err != nil {
		return nil, errors.Wrap(err, errors.ScrapingError, "failed to execute shows query")
	}

	animes := make([]Anime, 0, len(result.Data.Shows.Edges))
	for _, edge := range result.Data.Shows.Edges {
		animes = append(animes, Anime{
			Title:    edge.Name,
			URL:      fmt.Sprintf("https://allanime.to/anime/%s", edge.ID),
			Episodes: fmt.Sprintf("%d", edge.AvailableEpisodes.Sub),
		})
	}

	searchCache.Set(cacheKey, animes)
	return animes, nil
}

func (c *Client) Search(ctx context.Context, query string) ([]Anime, error) {
	qb := graphql.NewQueryBuilder(apiURL, c.httpClient).
		SetQuery(graphql.ShowsQuery).
		AddSearchInput(query, false, false).
		AddPagination(40, 1).
		AddTranslationType("sub").
		AddCountryOrigin("ALL")

	return c.executeShowsQuery(ctx, qb)
}

func (c *Client) GetTrending(ctx context.Context) ([]Anime, error) {
	qb := graphql.NewQueryBuilder(apiURL, c.httpClient).
		SetQuery(graphql.ShowsQuery).
		AddPagination(20, 1).
		AddTranslationType("sub").
		AddCountryOrigin("JP")

	return c.executeShowsQuery(ctx, qb)
}

func (c *Client) GetPopular(ctx context.Context) ([]Anime, error) {
	qb := graphql.NewQueryBuilder(apiURL, c.httpClient).
		SetQuery(graphql.ShowsQuery).
		AddPagination(20, 3).
		AddTranslationType("sub").
		AddCountryOrigin("JP")

	return c.executeShowsQuery(ctx, qb)
}

func (c *Client) DownloadEpisode(ctx context.Context, showID, episode, outputPath string) error {
	videoURL, err := GetVideoURL(showID, episode)
	if err != nil {
		return errors.Wrapf(err, errors.ScrapingError, "failed to get video URL for episode %s", episode)
	}

	if videoURL == "" {
		return errors.New(errors.ScrapingError, fmt.Sprintf("no video URL found for episode %s", episode))
	}

	return c.httpClient.DownloadFile(ctx, videoURL, outputPath)
}

var defaultClient = NewClient()

func Search(query string) ([]Anime, error) {
	return defaultClient.Search(context.Background(), query)
}

func GetTrending() ([]Anime, error) {
	return defaultClient.GetTrending(context.Background())
}

func GetPopular() ([]Anime, error) {
	return defaultClient.GetPopular(context.Background())
}

func DownloadEpisode(showID, episode, outputPath string) error {
	return defaultClient.DownloadEpisode(context.Background(), showID, episode, outputPath)
}
