package scraper

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type Stream struct {
	Link          string `json:"link"`
	Hls           bool   `json:"hls"`
	ResolutionStr string `json:"resolutionStr"`
	SourceName    string `json:"sourceName"`
}

type VideoStream struct {
	SourceUrl  string `json:"sourceUrl"`
	SourceName string `json:"sourceName"`
}

type VideoData struct {
	SourceUrls []VideoStream `json:"sourceUrls"`
}

type VideoResult struct {
	Data struct {
		Episode struct {
			SourceUrls []VideoStream `json:"sourceUrls"`
		} `json:"episode"`
	} `json:"data"`
}

func buildVideoQuery(showID, episode string) GraphQLQuery {
	return GraphQLQuery{
		Query: `query ($showId: String!, $translationType: VaildTranslationTypeEnumType!, $episodeString: String!) {
			episode(
				showId: $showId
				translationType: $translationType
				episodeString: $episodeString
			) {
				episodeString
				sourceUrls
			}
		}`,
		Variables: map[string]any{
			"showId":          showID,
			"translationType": "sub",
			"episodeString":   episode,
		},
	}
}

func getVideoSourceURLs(showID, episode string) (*VideoResult, error) {
	var videoResult VideoResult

	err := executeWithFallback(func(baseURL string) error {
		gqlQuery := buildVideoQuery(showID, episode)
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

		if err := json.NewDecoder(resp.Body).Decode(&videoResult); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &videoResult, nil
}
