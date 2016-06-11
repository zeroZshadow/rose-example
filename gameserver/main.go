// gameserver project main.go
// The GameServer runs the rooms that users play in

package main

import (
	"flag"

	"github.com/zeroZshadow/rose"
	"github.com/zeroZshadow/rose-example/gameserver/client"
	"github.com/zeroZshadow/rose-example/gameserver/config"
	"github.com/zeroZshadow/rose-example/gameserver/master"
	"github.com/zeroZshadow/rose-example/gameserver/node"
	"github.com/zeroZshadow/rose-example/gameserver/room"
	"github.com/zeroZshadow/rose-example/shared"

	"github.com/op/go-logging"
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
	SetupMessageHandlers()
	room.SetupMessageHandlers()

	// Create Server without origin checking and listen on /ws
	server := rose.New()
	server.Listen("/ws", client.New)

	// Setup listener
	err := server.Serve(cfg.Address)
	if err != nil {
		log.Fatalf("Unable to start server!\n%s", err.Error())
	}

	// Get port
	port, err := server.Port()
	if err != nil {
		log.Fatalf("Unable to get port for server!\n%s", err.Error())
	}

	log.Noticef("Gameserver serving on port %d", port)

	// Connect to the Master server
	node.Instantiate(server, cfg.Region, cfg.MasterAddress, port, master.New)
	node.Instance.Start()

	// Wait for things to Close
	server.Wait()
	log.Info("Stopped.")
}
