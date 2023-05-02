package exporter

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-github/github"
	mogo "github.com/grokify/mogo/time/timeutil"
	"github.com/ministryofjustice/cloud-platform-environments/pkg/authenticate"
)

const (
	owner      = "ministryofjustice"
	repo       = "cloud-platform"
	escapePath = "runbooks/source/incident-log.html.md.erb"
)

func FetchIncidentMTTR() (mttrReport []map[string]float64, error error) {

	token := os.Getenv("GITHUB_OAUTH_TOKEN")
	// Authenticate to github using auth token
	client, err := authenticate.GitHubClient(token)
	if err != nil {
		log.Fatalln(err.Error())
	}
	opt := &github.RepositoryContentGetOptions{Ref: "main"}
	fileContent, _, _, err := client.Repositories.GetContents(context.Background(), owner, repo, escapePath, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to GetContents from the github repo: %w", err)
	}
	if fileContent != nil {
		// TODO Read the fileContent from github.RepositoryContent
		content, err := fileContent.GetContent()
		if err != nil {
			return nil, fmt.Errorf("failed to the raw GetContent from the github repo: %w", err)
		}
		// Read the file line by line
		lines := scanLines(content)
		mttrReport, err = parseMTTRQuarter(lines)
	}
	return mttrReport, nil
}

// scanLines reads a file into a slice of lines
func scanLines(content string) (lines []string) {
	// Read the file line by line
	scanner := bufio.NewScanner(strings.NewReader(content))
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		// Ignore empty lines
		if scanner.Text() == "" {
			continue
		}
		lines = append(lines, scanner.Text())
	}
	return lines
}

// parseMTTRQuarter parses the incident log file and returns the MTTR for the current quarter
func parseMTTRQuarter(lines []string) (mttrReport []map[string]float64, err error) {
	mttrMap := make(map[string]float64)
	// Loop through the lines and match the quarter
	for i, line := range lines {
		match, err := regexp.MatchString("^## Q[1-4]", line)
		if err != nil {
			return nil, fmt.Errorf("failed to the match reexp Q1-4: %w", err)
		}
		if match {
			// parse the quarter value from the line
			quarter := line[3:10]
			// get the current quarter
			current_quarter := mogo.FormatQuarter(time.Now())
			if quarter == current_quarter {
				// goto the next line
				mttr := string(lines[i+1])
				// Parse the MTTRepair value from the line
				mttr = mttr[27:]
				// take the single space between duration
				mttr = strings.ReplaceAll(mttr, " ", "")
				mttr_time, err := time.ParseDuration(mttr)
				if err != nil {
					return nil, fmt.Errorf("failed to the MTT Repair from incident log: %w", err)
				}
				mttr_minutes := mttr_time.Minutes()
				total_minutes := mttr_minutes

				mttrMap["incidents_mean_time_to_repair"] = total_minutes
				// goto the next line
				mttr = string(lines[i+2])
				// Parse the MTTResolve value from the line
				mttr = mttr[27:]
				// take the single space between duration
				mttr = strings.ReplaceAll(mttr, " ", "")
				mttr_time, err = time.ParseDuration(mttr)
				if err != nil {
					return nil, fmt.Errorf("failed to the MTT Resolve from incident log: %w", err)
				}
				mttr_minutes = mttr_time.Minutes()
				total_minutes = mttr_minutes
				mttrMap["incidents_mean_time_to_resolve"] = total_minutes
				mttrReport = append(mttrReport, mttrMap)

			}
		}
	}
	if mttrReport == nil {
		mttr_minutes := 0
		total_minutes := float64(mttr_minutes)
		mttrMap["incidents_mean_time_to_resolve"] = total_minutes
		mttrReport = append(mttrReport, mttrMap)
		return mttrReport, nil
	}
	return mttrReport, nil

}
