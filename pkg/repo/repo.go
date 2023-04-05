package repo

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

type Repo struct {
	Name       string
	Visibility string
	IsArchived bool
	IsFork     bool
	ID         string
	PushedAt   time.Time
	CreatedAt  time.Time
	Owner      string
	Topics     []string
	Teams      []Team
}

type Team struct {
	Name string
	Role string
}

type Topics struct {
	Nodes []struct {
		Topic struct {
			Name string
		}
	}
}

func getOrgRepos(orgName, token string) ([]*Repo, error) {
	client := client.NewV4Client(token)

	variables := map[string]interface{}{
		"orgName": githubv4.String(orgName),
		"cursor":  (*githubv4.String)(nil),
	}

	var query struct {
		Organization struct {
			Repositories struct {
				PageInfo struct {
					EndCursor   githubv4.String
					HasNextPage bool
				}
				Nodes []struct {
					Name             string
					Visibility       string
					IsArchived       bool
					IsFork           bool
					ID               string
					PushedAt         time.Time
					CreatedAt        time.Time
					RepositoryTopics struct {
						Nodes []struct {
							Topic struct {
								Name string
							}
						}
					} `graphql:"repositoryTopics(first: 100)"`
					Owner struct {
						Login string
					}
				}
			} `graphql:"repositories(first: 100, after: $cursor)"`
		} `graphql:"organization(login: $orgName)"`
	}

	allRepos := []*Repo{}
	start := time.Now()
	log.Printf("Fetching all repos for the %s organization.", orgName)
	for {
		err := client.Query(context.Background(), &query, variables)
		if err != nil {
			panic(err)
		}

		for _, repo := range query.Organization.Repositories.Nodes {

			teams, _ := getTeamsRoleForRepo(orgName, repo.Name, token)

			topics := []string{}
			for _, t := range repo.RepositoryTopics.Nodes {
				topics = append(topics, t.Topic.Name)
			}

			allRepos = append(allRepos, &Repo{
				Name:       repo.Name,
				Visibility: repo.Visibility,
				IsArchived: repo.IsArchived,
				IsFork:     repo.IsFork,
				ID:         repo.ID,
				PushedAt:   repo.PushedAt,
				CreatedAt:  repo.CreatedAt,
				Owner:      repo.Owner.Login,
				Topics:     topics,
				Teams:      teams,
			})
		}

		if !query.Organization.Repositories.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(query.Organization.Repositories.PageInfo.EndCursor)
	}

	log.Printf("Found %d repositories in %s", len(allRepos), orgName)
	log.Printf("Fetched all repos in %v", time.Since(start))

	return allRepos, nil
}

func getTeamsRoleForRepo(orgName, repoName, token string) ([]Team, error) {
	client := client.NewV4Client(token)

	variables := map[string]interface{}{
		"orgName":  githubv4.String(orgName),
		"repoName": githubv4.String(repoName),
		"cursor":   (*githubv4.String)(nil),
	}

	var query struct {
		Organization struct {
			Teams struct {
				PageInfo struct {
					EndCursor   githubv4.String
					HasNextPage bool
				}
				Nodes []struct {
					Repositories struct {
						Edges []struct {
							Permission string
							Node       struct {
								Name string
							}
						}
					} `graphql:"repositories(first: 1, query: $repoName)"`
				}
			} `graphql:"teams(first: 100, after: $cursor)"`
		} `graphql:"organization(login: $orgName)"`
	}

	allTeams := []Team{}
	start := time.Now()
	log.Printf("Fetching all teams for the %s repository.", repoName)
	for {
		err := client.Query(context.Background(), &query, variables)
		if err != nil {
			panic(err)
		}

		for _, team := range query.Organization.Teams.Nodes {

			if len(team.Repositories.Edges) > 0 {

				allTeams = append(allTeams, Team{
					Name: team.Repositories.Edges[0].Node.Name,
					Role: team.Repositories.Edges[0].Permission,
				})
			}
		}

		if !query.Organization.Teams.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(query.Organization.Teams.PageInfo.EndCursor)
	}

	log.Printf("Found %d teams in %s", len(allTeams), repoName)
	log.Printf("Fetched all teams in %v", time.Since(start))

	return allTeams, nil
}

func GenerateRepoReport(orgName, token string) error {
	file, err := os.Create("repos.csv")
	if err != nil {
		fmt.Println("Error creating the CSV file:", err)
	}

	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"id", "owner", "name", "visibility", "archived", "is_fork", "created_at", "pushed_at", "teams", "topics"}

	err = writer.Write(header)
	if err != nil {
		fmt.Println("Error writing the header row:", err)
		return nil
	}

	repos, err := getOrgRepos(orgName, token)
	if err != nil {
		log.Fatal(err)
	}

	for _, repo := range repos {
		var teams []string
		for _, team := range repo.Teams {
			teams = append(teams, team.Name+":"+team.Role)
		}

		record := []string{
			repo.ID,
			string(repo.Owner),
			string(repo.Name),
			string(repo.Visibility),
			fmt.Sprintf("%t", repo.IsArchived),
			fmt.Sprintf("%t", repo.IsFork),
			repo.CreatedAt.Format(time.RFC3339),
			repo.PushedAt.Format(time.RFC3339),
			fmt.Sprintf("%v", teams),
			fmt.Sprintf("%v", repo.Topics),
		}

		err := writer.Write(record)
		if err != nil {
			fmt.Println("Error writing the record:", err)
		}
	}

	log.Printf("Wrote %d records to repos.csv", len(repos))

	return nil
}
