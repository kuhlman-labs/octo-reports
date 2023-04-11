package octoreports

import (
	"context"
	"encoding/csv"
	"log"
	"os"
	"time"

	"github.com/shurcooL/githubv4"
)

func getEnterpriseOrgs(enterpriseSlug string, client *githubv4.Client) ([]*Org, error) {

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

func getEnterpriseMembers(enterpriseSlug string, client *githubv4.Client) ([]*Member, error) {

	variables := map[string]interface{}{
		"enterpriseSlug": githubv4.String(enterpriseSlug),
		"cursor":         (*githubv4.String)(nil),
	}

	var query struct {
		Enterprise struct {
			Members struct {
				PageInfo struct {
					EndCursor   githubv4.String
					HasNextPage bool
				}
				Nodes []struct {
					EnterpriseUserAccount struct {
						Id    string
						Login string
						Name  string
					} `graphql:"... on EnterpriseUserAccount"`
					User struct {
						Id    string
						Login string
						Name  string
					} `graphql:"... on User"`
				}
			} `graphql:"members(first: 100, after: $cursor)"`
		} `graphql:"enterprise(slug: $enterpriseSlug)"`
	}

	allMembers := []*Member{}
	start := time.Now()
	log.Printf("Fetching all members for the %s Enterprise.", enterpriseSlug)
	for {
		err := client.Query(context.Background(), &query, variables)
		if err != nil {
			panic(err)
		}

		for _, member := range query.Enterprise.Members.Nodes {
			allMembers = append(allMembers, &Member{
				Login: member.EnterpriseUserAccount.Login,
				Name:  member.EnterpriseUserAccount.Name,
				Id:    member.EnterpriseUserAccount.Id,
			})
		}

		if !query.Enterprise.Members.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(query.Enterprise.Members.PageInfo.EndCursor)
	}

	log.Printf("Found %d members in the %s Enterprise", len(allMembers), enterpriseSlug)
	log.Printf("Fetched all members in %s", time.Since(start))

	return allMembers, nil
}

func GenerateEnterpriseMembershipReport(enterpriseSlug string, client *githubv4.Client) error {

	file, err := os.Create("enterprise-membership-report.csv")
	if err != nil {
		panic(err)
	}

	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"Login", "Name", "Id"})

	allMembers, err := getEnterpriseMembers(enterpriseSlug, client)
	if err != nil {
		panic(err)
	}

	for _, member := range allMembers {
		writer.Write([]string{string(member.Login), string(member.Name), string(member.Id)})
	}

	log.Printf("Wrote %d records to enterprise-membership-report.csv", len(allMembers))

	return nil
}
