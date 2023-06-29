package octoreports

import (
	"context"
	"net/http"

	"github.com/google/go-github/v50/github"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type RateLimit struct {
	Cost      githubv4.Int
	Limit     githubv4.Int
	NodeCount githubv4.Int
	Remaining githubv4.Int
	ResetAt   githubv4.DateTime
	Used      githubv4.Int
}

func NewV4Client(url, token string) *githubv4.Client {

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}

	ctx := context.Background()

	httpClient := &http.Client{
		Transport: transport,
	}

	ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)

	httpClient = oauth2.NewClient(ctx, src)

	if url != "https://api.github.com/graphql" {
		client := githubv4.NewEnterpriseClient(url, httpClient)
		return client
	}

	client := githubv4.NewClient(httpClient)

	return client
}

func NewV3Client(token string) *github.Client {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)

	client := github.NewClient(httpClient)
	return client
}
