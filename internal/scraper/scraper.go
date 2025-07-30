package scraper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
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

func Search(query string) ([]Anime, error) {
	var animes []Anime
	gqlQuery := buildSearchQuery(query)
	jsonData, err := json.Marshal(gqlQuery)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/121.0")
	req.Header.Set("Referer", "https://allanime.to")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var searchResult SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
		return nil, err
	}

	for _, edge := range searchResult.Data.Shows.Edges {
		animes = append(animes, Anime{
			Title:    edge.Name,
			URL:      fmt.Sprintf("https://allanime.to/anime/%s", edge.ID),
			Episodes: fmt.Sprintf("%d", edge.AvailableEpisodes.Sub),
		})
	}

	return animes, nil
}
