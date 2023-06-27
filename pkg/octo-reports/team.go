package octoreports

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/shurcooL/githubv4"
)

type Team struct {
	ID          string
	Name        string
	Slug        string
	Description string
	Role        string
	Members     []*Member
}

func getOrgTeams(orgName string, client *githubv4.Client) ([]Team, error) {

	variables := map[string]interface{}{
		"orgName": githubv4.String(orgName),
		"cursor":  (*githubv4.String)(nil),
	}

	var query struct {
		Organization struct {
			Teams struct {
				PageInfo struct {
					EndCursor   githubv4.String
					HasNextPage bool
				}
				Nodes []struct {
					ID          string
					Name        string
					Slug        string
					Description string
				}
			} `graphql:"teams(first: 100, after: $cursor)"`
		} `graphql:"organization(login : $orgName)"`
		RateLimit RateLimit
	}

	allTeams := []Team{}
	startTime := time.Now()
	log.Printf("Fetching all teams for %s", orgName)
	for {
		err := client.Query(context.Background(), &query, variables)
		if err != nil {
			panic(err)
		}

		// check rate limit
		if query.RateLimit.Remaining < 100 {
			log.Printf("Rate limit remaining: %d", query.RateLimit.Remaining)
			time.Sleep(time.Minute)
		}

		for _, team := range query.Organization.Teams.Nodes {

			allMembers, err := getTeamMembers(orgName, team.Slug, client)
			if err != nil {
				panic(err)
			}

			allTeams = append(allTeams, Team{
				ID:          team.ID,
				Name:        team.Name,
				Slug:        team.Slug,
				Description: team.Description,
				Members:     allMembers,
			})
		}

		if !query.Organization.Teams.PageInfo.HasNextPage {
			break
		}

		variables["cursor"] = githubv4.NewString(query.Organization.Teams.PageInfo.EndCursor)

	}
	log.Printf("Found %d teams in the %s organization.", len(allTeams), orgName)
	log.Printf("Fetched all teams in  %s.", time.Since(startTime))

	return allTeams, nil
}

func getTeamMembers(orgName, teamSlug string, client *githubv4.Client) ([]*Member, error) {

	variables := map[string]interface{}{
		"orgName":  githubv4.String(orgName),
		"teamSlug": githubv4.String(teamSlug),
		"cursor":   (*githubv4.String)(nil),
	}

	var query struct {
		Organization struct {
			Teams struct {
				Nodes []struct {
					Members struct {
						PageInfo struct {
							EndCursor   githubv4.String
							HasNextPage bool
						}
						Nodes []struct {
							Login string
						}
					} `graphql:"members(first: 100, after: $cursor)"`
				}
			} `graphql:"teams(first: 1, query: $teamSlug)"`
		} `graphql:"organization(login: $orgName)"`
		RateLimit RateLimit
	}

	allMembers := []*Member{}
	startTime := time.Now()
	log.Printf("Fetching all members for %s/%s", orgName, teamSlug)
	for {
		err := client.Query(context.Background(), &query, variables)
		if err != nil {
			panic(err)
		}

		// check rate limit
		if query.RateLimit.Remaining < 100 {
			log.Printf("Rate limit remaining: %d", query.RateLimit.Remaining)
			time.Sleep(time.Minute)
		}

		for _, member := range query.Organization.Teams.Nodes[0].Members.Nodes {
			allMembers = append(allMembers, &Member{
				Login: member.Login,
			})
		}

		if !query.Organization.Teams.Nodes[0].Members.PageInfo.HasNextPage {
			break
		}

		variables["cursor"] = githubv4.NewString(query.Organization.Teams.Nodes[0].Members.PageInfo.EndCursor)

	}

	log.Printf("Found %d members in the %s/%s team.", len(allMembers), orgName, teamSlug)
	log.Printf("Fetched all members in  %s.", time.Since(startTime))

	return allMembers, nil

}

func GenerateTeamReport(enterpriseSlug string, client *githubv4.Client) error {
	file, err := os.Create("teams.csv")
	if err != nil {
		log.Println("Error creating the CSV file:", err)
	}

	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"id", "organization", "name", "slug", "description", "members"}

	err = writer.Write(header)
	if err != nil {
		log.Println("Error writing the header row:", err)
		return nil
	}

	orgs, err := getEnterpriseOrgs(enterpriseSlug, client)
	if err != nil {
		log.Fatal(err)
	}

	for _, org := range orgs {
		teams, err := getOrgTeams(string(org.Login), client)
		if err != nil {
			log.Fatal(err)
		}

		for _, team := range teams {

			members := []string{}

			for _, member := range team.Members {
				members = append(members, member.Login)
			}

			record := []string{
				team.ID,
				string(org.Login),
				team.Name,
				team.Slug,
				team.Description,
				fmt.Sprintf("%v", members),
			}
			err := writer.Write(record)
			if err != nil {
				log.Println("Error writing the record:", err)
			}

		}
	}

	return nil
}
