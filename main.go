package main

import (
	"jira-export/app"
)

func main() {
	// Use the cobra root cmd to execute the app
	app.RootCmd.Execute()
}
