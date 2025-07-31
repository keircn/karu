package scraper

import (
	"context"

	"github.com/keircn/karu/pkg/errors"
	"github.com/keircn/karu/pkg/graphql"
)

type EpisodeData struct {
	Data struct {
		Show struct {
			AvailableEpisodesDetail struct {
				Sub []string `json:"sub"`
				Dub []string `json:"dub"`
			} `json:"availableEpisodesDetail"`
		} `json:"show"`
	} `json:"data"`
}

const EpisodesQuery = `query ($showId: String!) {
	show(_id: $showId) {
		_id
		availableEpisodesDetail
	}
}`

func (c *Client) GetEpisodes(ctx context.Context, showID string) ([]string, error) {
	initCaches()

	cacheKey := generateCacheKey("episodes", map[string]interface{}{"showId": showID})

	if cached, found := episodeCache.Get(cacheKey); found {
		return cached.([]string), nil
	}

	var episodes []string

	err := executeWithFallback(func(baseURL string) error {
		qb := graphql.NewQueryBuilder(baseURL, c.httpClient).
			SetQuery(EpisodesQuery).
			AddVariable("showId", showID)

		var episodeData EpisodeData
		if err := qb.Execute(ctx, &episodeData); err != nil {
			return err
		}

		episodes = episodeData.Data.Show.AvailableEpisodesDetail.Sub
		return nil
	})

	if err != nil {
		return nil, errors.Wrap(err, errors.ScrapingError, "failed to get episodes")
	}

	episodeCache.Set(cacheKey, episodes)
	return episodes, nil
}

func GetEpisodes(showID string) ([]string, error) {
	return defaultClient.GetEpisodes(context.Background(), showID)
}
