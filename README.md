# Octo-Reports
 Octo-Reports is a powerful command line tool written in Go that generates comprehensive reports on your GitHub Enterprise environment. This tool is designed to help you better manage your organization, repositories, and teams by providing you with detailed and organized information about your GitHub Enterprise setup.

## Features
Octo-Reports can generate the following reports:

* Organization Report: List out all admins and members for each organization in a GitHub Enterprise environment.
* Repository Report: List out all repositories contained in a GitHub Enterprise environment.
* Team Report: List out all teams in the Enterprise and their members.

## Installation
To install Octo-Reports, make sure you have the Go programming language installed on your system. You can download and install Go from the official website.

Once you have Go installed, run the following command to install Octo-Reports:

```bash
go get -u github.com/yourusername/octo-reports
```
This command will fetch the latest version of Octo-Reports and install it to your $GOPATH.

## Usage
Before you start using Octo-Reports, you need to set up a GitHub personal access token (PAT) with the appropriate permissions. You can create a PAT by following these instructions.

Once you have your PAT, you can use Octo-Reports to generate various reports as follows:

### Generate an Organization Report

```bash
octo-reports orgs -t <your_github_pat>
```

### Generate a Repository Report

```bash
octo-reports repos -t <your_github_pat>
```

### Generate a Team Report

```bash
octo-reports teams -t <your_github_pat>
```

### Optional Flags
-o, --output: Specify the output format for the report. Supported formats are json, csv, and table (default: table).
-f, --file: Specify the output file for the report. If not provided, the output will be displayed on the console.

## Contributing
We welcome and appreciate contributions to Octo-Reports. If you'd like to contribute, please fork the repository and submit a pull request with your changes.

## License
Octo-Reports is licensed under the MIT License. See the LICENSE file for more details.
