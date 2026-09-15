package api

import (
	"net/http"

	"github.com/shurcooL/graphql"
)

func NewGQLClient(url string, httpClient *http.Client) *graphql.Client {
	client := graphql.NewClient(url, httpClient)
	return client
}
