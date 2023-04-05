package org

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/kuhlman-labs/octo-reports/internal/client"
	"github.com/kuhlman-labs/octo-reports/pkg/enterprise"
	"github.com/shurcooL/githubv4"
)

type Member struct {
	Login githubv4.String
	Role  githubv4.String
}

func getOrgMembersWithRole(orgName, token string) ([]*Member, error) {
	client := client.NewV4Client(token)

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

func GenerateMembershipReport(enterpriseSlug, token string) {
	file, err := os.Create("orgs.csv")
	if err != nil {
		fmt.Println("Error creating the CSV file: %w", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"Org Name", "Org ID", "Org Admins", "Org Members"}
	err = writer.Write(header)
	if err != nil {
		fmt.Println("Error writing the header row:", err)
		return
	}

	orgs, error := enterprise.GetEnterpriseOrgs(enterpriseSlug, token)
	if error != nil {
		fmt.Println("Error getting orgs:", error)
		return
	}

	for _, org := range orgs {
		orgMembers, error := getOrgMembersWithRole(string(org.Login), token)
		if error != nil {
			fmt.Println("Error getting org members:", error)
			return
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
			fmt.Println("Error writing a record to the CSV:", err)
			return
		}
	}

	log.Printf("Wrote %d records to orgs.csv", len(orgs))
}
