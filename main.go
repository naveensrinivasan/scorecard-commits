package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
)

type ScorecardData struct {
	Date time.Time `json:"date"`
	Repo struct {
		Name   string `json:"name"`
		Commit string `json:"commit"`
	} `json:"repo"`
	Scorecard struct {
		Version string `json:"version"`
		Commit  string `json:"commit"`
	} `json:"scorecard"`
	Score  float64 `json:"score"`
	Checks []struct {
		Name          string   `json:"name"`
		Score         int      `json:"score"`
		Reason        string   `json:"reason"`
		Details       []string `json:"details"`
		Documentation struct {
			Short string `json:"short"`
			Url   string `json:"url"`
		} `json:"documentation"`
	} `json:"checks"`
}
type ScoreCardVersion struct {
	Version string
	Commit  string
}
type Commit struct {
	Commit string
	Date   time.Time
}
type RepoData struct {
	Name   string
	Commit string
	Date   time.Time
}

type ScorecardRow struct {
	Date time.Time
	Repo struct {
		Name   string
		Commit string
	}
	Scorecard struct {
		Version string
		Commit  string
	}
	Score  float64
	Checks []struct {
		Name          string   `json:"name"`
		Score         int      `json:"score"`
		Reason        string   `json:"reason"`
		Details       []string `json:"details"`
		Documentation struct {
			Short string `json:"short"`
			Url   string `json:"url"`
		} `json:"documentation"`
	}
}

func main() {
	ctx := context.Background()
	days := 20
	// Get the repository location, project, and repo name from the command line
	if len(os.Args) < 4 {
		fmt.Println("Usage: <executable> <repoDir> <project> <repo>")
		os.Exit(1)
	}
	repoDir := os.Args[1]
	project := os.Args[2]
	repo := os.Args[3]
	var err error
	if len(os.Args) > 4 {
		days, err = strconv.Atoi(os.Args[4])
		if err != nil {
			fmt.Println("Error parsing days:", err)
			return
		}
	}

	// Get the commit from the given date
	date := time.Now().AddDate(0, 0, -days) // 7 days ago
	commits, err := getCommitsFromDate(repoDir, date)
	if err != nil {
		fmt.Println("Error getting commits:", err)
		return
	}

	// Process each commit
	for _, commit := range commits {
		// Get the scorecard data for the commit
		scorecardData, err := getScorecardData(project, repo, commit.Commit)
		if err != nil {
			fmt.Println("Error getting scorecard data:", err)
			continue
		}

		// Prepare the data for the BigQuery table
		row := ScorecardRow{
			Date: commit.Date,
			Repo: struct {
				Name   string
				Commit string
			}{
				Name:   fmt.Sprintf("github.com/%s/%s", project, repo),
				Commit: commit.Commit,
			},
			Scorecard: struct {
				Version string
				Commit  string
			}{
				Version: scorecardData.Scorecard.Version,
				Commit:  scorecardData.Scorecard.Commit,
			},
			Score:  scorecardData.Score,
			Checks: scorecardData.Checks,
		}

		// Save the data to the BigQuery table
		err = saveToBigQuery(ctx, row)
		if err != nil {
			fmt.Println("Error saving to BigQuery:", err)
			continue
		}

		fmt.Printf("Data for commit %s saved to BigQuery successfully!\n", commit.Commit)
	}

	fmt.Println("All data saved to BigQuery successfully!")
}

func getCommitsFromDate(repoDir string, date time.Time) ([]RepoData, error) {
	// Use the Git CLI to get the commits from the given date to the latest
	cmd := exec.Command("git", "log", "--since", date.Format("2006-01-02"), "--format=%H,%ct", "-n", "100")
	cmd.Dir = repoDir
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var commits []RepoData
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		parts := strings.Split(line, ",")
		if len(parts) != 2 {
			continue
		}
		commit := parts[0]
		commitTime, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			continue
		}
		commits = append(commits, RepoData{
			Name:   repoDir,
			Commit: commit,
			Date:   time.Unix(commitTime, 0),
		})
	}

	return commits, nil
}

func getScorecardData(project, repo, commitSHA string) (ScorecardData, error) {
	url := fmt.Sprintf("https://api.securityscorecards.dev/projects/github.com/%s/%s?commit=%s", project, repo, commitSHA)
	resp, err := http.Get(url)
	if err != nil {
		return ScorecardData{}, fmt.Errorf("failed to get scorecard data: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ScorecardData{}, fmt.Errorf("failed to read response body: %w", err)
	}

	var scorecardData ScorecardData
	err = json.Unmarshal(body, &scorecardData)
	if err != nil {
		return ScorecardData{}, fmt.Errorf("failed to unmarshal scorecard data: %w", err)
	}

	return scorecardData, nil
}

func saveToBigQuery(ctx context.Context, row ScorecardRow) error {
	// Create a BigQuery client
	client, err := bigquery.NewClient(ctx, "openssf")
	if err != nil {
		return err
	}

	// Create a new table reference
	table := client.Dataset("phren").Table("scorecard")

	// Prepare the data for insertion
	u := table.Inserter()
	err = u.Put(ctx, row)
	if err != nil {
		return fmt.Errorf("failed to insert data into BigQuery: %w", err)
	}

	return nil
}
