package graphql

import (
	"context"
	"encoding/json"

	"github.com/keircn/karu/pkg/errors"
	"github.com/keircn/karu/pkg/http"
)

type Query struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

type QueryBuilder struct {
	baseQuery  string
	variables  map[string]any
	httpClient *http.Client
	endpoint   string
}

func NewQueryBuilder(endpoint string, httpClient *http.Client) *QueryBuilder {
	return &QueryBuilder{
		variables:  make(map[string]any),
		httpClient: httpClient,
		endpoint:   endpoint,
	}
}

func (qb *QueryBuilder) SetQuery(query string) *QueryBuilder {
	qb.baseQuery = query
	return qb
}

func (qb *QueryBuilder) AddVariable(key string, value any) *QueryBuilder {
	qb.variables[key] = value
	return qb
}

func (qb *QueryBuilder) AddSearchInput(query string, allowAdult, allowUnknown bool) *QueryBuilder {
	searchInput := map[string]any{
		"allowAdult":   allowAdult,
		"allowUnknown": allowUnknown,
		"query":        query,
	}
	return qb.AddVariable("search", searchInput)
}

func (qb *QueryBuilder) AddPagination(limit, page int) *QueryBuilder {
	return qb.AddVariable("limit", limit).AddVariable("page", page)
}

func (qb *QueryBuilder) AddTranslationType(translationType string) *QueryBuilder {
	return qb.AddVariable("translationType", translationType)
}

func (qb *QueryBuilder) AddCountryOrigin(countryOrigin string) *QueryBuilder {
	return qb.AddVariable("countryOrigin", countryOrigin)
}

func (qb *QueryBuilder) Build() Query {
	return Query{
		Query:     qb.baseQuery,
		Variables: qb.variables,
	}
}

func (qb *QueryBuilder) Execute(ctx context.Context, result interface{}) error {
	query := qb.Build()

	resp, err := qb.httpClient.PostJSON(ctx, qb.endpoint, query)
	if err != nil {
		return errors.Wrap(err, errors.NetworkError, "failed to execute GraphQL query")
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return errors.Wrap(err, errors.NetworkError, "failed to decode GraphQL response")
	}

	return nil
}

const ShowsQuery = `query($search: SearchInput, $limit: Int, $page: Int, $translationType: VaildTranslationTypeEnumType, $countryOrigin: VaildCountryOriginEnumType) {
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
}`
