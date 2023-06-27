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

type Collaborator struct {
	Permission string
	Login      string
	Name       string
	Email      string
	DatabaseID uint64
}

type Topics struct {
	Nodes []struct {
		Topic struct {
			Name string
		}
	}
}

func getOrgRepos(orgName string, getTeams bool, client *githubv4.Client) ([]*Repo, error) {

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
		RateLimit RateLimit
	}

	allRepos := []*Repo{}
	start := time.Now()
	log.Printf("Fetching all repos for the %s organization.", orgName)
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

		for _, repo := range query.Organization.Repositories.Nodes {

			if getTeams {

				teams, _ := getTeamsRoleForRepo(orgName, repo.Name, *client)

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
			} else {
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
				})
			}
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

func getTeamsRoleForRepo(orgName, repoName string, client githubv4.Client) ([]Team, error) {

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
					Slug         string
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
		RateLimit RateLimit
	}

	allTeams := []Team{}
	start := time.Now()
	log.Printf("Fetching all teams for the %s repository.", repoName)
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

		for _, team := range query.Organization.Teams.Nodes {

			if len(team.Repositories.Edges) > 0 {

				allTeams = append(allTeams, Team{
					Name: team.Slug,
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

func GenerateRepoReport(enterpriseSlug string, client *githubv4.Client) error {
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

	orgs, _ := getEnterpriseOrgs(enterpriseSlug, client)
	for _, org := range orgs {

		repos, err := getOrgRepos(string(org.Login), true, client)
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
	}

	return nil
}

func getRepoCollaborators(orgName, repoName string, client *githubv4.Client) ([]*Collaborator, error) {

	variables := map[string]interface{}{
		"orgName":  githubv4.String(orgName),
		"repoName": githubv4.String(repoName),
		"cursor":   (*githubv4.String)(nil),
	}

	var query struct {
		Organization struct {
			Repository struct {
				Collaborators struct {
					PageInfo struct {
						EndCursor   githubv4.String
						HasNextPage bool
					}
					Edges []struct {
						Permission string
						Node       struct {
							Login      string
							Name       string
							Email      string
							DatabaseID uint64
						}
					}
				} `graphql:"collaborators(affiliation: ALL, first: 100, after: $cursor)"`
			} `graphql:"repository(name: $repoName)"`
		} `graphql:"organization(login: $orgName)"`
		RateLimit RateLimit
	}

	allCollaborators := []*Collaborator{}

	start := time.Now()
	log.Printf("Fetching all collaborators for the %s repository.", repoName)
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

		for _, collaborator := range query.Organization.Repository.Collaborators.Edges {
			allCollaborators = append(allCollaborators, &Collaborator{
				Login:      collaborator.Node.Login,
				Name:       collaborator.Node.Name,
				Email:      collaborator.Node.Email,
				DatabaseID: collaborator.Node.DatabaseID,
				Permission: collaborator.Permission,
			})
		}

		if !query.Organization.Repository.Collaborators.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(query.Organization.Repository.Collaborators.PageInfo.EndCursor)
	}

	log.Printf("Found %d collaborators in %s", len(allCollaborators), repoName)
	log.Printf("Fetched all collaborators in %v", time.Since(start))

	return allCollaborators, nil
}

func GenerateCollaboratorReport(orgName string, client *githubv4.Client) error {
	file, err := os.Create("collaborators.csv")
	if err != nil {
		fmt.Println("Error creating the CSV file:", err)
	}

	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"repo_id", "org", "repo", "is_archived", "Collaborators"}

	err = writer.Write(header)
	if err != nil {
		fmt.Println("Error writing the header row:", err)
		return nil
	}

	repos, err := getOrgRepos(orgName, false, client)
	if err != nil {
		log.Fatal(err)
	}

	for _, repo := range repos {
		collaborators, err := getRepoCollaborators(orgName, string(repo.Name), client)
		if err != nil {
			log.Fatal(err)
		}

		collaboratorList := []string{}
		for _, collaborator := range collaborators {
			collaboratorList = append(collaboratorList, fmt.Sprintf("%d:%s:%s:%s:%s", collaborator.DatabaseID, collaborator.Name, collaborator.Email, collaborator.Login, collaborator.Permission))
		}

		record := []string{
			repo.ID,
			orgName,
			string(repo.Name),
			fmt.Sprintf("%t", repo.IsArchived),
			fmt.Sprintf("%v", collaboratorList),
		}

		err = writer.Write(record)
		if err != nil {
			fmt.Println("Error writing the record:", err)
		}
	}

	log.Printf("Wrote %d records to collaborators.csv", len(repos))

	return nil
}
