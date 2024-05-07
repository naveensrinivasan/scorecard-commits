# Scorecard Ingester

The Scorecard Ingester is a Go-based tool that fetches security scorecard data from the [Security Scorecard API](https://api.securityscorecards.dev/) and ingests it into a Google BigQuery table.

## Overview

The Scorecard Ingester is designed to automate the process of collecting and storing security scorecard data for a given GitHub repository. It performs the following steps:

1. Retrieves the commit history for the repository over the last X days.
2. For each commit, it fetches the corresponding security scorecard data from the Security Scorecard API.
3. Transforms the scorecard data into a format suitable for storage in a BigQuery table.
4. Saves the transformed data to the BigQuery table.

The tool is intended to be run periodically (e.g., daily or weekly) to keep the BigQuery table up-to-date with the latest security scorecard data for the repository.

## Why Use the Scorecard Ingester?

The Scorecard Ingester is useful for the following reasons:

- **Automation**: It can be scheduled to run periodically to keep the BigQuery table up-to-date.
- **Data Analysis**: It provides a structured way to analyze the security scorecard data for a given repository.
- **Data Storage**: It stores the security scorecard data in a BigQuery table, which is useful for data analysis and visualization.

### What can you do with the data?
- I want to upgrade my dependecy to not the latest version but I want to know the security score of the repository for that version.
- I want to know the security score of the repository for a given commit.
- I want to know how the security posture of the repository between release 1.0 and 2.0.
- I want to know the security score of the repository for a given release.
- How was the security scorecard for a given repository trending over time?
- What were the most common checks that were failing for a given repository?
- What were the most common reasons for failing checks for a given repository?

## Usage

To use the Scorecard Ingester, you'll need to have the following:

- A Google Cloud project with BigQuery enabled
- The path to a local Git repository you want to analyze
- The name of the GitHub repository you want to analyze
- The Google Cloud project ID

To run the tool, use the following command-line arguments:

```
go run main.go --repo-dir=/path/to/git/repo --project=your-gcp-project-id --repo=github-org/repo-name
```

Replace the following placeholders:

- `/path/to/git/repo`: The path to the local Git repository you want to analyze.
- `your-gcp-project-id`: The ID of your Google Cloud project.
- `github-org/repo-name`: The name of the GitHub repository you want to analyze.

The tool will then fetch the security scorecard data for the last 20 days of commits and save it to the BigQuery table named `phren.scorecard`.

## BigQuery Table Schema

The BigQuery table `phren.scorecard` has the following schema:

```
+---------------+---------------+------+-----+---------+-------+
| Field         | Type          | Null | Key | Default | Extra |
+---------------+---------------+------+-----+---------+-------+
| date          | TIMESTAMP     | NO   |     | NULL    |       |
| repo.name     | STRING        | NO   |     | NULL    |       |
| repo.commit   | STRING        | NO   |     | NULL    |       |
| scorecard.version | STRING    | NO   |     | NULL    |       |
| scorecard.commit | STRING     | NO   |     | NULL    |       |
| score         | FLOAT64       | NO   |     | NULL    |       |
| checks        | RECORD        | YES  |     | NULL    |       |
| checks.name   | STRING        | YES  |     | NULL    |       |
| checks.score  | INT64         | YES  |     | NULL    |       |
| checks.reason | STRING        | YES  |     | NULL    |       |
| checks.details| ARRAY<STRING> | YES  |     | NULL    |       |
| checks.documentation.short | STRING | YES |     | NULL  |       |
| checks.documentation.url  | STRING | YES |     | NULL  |       |
+---------------+---------------+------+-----+---------+-------+
```

This schema matches the structure of the `ScorecardRow` struct defined in the `main.go` file.

## Contributing

To contribute to the Scorecard Ingester project, please follow these steps:

1. Fork the repository.
2. Create a new branch for your changes.
3. Implement your changes in the appropriate packages.
4. Add tests for your changes.
5. Run the tests and ensure they pass.
6. Update the documentation if necessary.
7. Submit a pull request.

By following these guidelines, the codebase will remain clean, maintainable, and easy to contribute to.

## License

This project is licensed under the [MIT License](LICENSE).
