package main

import (
	"flag"
	"log"
	"os"

	octoreports "github.com/kuhlman-labs/octo-reports/pkg/octo-reports"
)

func main() {

	// Subcommands
	enterpriseCommand := flag.NewFlagSet("enterprise-report", flag.ExitOnError)
	orgCommand := flag.NewFlagSet("org-report", flag.ExitOnError)
	teamCommand := flag.NewFlagSet("team-report", flag.ExitOnError)
	repoCommand := flag.NewFlagSet("repo-report", flag.ExitOnError)
	collaboratorCommand := flag.NewFlagSet("collaborator-report", flag.ExitOnError)
	//loginCommand := flag.NewFlagSet("login", flag.ExitOnError)

	// Enterprise flags
	enterpriseSlugPointer := enterpriseCommand.String("enterprise-slug", "", "(Required) The slug of the enterprise to run the report for.")
	tokenPointer := enterpriseCommand.String("token", "", "(Required) The token to use to authenticate to the GitHub Enterprise instance.")
	enterpriseCommand.String("url", "https://api.github.com/graphql", "(Required) The URL of the GitHub Enterprise instance, if not set it will default to the public GitHub API.")

	// Org flags
	orgEnterpriseSlugPointer := orgCommand.String("enterprise-slug", "", "(Required) The slug of the enterprise to run the report for.")
	orgTokenPointer := orgCommand.String("token", "", "(Required) The token to use to authenticate to the GitHub Enterprise instance.")
	orgCommand.String("url", "https://api.github.com/graphql", "(Optional) The URL of the GitHub Enterprise instance, if not set it will default to the public GitHub API.")

	// Team flags
	teamEnterpriseSlugPointer := teamCommand.String("enterprise-slug", "", "(Required) The slug of the enterprise to run the report for.")
	teamTokenPointer := teamCommand.String("token", "", "(Required) The token to use to authenticate to the GitHub Enterprise instance.")
	teamCommand.String("url", "https://api.github.com/graphql", "(Optional) The URL of the GitHub Enterprise instance, if not set it will default to the public GitHub API.")

	// Repo flags
	repoEnterpriseSlugPointer := repoCommand.String("enterprise-slug", "", "(Required) The slug of the enterprise to run the report for.")
	repoTokenPointer := repoCommand.String("token", "", "(Required) The token to use to authenticate to the GitHub Enterprise instance.")
	repoCommand.String("url", "https://api.github.com/graphql", "(Optional) The URL of the GitHub Enterprise instance, if not set it will default to the public GitHub API.")

	// Collaborator flags
	collaboratorOrgPointer := collaboratorCommand.String("org", "", "(Required) The login of the organization to run the report for.")
	collaboratorTokenPointer := collaboratorCommand.String("token", "", "(Required) The token to use to authenticate to the GitHub Enterprise instance.")
	collaboratorCommand.String("url", "https://api.github.com/graphql", "(Optional) The URL of the GitHub Enterprise instance, if not set it will default to the public GitHub API.")

	// Login flags
	//loginClientIdPointer := loginCommand.String("client-id", "", "(Required) The client ID of the GitHub App to use for authentication.")

	if len(os.Args) < 2 {
		log.Fatalf("Please specify a subcommand. Can be one of: enterprise-report, org-report, team-report, repo-report, collaborator-report")
	}

	switch os.Args[1] {
	case "enterprise-report":
		enterpriseCommand.Parse(os.Args[2:])
	case "org-report":
		orgCommand.Parse(os.Args[2:])
	case "team-report":
		teamCommand.Parse(os.Args[2:])
	case "repo-report":
		repoCommand.Parse(os.Args[2:])
	case "collaborator-report":
		collaboratorCommand.Parse(os.Args[2:])
	//case "login":
	//	loginCommand.Parse(os.Args[2:])
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}

	if enterpriseCommand.Parsed() {
		if *enterpriseSlugPointer == "" {
			enterpriseCommand.PrintDefaults()
			log.Fatalf("enterprise-slug is required")
		}
		if *tokenPointer == "" {
			enterpriseCommand.PrintDefaults()
			log.Fatalf("token is required")
		}

		url := enterpriseCommand.Lookup("url").Value.String()
		token := enterpriseCommand.Lookup("token").Value.String()
		enterpriseName := enterpriseCommand.Lookup("enterprise-slug").Value.String()

		client := octoreports.NewV4Client(url, token)

		octoreports.GenerateEnterpriseMembershipReport(enterpriseName, client)
	}

	if orgCommand.Parsed() {
		if *orgEnterpriseSlugPointer == "" {
			orgCommand.PrintDefaults()
			log.Fatalf("enterprise-slug is required")
		}
		if *orgTokenPointer == "" {
			orgCommand.PrintDefaults()
			log.Fatalf("token is required")
		}

		url := orgCommand.Lookup("url").Value.String()
		token := orgCommand.Lookup("token").Value.String()
		enterpriseName := orgCommand.Lookup("enterprise-slug").Value.String()

		client := octoreports.NewV4Client(url, token)

		octoreports.GenerateOrgMembershipReport(enterpriseName, client)
	}

	if teamCommand.Parsed() {
		if *teamEnterpriseSlugPointer == "" {
			teamCommand.PrintDefaults()
			log.Fatalf("enterprise-slug is required")
		}
		if *teamTokenPointer == "" {
			teamCommand.PrintDefaults()
			log.Fatalf("token is required")
		}

		url := teamCommand.Lookup("url").Value.String()
		token := teamCommand.Lookup("token").Value.String()
		enterpriseName := teamCommand.Lookup("enterprise-slug").Value.String()

		client := octoreports.NewV4Client(url, token)

		octoreports.GenerateTeamReport(enterpriseName, client)
	}

	if repoCommand.Parsed() {
		if *repoEnterpriseSlugPointer == "" {
			repoCommand.PrintDefaults()
			log.Fatalf("enterprise-slug is required")
		}
		if *repoTokenPointer == "" {
			repoCommand.PrintDefaults()
			log.Fatalf("token is required")
		}

		url := repoCommand.Lookup("url").Value.String()
		token := repoCommand.Lookup("token").Value.String()

		enterpriseName := repoCommand.Lookup("enterprise-slug").Value.String()

		client := octoreports.NewV4Client(url, token)

		octoreports.GenerateRepoReport(enterpriseName, client)
	}

	if collaboratorCommand.Parsed() {
		if *collaboratorOrgPointer == "" {
			collaboratorCommand.PrintDefaults()
			log.Fatalf("org is required")
		}
		if *collaboratorTokenPointer == "" {
			collaboratorCommand.PrintDefaults()
			log.Fatalf("token is required")
		}

		url := collaboratorCommand.Lookup("url").Value.String()
		token := collaboratorCommand.Lookup("token").Value.String()
		org := collaboratorCommand.Lookup("org").Value.String()

		client := octoreports.NewV4Client(url, token)

		octoreports.GenerateCollaboratorReport(org, client)
	}
	// TODO: Implement login once Enterprise Apps are GA
	/*
		if loginCommand.Parsed() {
			if *loginClientIdPointer == "" {
				loginCommand.PrintDefaults()
				log.Fatalf("client-id is required")
			}

			clientId := loginCommand.Lookup("client-id").Value.String()

			// Oauth Login
			token, err := octoreports.RequestCode("https://github.com", clientId)
			if err != nil {
				panic(err)
			}
			//write token to file
			err = os.WriteFile("token.txt", []byte(token), 0644)
			if err != nil {
				panic(err)
			}

		}
	*/
}
