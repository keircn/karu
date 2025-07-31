package scraper

import (
	"context"

	"github.com/keircn/karu/pkg/errors"
	"github.com/keircn/karu/pkg/graphql"
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

const VideoQuery = `query ($showId: String!, $translationType: VaildTranslationTypeEnumType!, $episodeString: String!) {
	episode(
		showId: $showId
		translationType: $translationType
		episodeString: $episodeString
	) {
		episodeString
		sourceUrls
	}
}`

func (c *Client) getVideoSourceURLs(ctx context.Context, showID, episode string) (*VideoResult, error) {
	var videoResult VideoResult

	err := executeWithFallback(func(baseURL string) error {
		qb := graphql.NewQueryBuilder(baseURL, c.httpClient).
			SetQuery(VideoQuery).
			AddVariable("showId", showID).
			AddVariable("translationType", "sub").
			AddVariable("episodeString", episode)

		return qb.Execute(ctx, &videoResult)
	})

	if err != nil {
		return nil, errors.Wrap(err, errors.ScrapingError, "failed to get video source URLs")
	}

	return &videoResult, nil
}

func getVideoSourceURLs(showID, episode string) (*VideoResult, error) {
	return defaultClient.getVideoSourceURLs(context.Background(), showID, episode)
}
