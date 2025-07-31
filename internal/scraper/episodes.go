package scraper

import (
	"bytes"
	"encoding/json"
	"net/http"
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

func buildEpisodesQuery(showID string) GraphQLQuery {
	return GraphQLQuery{
		Query: `query ($showId: String!) {
			show(_id: $showId) {
				_id
				availableEpisodesDetail
			}
		}`,
		Variables: map[string]any{
			"showId": showID,
		},
	}
}

func GetEpisodes(showID string) ([]string, error) {
	initCaches()

	cacheKey := generateCacheKey("episodes", map[string]interface{}{"showId": showID})

	if cached, found := episodeCache.Get(cacheKey); found {
		return cached.([]string), nil
	}

	var episodes []string

	err := executeWithFallback(func(baseURL string) error {
		gqlQuery := buildEpisodesQuery(showID)
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

		var episodeData EpisodeData
		if err := json.NewDecoder(resp.Body).Decode(&episodeData); err != nil {
			return err
		}

		episodes = episodeData.Data.Show.AvailableEpisodesDetail.Sub
		return nil
	})

	if err != nil {
		return nil, err
	}

	episodeCache.Set(cacheKey, episodes)
	return episodes, nil
}
