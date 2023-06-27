package octoreports

import (
	"context"
	"encoding/csv"
	"log"
	"os"
	"time"

	"github.com/shurcooL/githubv4"
)

type Package struct {
	Name       githubv4.String
	ID         githubv4.String
	Repository struct {
		Name githubv4.String
	}
}

func getPackages(orgName string, client *githubv4.Client) ([]*Package, error) {

	variables := map[string]interface{}{
		"orgName": githubv4.String(orgName),
		"cursor":  (*githubv4.String)(nil),
	}

	var query struct {
		Organization struct {
			Packages struct {
				PageInfo struct {
					EndCursor   githubv4.String
					HasNextPage bool
				}
				Nodes []Package
			} `graphql:"packages(first: 100, after: $cursor)"`
		} `graphql:"organization(login: $orgName)"`
		RateLimit RateLimit
	}

	allPackages := []*Package{}
	start := time.Now()
	log.Printf("Fetching packages for the %s organization", orgName)
	for {
		err := client.Query(context.Background(), &query, variables)
		if err != nil {
			panic(err)
		}

		// check rate limit
		if query.RateLimit.Remaining < 100 {
			log.Printf("Rate limit: %d/%d, resets at: %s", query.RateLimit.Remaining, query.RateLimit.Limit, query.RateLimit.ResetAt)
			time.Sleep(time.Until(query.RateLimit.ResetAt.Time))
		}

		for _, node := range query.Organization.Packages.Nodes {
			allPackages = append(allPackages, &Package{
				Name:       node.Name,
				Repository: node.Repository,
			})
		}

		if !query.Organization.Packages.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(query.Organization.Packages.PageInfo.EndCursor)
	}
	log.Printf("Fetched %d packages in %v", len(allPackages), time.Since(start))

	return allPackages, nil
}

func GenerateOrgPackageReport(orgName string, client *githubv4.Client) {
	packages, err := getPackages(orgName, client)
	if err != nil {
		panic(err)
	}

	file, err := os.Create("packages.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"Package Name", "Repository Name"})

	for _, pkg := range packages {
		writer.Write([]string{string(pkg.Name), string(pkg.Repository.Name)})
	}
}
