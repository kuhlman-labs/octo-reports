package enterprise

import (
	"context"
	"log"
	"time"

	"github.com/kuhlman-labs/octo-reports/internal/client"
	"github.com/shurcooL/githubv4"
)

type Org struct {
	Login githubv4.String
	ID    githubv4.String
}

func GetEnterpriseOrgs(enterpriseSlug, token string) ([]*Org, error) {
	client := client.NewV4Client(token)

	variables := map[string]interface{}{
		"enterpriseSlug": githubv4.String(enterpriseSlug),
		"cursor":         (*githubv4.String)(nil),
	}

	var query struct {
		Enterprise struct {
			Organizations struct {
				PageInfo struct {
					EndCursor   githubv4.String
					HasNextPage bool
				}
				Nodes []struct {
					Login githubv4.String
					ID    githubv4.String
				}
			} `graphql:"organizations(first: 100, after: $cursor)"`
		} `graphql:"enterprise(slug: $enterpriseSlug)"`
	}

	allOrgs := []*Org{}
	start := time.Now()
	log.Printf("Fetching all orgs for %s", enterpriseSlug)
	for {
		err := client.Query(context.Background(), &query, variables)
		if err != nil {
			panic(err)
		}

		for _, org := range query.Enterprise.Organizations.Nodes {
			allOrgs = append(allOrgs, &Org{
				Login: org.Login,
				ID:    org.ID,
			})
		}

		if !query.Enterprise.Organizations.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(query.Enterprise.Organizations.PageInfo.EndCursor)
	}

	log.Printf("Found %d orgs in the %s Enterprise", len(allOrgs), enterpriseSlug)
	log.Printf("Fetched all orgs in %s", time.Since(start))

	return allOrgs, nil
}
