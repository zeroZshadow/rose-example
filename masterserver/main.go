// masterserver project main.go
// The Masterserver is the server the client first connects to. It handles requests to create and join rooms.

package main

import (
	"flag"

	"github.com/op/go-logging"
	"github.com/zeroZshadow/rose"
	"github.com/zeroZshadow/rose-example/masterserver/client"
	"github.com/zeroZshadow/rose-example/masterserver/config"
	"github.com/zeroZshadow/rose-example/masterserver/node"
	"github.com/zeroZshadow/rose-example/shared"
)

var (
	configFile string
	logFile    string
	log        = logging.MustGetLogger("global")
)

func init() {
	flag.StringVar(&configFile, "config", "", "Path to config file")
	flag.StringVar(&logFile, "log", "", "Path to log file")
}

func main() {
	// Parse parameters
	flag.Parse()

	// Setup logging
	shared.InitLogger(logFile)
	defer shared.CloseLogger()

	// Initialize configuration
	cfg := config.New()

	// Overwrite configuration if file was passed
	if configFile != "" {
		err := cfg.FromFile(configFile)
		if err != nil {
			log.Warningf("Error while loading config: %s", err.Error())
			log.Notice("Loaded default config")
		} else {
			log.Noticef("Loaded config from file: %s", configFile)
		}
	} else {
		log.Notice("Loaded default config")
	}

	// Set as global config
	config.GlobalConfig = cfg

	// Setup handlers
	client.SetupMessageHandlers()
	node.SetupMessageHandlers()

	// Create protoserver without origin checking and listen on /ws
	server := rose.New(nil)
	server.Listen("/client", client.New)
	server.Listen("/cluster", node.New)

	// Setup listener
	err := server.Serve(cfg.Address)
	if err != nil {
		log.Fatalf("Unable to start server!\n%s", err.Error())
	}

	// Wait for things to Close
	server.Wait()
	log.Info("Stopped.")
}
