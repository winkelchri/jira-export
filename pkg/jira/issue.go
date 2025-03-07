package jira

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
)

// Issues
type Issues []Issue

// WriteCSV writes the Issues to a CSV file
func (i *Issues) WriteCSV(filename string) error {
	// Open the output CSV file for writing.
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Create a new CSV writer.
	writer := csv.NewWriter(file)

	// Write the header
	header := []string{
		"key",
		"reporter",
		"assignee",
		"creator",
		"title",
		"components",
		"status",
		"issuetype",
		"resolutiondate",
		"updated",
		"created",
		"statusCategoryChangeDate",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("error writing header: %v", err)
	}

	// Write the rows
	for _, issue := range *i {
		row := []string{
			issue.Key,
			issue.Reporter.DisplayName,
			issue.Assignee.DisplayName,
			issue.Creator.DisplayName,
			issue.Title,
			strings.Join(issue.Components, "|"),
			issue.Status,
			issue.IssueType,
			issue.ResolutionDate,
			issue.Updated,
			issue.Created,
			issue.StatusCategoryChangeDate,
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("error writing row: %v", err)
		}
	}

	// Flush the writer
	writer.Flush()

	return nil
}

// Issue
type Issue struct {
	Assignee                 JiraIssueUser `json:"assignee"`
	Components               []string      `json:"components"`
	Created                  string        `json:"created"`
	Creator                  JiraIssueUser `json:"creator"`
	Description              string        `json:"description"`
	ID                       string        `json:"id"`
	IssueType                string        `json:"issuetype"`
	Key                      string        `json:"key"`
	Reporter                 JiraIssueUser `json:"reporter"`
	ResolutionDate           string        `json:"resolutiondate"`
	Self                     string        `json:"self"`
	Summary                  string        `json:"summary"`
	Status                   string        `json:"status"`
	StatusCategoryChangeDate string        `json:"statuscategorychangedate"`
	Title                    string        `json:"title"`
	Updated                  string        `json:"updated"`
}

func IssueFromInterface(i any) (issue Issue, err error) {

	// Convert the interface to a map
	issueMap, ok := i.(map[string]any)
	if !ok {
		return issue, fmt.Errorf("error converting raw issue to map")
	}

	// Extract the "fields" object from the issueMap
	fieldsMap, ok := issueMap["fields"].(map[string]any)
	if !ok {
		return issue, fmt.Errorf("error converting fields to map")
	}

	// Set the assignee
	if assignee, ok := fieldsMap["assignee"].(map[string]any); ok {
		issue.Assignee.FromInterface(assignee)
	}

	// Set the Title field
	if title, ok := fieldsMap["summary"].(string); ok {
		issue.Title = title
	}

	// Set the Description field
	if description, ok := fieldsMap["description"].(map[string]any); ok {
		issue.Description = extractDescription(description)
	}

	// Set the components
	if components, ok := fieldsMap["components"].([]any); ok {
		for _, component := range components {
			if c, ok := component.(map[string]any); ok {
				if name, ok := c["name"].(string); ok {
					issue.Components = append(issue.Components, name)
				}
			}
		}
	}

	// Set the created date
	if created, ok := fieldsMap["created"].(string); ok {
		issue.Created = created
	}

	// Set the Creator field
	if creator, ok := fieldsMap["creator"].(map[string]any); ok {
		issue.Creator.FromInterface(creator)
	}

	// Set the ID field
	if id, ok := issueMap["id"].(string); ok {
		issue.ID = id
	}

	// Set the IssueType field
	if issueType, ok := fieldsMap["issuetype"].(map[string]any); ok {
		if name, ok := issueType["name"].(string); ok {
			issue.IssueType = name
		}
	}

	// Set the Key field
	if key, ok := issueMap["key"].(string); ok {
		issue.Key = key
	}

	// Set the Reporter field
	if reporter, ok := fieldsMap["reporter"].(map[string]any); ok {
		issue.Reporter.FromInterface(reporter)
	}

	// Set the ResolutionDate field
	if resolutionDate, ok := fieldsMap["resolutiondate"].(string); ok {
		issue.ResolutionDate = resolutionDate
	}

	// Set the Self field
	if self, ok := issueMap["self"].(string); ok {
		issue.Self = self
	}

	// Set the Summary field
	if summary, ok := fieldsMap["summary"].(string); ok {
		issue.Summary = summary
	}

	// Set the Status field
	if status, ok := fieldsMap["status"].(map[string]any); ok {
		if name, ok := status["name"].(string); ok {
			issue.Status = name
		}
	}

	// Set the StatusCategoryChangeDate field
	if statusCategoryChangeDate, ok := fieldsMap["statuscategorychangedate"].(string); ok {
		issue.StatusCategoryChangeDate = statusCategoryChangeDate
	}

	// Set the Updated field
	if updated, ok := fieldsMap["updated"].(string); ok {
		issue.Updated = updated
	}

	return issue, nil
}

func extractDescription(i map[string]any) string {
	var out string

	// If the type is a doc, then extract the content
	if i["type"] == "doc" {
		if content, ok := i["content"].([]any); ok {
			for _, c := range content {
				if m, ok := c.(map[string]any); ok {
					out = out + extractDescription(m)
				}
			}
		}
	}

	// If the type is a paragraph, then extract the content
	if i["type"] == "paragraph" {
		if content, ok := i["content"].([]any); ok {
			for _, c := range content {
				if m, ok := c.(map[string]any); ok {
					out = out + extractDescription(m)
				}
			}
		}
	}

	// If the type is a table, then extract the content from the cells
	if i["type"] == "table" {
		if content, ok := i["content"].([]any); ok {
			for _, c := range content {
				if row, ok := c.(map[string]any); ok {
					out = out + extractDescription(row)
				}
			}
		}
	}

	// If the type is a tableRow, then extract the content from the cells
	if i["type"] == "tableRow" {
		if content, ok := i["content"].([]any); ok {
			for _, c := range content {
				if cell, ok := c.(map[string]any); ok {
					out = out + extractDescription(cell)
				}
			}
		}
	}

	// If the type is a tableCell, then extract the content from the paragraphs
	if i["type"] == "tableCell" {
		if content, ok := i["content"].([]any); ok {
			for _, c := range content {
				if paragraph, ok := c.(map[string]any); ok {
					out = out + extractDescription(paragraph)
				}
			}
		}
	}

	if i["type"] == "inlineCard" {
		if attrs, ok := i["attrs"].(map[string]any); ok {
			if url, ok := attrs["url"].(string); ok {
				out = out + "<" + url + ">"
			}
		}
	}

	if i["type"] == "bulletList" {
		if content, ok := i["content"].([]any); ok {
			for _, c := range content {
				if m, ok := c.(map[string]any); ok {
					out = out + extractDescription(m)
				}
			}
		}
	}

	if i["type"] == "listItem" {
		if content, ok := i["content"].([]any); ok {
			for _, c := range content {
				if m, ok := c.(map[string]any); ok {
					out = out + " * " + extractDescription(m) + "\n"
				}
			}
		}
	}

	// If the type is a text, then extract the text
	if i["type"] == "text" {
		if text, ok := i["text"].(string); ok {
			if strings.Contains("Termination", text) {
				fmt.Println("Termination")
			}

			out = strings.TrimSpace(text)
		}
	}

	return out
}

type JiraIssueUser struct {
	Self         string `json:"self"`
	DisplayName  string `json:"displayName"`
	Active       bool   `json:"active"`
	EmailAddress string `json:"emailAddress,omitempty"`
}

func (j *JiraIssueUser) FromInterface(i any) error {
	m, ok := i.(map[string]any)
	if !ok {
		return fmt.Errorf("error converting to map")
	}

	j.Self = m["self"].(string)
	j.Active = m["active"].(bool)
	if m["emailAddress"] != nil {
		j.EmailAddress = m["emailAddress"].(string)
	}
	j.DisplayName = m["displayName"].(string)
	j.DisplayName = strings.Replace(j.DisplayName, ",", "", -1)
	j.DisplayName = strings.TrimSpace(j.DisplayName)

	return nil
}
