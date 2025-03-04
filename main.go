package main

import (
	"jira-export/app"
	"jira-export/pkg/logger"
)

func main() {

	logger.Logger.Info("Starting Jira Export")

	// Use the cobra root cmd to execute the app
	app.RootCmd.Execute()
}
