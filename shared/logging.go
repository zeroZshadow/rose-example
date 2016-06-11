package shared

import (
	"os"

	"github.com/op/go-logging"
)

var fancyFormat = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{shortfunc} > %{level:.4s} %{color:reset} %{message}`,
)

var fileFormat = logging.MustStringFormatter(
	`%{time:15:04:05} %{shortfunc} > %{level:.4s} %{message}`,
)

var log = logging.MustGetLogger("global")
var errfile *os.File

// InitLogger initialize the logger
func InitLogger(filename string) {
	// StdErr backend
	backendStderr := logging.NewLogBackend(os.Stderr, "", 0)
	backendStderrFormatter := logging.NewBackendFormatter(backendStderr, fancyFormat)
	logging.SetBackend(backendStderrFormatter)

	if filename != "" {
		// Open file to log to
		file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Warning("Unable to log to file", err)
			return
		}
		errfile = file

		// Add a backend that writes to file
		backendFile := logging.NewLogBackend(errfile, "", 0)
		backendFileFormatter := logging.NewBackendFormatter(backendFile, fileFormat)

		// Only write errors or worse
		backendFileLeveled := logging.AddModuleLevel(backendFileFormatter)
		backendFileLeveled.SetLevel(logging.ERROR, "")

		// Set backends
		logging.SetBackend(backendFileLeveled, backendStderrFormatter)
	}

	log.Info("Logging initialized")
}

// CloseLogger Close open logger files
func CloseLogger() {
	if errfile != nil {
		errfile.Close()
	}
}
