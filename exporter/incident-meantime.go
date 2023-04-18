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

func FetchIncidentMTTR() ([]map[string]float64, error) {

	mttrReport := make([]map[string]float64, 0)

	mttrMap := make(map[string]float64)

	token := os.Getenv("GITHUB_OAUTH_TOKEN")
	// Authenticate to github using auth token
	client, err := authenticate.GitHubClient(token)
	if err != nil {
		log.Fatalln(err.Error())
	}
	opt := &github.RepositoryContentGetOptions{Ref: "main"}
	fileContent, _, _, err := client.Repositories.GetContents(context.Background(), owner, repo, escapePath, opt)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	if fileContent != nil {
		// TODO Read the fileContent from github.RepositoryContent
		content, err := fileContent.GetContent()
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		// Read the file line by line
		scanner := bufio.NewScanner(strings.NewReader(content))
		scanner.Split(bufio.ScanLines)
		var lines []string
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}

		// Loop through the lines
		for i, line := range lines {
			match, err := regexp.MatchString("^## Q[1-4]", line)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			if match {
				// parse the quarter value from the line
				quarter := line[3:10]
				// get the current quarter
				current_quarter := mogo.FormatQuarter(time.Now())
				if quarter == current_quarter {
					// goto the next line
					mttr := string(lines[i+2])
					// Parse the MTTRepair value from the line
					mttr = mttr[27:]
					// take the single space between duration
					mttr = strings.ReplaceAll(mttr, " ", "")
					mttr_time, err := time.ParseDuration(mttr)
					if err != nil {
						fmt.Println(err)
						return nil, err
					}
					mttr_hours := mttr_time.Hours()
					mttr_minutes := mttr_time.Minutes()
					total_minutes := mttr_hours*60 + mttr_minutes
					fmt.Println("mttrepair", total_minutes)

					mttrMap["incidents_mean_time_to_repair"] = total_minutes
					// goto the next line
					mttr = string(lines[i+4])
					// Parse the MTTResolve value from the line
					mttr = mttr[27:]
					// take the single space between duration
					mttr = strings.ReplaceAll(mttr, " ", "")
					mttr_time, err = time.ParseDuration(mttr)
					if err != nil {
						fmt.Println(err)
						return nil, err
					}
					mttr_hours = mttr_time.Hours()
					mttr_minutes = mttr_time.Minutes()
					total_minutes = mttr_hours*60 + mttr_minutes
					fmt.Println("mttresolve", total_minutes)
					mttrMap["incidents_mean_time_to_repair"] = total_minutes
					mttrReport = append(mttrReport, mttrMap)

				}
			}
		}
	}
	return mttrReport, nil
}
