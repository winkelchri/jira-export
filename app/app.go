package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"jira-export/pkg/jira"
	"jira-export/pkg/logger"
	"jira-export/pkg/output"
	"jira-export/pkg/secrets"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	username   string
	token      string
	url        string
	jql        string
	outputDir  string
	maxResults int
)

const (
	MAX_RESULTS = 100
)

func init() {
	err := godotenv.Load()
	if err != nil {
		logger.Logger.Error("Error loading .env file", "error", err)
	}

	// Bind environment variables
	viper.SetEnvPrefix("JIRA_EXPORT")
	viper.BindEnv("username")
	viper.BindEnv("token")
	viper.BindEnv("url")
	viper.BindEnv("jql")

	// Bind flags
	RootCmd.PersistentFlags().StringVarP(&username, "username", "u", viper.GetString("username"), "Jira username")
	RootCmd.PersistentFlags().StringVarP(&token, "token", "t", viper.GetString("token"), "Jira token")
	RootCmd.PersistentFlags().StringVarP(&url, "url", "r", viper.GetString("url"), "Jira URL")
	RootCmd.PersistentFlags().StringVarP(&jql, "jql", "j", viper.GetString("jql"), "JQL query")
	// Trim surrounding single quotes if present
	jql = strings.Trim(jql, "'")
	RootCmd.PersistentFlags().StringVarP(&outputDir, "output", "o", "dist/jira/results", "Output directory")
	RootCmd.PersistentFlags().IntVarP(&maxResults, "max-results", "m", 100, "Max results")
}

var RootCmd = &cobra.Command{
	Use:   "jira-export",
	Short: "Export Jira issues to CSV and JSON",
	Long:  `Export Jira issues to CSV and JSON`,
	Run: func(cmd *cobra.Command, args []string) {
		if username == "" {
			logger.Logger.Error("Missing username")
			os.Exit(1)
		}

		if token == "" {
			logger.Logger.Error("Missing token")
			os.Exit(1)
		}

		if url == "" {
			logger.Logger.Error("Missing URL")
			os.Exit(1)
		}

		if jql == "" {
			logger.Logger.Error("Missing JQL query")
			os.Exit(1)
		}

		secrets := secrets.Secrets{
			Username: username,
			Token:    token,
			URL:      url,
		}

		err := Export(jql, outputDir, secrets, maxResults)
		if err != nil {
			logger.Logger.Error("Export failed", "error", err)
		}

	},
}

func Export(jqlQuery string, outputDir string, secrets secrets.Secrets, maxResults int) error {

	// Create a JiraAPI object
	jiraAPI := jira.NewJiraAPI(secrets, maxResults)
	data := jira.JiraSearchResults{}
	outputFileName := "jira-export"

	logger.Logger.Debug("Exporting Jira issues", "jql", jqlQuery)

	data, err := jiraAPI.GetFilterResults(jqlQuery)
	if err != nil {
		return fmt.Errorf("error getting filter results: %v", err)
	}

	issues, err := data.IssuesToJiraIssues()
	if err != nil {
		return fmt.Errorf("error converting issues: %v", err)
	}

	logger.Logger.Info("Exported Jira issues", "count", len(issues))

	// Write the issues to a JSON file
	jsonData, err := json.Marshal(issues)
	if err != nil {
		return fmt.Errorf("error marshalling json: %v", err)
	}

	r := io.NopCloser(bytes.NewReader(jsonData))
	jsonFile := fmt.Sprintf("%s/%s.json", outputDir, outputFileName)
	err = output.StoreJSON(r, jsonFile)
	if err != nil {
		return fmt.Errorf("error storing json: %v", err)
	}

	// Write the issues to a csv file
	csvFile := fmt.Sprintf("%s/%s.csv", outputDir, outputFileName)
	err = issues.WriteCSV(csvFile)
	if err != nil {
		return fmt.Errorf("error writing csv: %v", err)
	}

	return nil
}
