package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	octoreports "github.com/kuhlman-labs/octo-reports/pkg/octo-reports"
	"gopkg.in/yaml.v2"
)

// Config is a struct that holds the token and URL values from the config file
type Config struct {
	Token string `yaml:"token"`
	URL   string `yaml:"url"`
}

// loadConfig reads the config file and returns a Config struct
func loadConfig() Config {
	file, err := os.Open("config.yaml")
	if err != nil {
		log.Fatalf("Error opening config file: %v", err)
	}
	defer file.Close()

	var config Config
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatalf("Error decoding config file: %v", err)
	}

	return config
}

// parseRequiredFlags takes a flag set and a slice of flag names and checks if they are set or not
func parseRequiredFlags(fs *flag.FlagSet, flags []string) {
	fs.Parse(os.Args[2:])
	for _, name := range flags {
		if fs.Lookup(name).Value.String() == "" {
			fs.PrintDefaults()
			log.Fatalf("%s is required", name)
		}
	}
}

func main() {

	// Subcommands
	enterpriseCommand := flag.NewFlagSet("enterprise-report", flag.ExitOnError)
	orgCommand := flag.NewFlagSet("org-report", flag.ExitOnError)
	teamCommand := flag.NewFlagSet("team-report", flag.ExitOnError)
	repoCommand := flag.NewFlagSet("repo-report", flag.ExitOnError)
	collaboratorCommand := flag.NewFlagSet("collaborator-report", flag.ExitOnError)
	packageCommand := flag.NewFlagSet("package-report", flag.ExitOnError)
	//loginCommand := flag.NewFlagSet("login", flag.ExitOnError)

	// Enterprise flags
	enterpriseSlugPointer := enterpriseCommand.String("enterprise-slug", "", "(Required) The slug of the enterprise to run the report for.")

	// Org flags
	orgEnterpriseSlugPointer := orgCommand.String("enterprise-slug", "", "(Required) The slug of the enterprise to run the report for.")

	// Team flags
	teamEnterpriseSlugPointer := teamCommand.String("enterprise-slug", "", "(Required) The slug of the enterprise to run the report for.")

	// Repo flags
	repoEnterpriseSlugPointer := repoCommand.String("enterprise-slug", "", "(Required) The slug of the enterprise to run the report for.")

	// Collaborator flags
	collaboratorOrgPointer := collaboratorCommand.String("org", "", "(Required) The login of the organization to run the report for.")

	// Package flags
	packageOrgPointer := packageCommand.String("org", "", "(Required) The login of the organization to run the report for.")

	// Login flags
	//loginClientIdPointer := loginCommand.String("client-id", "", "(Required) The client ID of the GitHub App to use for authentication.")

	if len(os.Args) < 2 {
		log.Fatalf("Please specify a subcommand. Can be one of: enterprise-report, org-report, team-report, repo-report, collaborator-report")
	}

	// Load the config file
	config := loadConfig()

	switch os.Args[1] {
	case "enterprise-report":
		parseRequiredFlags(enterpriseCommand, []string{"enterprise-slug"})
		client := octoreports.NewV4Client(config.URL, config.Token)
		octoreports.GenerateEnterpriseMembershipReport(*enterpriseSlugPointer, client)
	case "org-report":
		parseRequiredFlags(orgCommand, []string{"enterprise-slug"})
		client := octoreports.NewV4Client(config.URL, config.Token)
		octoreports.GenerateOrgMembershipReport(*orgEnterpriseSlugPointer, client)
	case "team-report":
		parseRequiredFlags(teamCommand, []string{"enterprise-slug"})
		client := octoreports.NewV4Client(config.URL, config.Token)
		octoreports.GenerateTeamReport(*teamEnterpriseSlugPointer, client)
	case "repo-report":
		parseRequiredFlags(repoCommand, []string{"enterprise-slug"})
		client := octoreports.NewV4Client(config.URL, config.Token)
		octoreports.GenerateRepoReport(*repoEnterpriseSlugPointer, client)
	case "collaborator-report":
		parseRequiredFlags(collaboratorCommand, []string{"org"})
		client := octoreports.NewV4Client(config.URL, config.Token)
		octoreports.GenerateCollaboratorReport(*collaboratorOrgPointer, client)
	case "package-report":
		parseRequiredFlags(packageCommand, []string{"org"})
		client := octoreports.NewV4Client(config.URL, config.Token)
		octoreports.GenerateOrgPackageReport(*packageOrgPointer, client)
	// TODO: Implement login once Enterprise Apps are GA
	/*
		case "login":
			parseRequiredFlags(loginCommand, []string{"client-id"})
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
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}
}
