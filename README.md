# Octo-Reports
 Octo-Reports is a powerful command line tool written in Go that generates comprehensive reports on your GitHub Enterprise environment. This tool is designed to help you better manage your organization, repositories, and teams by providing you with detailed and organized information about your GitHub Enterprise setup.

## Features
Octo-Reports can generate the following reports:

* Enterprise Report: List out all members of a GitHub Enterprise environment.
* Organization Report: List out all admins and members for each organization in a GitHub Enterprise environment.
* Team Report: List out all teams in each organization in a GitHub Enterprise environment and their members.
* Repository Report: List out all repositories contained in a GitHub Enterprise environment and gathers information about each repository.
* Collaborator Report: List out all collaborators for each repository in a GitHub Enterprise organization.

## Installation
To install Octo-Reports, make sure you have the Go programming language installed on your system. You can download and install Go from the [official website](https://go.dev/doc/install).

Once you have Go installed, clone this repository and run the following command to install Octo-Reports:

```bash
go build octo-reports.go
```
This command will create an executable file called `octo-reports` in the current directory. You can move this file to a directory in your PATH to make it available system-wide.

## Usage
Before you start using Octo-Reports, you need to set up a GitHub personal access token (PAT) with the appropriate permissions. You can create a PAT by following [these](https://docs.github.com/en/enterprise-cloud@latest/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token) instructions.

Once you have your PAT, you can use Octo-Reports to generate various reports as follows:

### Generate an Enterprise Report

```bash
octo-reports enterprise-report -enterprise-slug <your_enterprise_slug> -token <your_github_pat>
```

Optionally you can specify the GitHub Enterprise URL with the -url flag.

```bash
octo-reports enterprise-report -enterprise-slug <your_enterprise_slug> -token <your_github_pat> -url <your_github_enterprise_url>
```

### Generate an Organization Report

```bash
octo-reports org-report -enterprise-slug <your_enterprise_slug> -token <your_github_pat>
```

### Generate a Repository Report

```bash
octo-reports repo-report -enterprise-slug <your_enterprise_slug> -token <your_github_pat>
```

### Generate a Team Report

```bash
octo-reports team-report -enterprise-slug <your_enterprise_slug> -token <your_github_pat>
```

### Generate a Collaborator Report

```bash
octo-reports collaborator-report -org <your_organization_id> -token <your_github_pat>
```

### Optional Flags
-url, --url: Specify the GitHub Enterprise URL. If not provided, the default GitHub API URL will be used. Example: `https://github.mycompany.com/api/graphql`

## Contributing
We welcome and appreciate contributions to Octo-Reports. If you'd like to contribute, please fork the repository and submit a pull request with your changes.

## License
Octo-Reports is licensed under the MIT License. See the [LICENSE](LICENSE) file for more details.
