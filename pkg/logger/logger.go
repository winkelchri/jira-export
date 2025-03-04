package logger

import (
	"os"

	"github.com/charmbracelet/log"
)

var Logger *log.Logger

func init() {
	Logger = log.New(os.Stdout)

	if os.Getenv("ENVIRONMENT") == "dev" {
		Logger.SetLevel(log.DebugLevel)
		Logger.SetReportTimestamp(true)
		Logger.SetReportCaller(true)
	} else {
		Logger.SetFormatter(log.JSONFormatter)
		Logger.SetLevel(log.InfoLevel)
		Logger.SetReportTimestamp(true) // or false, depending on your needs
		Logger.SetReportCaller(false)   // Generally, caller info is less
	}

}
