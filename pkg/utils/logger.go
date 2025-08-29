package utils

import (
	"fmt"
	"os"
	"time"
)

// Logger for debug messages
var (
	isVerbose = false
	logFile   *os.File
)

// Log prints debug messages to the log file if verbose mode is enabled
func Log(text string, args ...interface{}) {
	if isVerbose && logFile != nil {
		fmt.Fprintf(logFile, text+"\n", args...)
	}
}

// InitLogger initializes the logging system
func InitLogger(verbose bool) {
	isVerbose = verbose

	if verbose {
		// Create log filename with current date
		now := time.Now()
		logFileName := fmt.Sprintf("/tmp/awp_%s.log", now.Format("2006-01-02"))

		var err error
		logFile, err = os.Create(logFileName)
		if err != nil {
			fmt.Printf("Error creating log file: %v\n", err)
			return
		}

		Log("Verbose logging enabled")
	}
}

// CloseLogger closes the log file if it's open
func CloseLogger() {
	if logFile != nil {
		logFile.Close()
	}
}
