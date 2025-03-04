package jira

import (
	"fmt"
	"jira-export/pkg/logger"
	"time"
)

// JiraSearchResults
type JiraSearchResults struct {
	RequestURL    string        `json:"requestURL"`
	Expand        string        `json:"expand"`
	StartAt       int           `json:"startAt"`
	MaxResults    int           `json:"maxResults"`
	Total         int           `json:"total"`
	Issues        []interface{} `json:"issues"`
	ErrorMessages *[]string     `json:"errorMessages,omitempty"`
}

func (j *JiraSearchResults) IssuesToJiraIssues() (issues Issues, err error) {

	// Loop through the issues and convert them to JiraIssue objects
	// and add them to the issues slice in parallel
	now := time.Now()
	for _, issue := range j.Issues {
		// Convert the issue to a JiraIssue object
		i, err := IssueFromInterface(issue)
		if err != nil {
			return issues, fmt.Errorf("error converting issue to JiraIssue: %v", err)
		}

		// Add the JiraIssue object to the issues slice
		issues = append(issues, i)
	}

	logger.Logger.Debug("Processing time for converting issues to JiraIssues", "processing_time", time.Since(now))

	return issues, nil
}

// WriteCSV writes the JiraSearchResults to a CSV file
func (j *JiraSearchResults) WriteCSV(filename string) error {
	// Convert the JiraSearchResults to a slice of JiraIssue objects
	issues, err := j.IssuesToJiraIssues()
	if err != nil {
		return fmt.Errorf("error converting issues: %v", err)
	}

	// Write the issues to a CSV file
	if err := issues.WriteCSV(filename); err != nil {
		return fmt.Errorf("error writing CSV: %v", err)
	}

	return nil
}
