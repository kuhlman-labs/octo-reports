package octoreports

import (
	"context"
	"encoding/csv"
	"log"
	"os"
	"time"

	"github.com/shurcooL/githubv4"
)

type Org struct {
	Login githubv4.String
	ID    githubv4.String
}

type Member struct {
	Id    string
	Name  string
	Login string
	Role  string
}

func getOrgMembersWithRole(orgName string, client *githubv4.Client) ([]*Member, error) {

	variables := map[string]interface{}{
		"orgName": githubv4.String(orgName),
		"cursor":  (*githubv4.String)(nil),
	}

	var query struct {
		Organization struct {
			MembersWithRole struct {
				PageInfo struct {
					EndCursor   githubv4.String
					HasNextPage bool
				}
				Edges []struct {
					Role string
					Node struct {
						Login string
					}
				}
			} `graphql:"membersWithRole(first: 100, after: $cursor)"`
		} `graphql:"organization(login: $orgName)"`
	}

	allMembers := []*Member{}
	start := time.Now()
	log.Printf("Fetching members for %s", orgName)
	for {
		err := client.Query(context.Background(), &query, variables)
		if err != nil {
			panic(err)
		}

		for _, edge := range query.Organization.MembersWithRole.Edges {
			allMembers = append(allMembers, &Member{
				Login: edge.Node.Login,
				Role:  edge.Role,
			})
		}

		if !query.Organization.MembersWithRole.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(query.Organization.MembersWithRole.PageInfo.EndCursor)
	}

	log.Printf("Found %d members in %s", len(allMembers), orgName)
	log.Printf("Fetched all members in %s", time.Since(start))

	return allMembers, nil
}

func GenerateOrgMembershipReport(enterpriseSlug string, client *githubv4.Client) error {
	file, err := os.Create("enterprise-orgs-member-report.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"Org Name", "Org ID", "Org Admins", "Org Members"}
	err = writer.Write(header)
	if err != nil {
		panic(err)
	}

	orgs, err := getEnterpriseOrgs(enterpriseSlug, client)
	if err != nil {
		panic(err)
	}

	for _, org := range orgs {
		orgMembers, err := getOrgMembersWithRole(string(org.Login), client)
		if err != nil {
			panic(err)
		}
		var admins, members string
		for _, member := range orgMembers {
			if string(member.Role) == "ADMIN" {
				admins += string(member.Login) + ", "
			} else {
				members += string(member.Login) + ", "
			}
		}
		record := []string{string(org.Login), string(org.ID), admins, members}
		err = writer.Write(record)
		if err != nil {
			panic(err)
		}
	}

	log.Printf("Wrote %d records to orgs.csv", len(orgs))

	return nil
}
