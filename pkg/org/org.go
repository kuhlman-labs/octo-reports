package org

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/kuhlman-labs/octo-reports/internal/client"
	"github.com/shurcooL/githubv4"
)

type Org struct {
	Login githubv4.String
	ID    githubv4.String
}

type Member struct {
	Login githubv4.String
	Role  githubv4.String
}

func getEnterpriseOrgs(enterpriseSlug, token string) []*Org {
	client := client.NewClient(token)

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

	return allOrgs
}

func getOrgMembersWithRole(orgName, token string) []*Member {
	client := client.NewClient(token)

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
					Role githubv4.String
					Node struct {
						Login githubv4.String
					}
				}
			} `graphql:"membersWithRole(first: 100, after: $cursor)"`
		} `graphql:"organization(login: $orgName)"`
	}

	allMembers := []*Member{}
	start := time.Now()
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

	return allMembers
}

func GenerateMembershipReport(enterpriseSlug, token string) {
	file, err := os.Create("orgs.csv")
	if err != nil {
		fmt.Println("Error creating the CSV file:", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write the header row
	header := []string{"Org Name", "Org ID", "Org Admins", "Org Members"}
	err = writer.Write(header)
	if err != nil {
		fmt.Println("Error writing the header row:", err)
		return
	}

	orgs := getEnterpriseOrgs("GitHub", token)

	for _, org := range orgs {
		orgMembers := getOrgMembersWithRole(string(org.Login), token)
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
			fmt.Println("Error writing a record to the CSV:", err)
			return
		}
	}

	fmt.Println("Successfully wrote to the CSV file")
}
