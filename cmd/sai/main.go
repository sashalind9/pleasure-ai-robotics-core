package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/sashalind/sex-artifical-intelligence/pkg/core"
	"github.com/sashalind/sex-artifical-intelligence/pkg/diagnostics"
	"github.com/sashalind/sex-artifical-intelligence/pkg/safety"
)

// bozhe moy, main entry point of our glorious system
// we initialize everything here, da?
func main() {
	log.Println("Starting Sex Artificial Intelligence System v0.1.0")
	
	// initialize core systems blyat
	system, err := core.NewSystem()
	if err != nil {
		log.Fatalf("Failed to initialize core system: %v", err)
	}

	// safety first, tovarisch
	safety.InitializeSafetyProtocols(system)
	
	// diagnostic systems for when everything goes to blyat
	diagnostics.StartMonitoring(system)

	// graceful shutdown, like good vodka
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	<-sigChan
	log.Println("Shutting down systems... Do svidaniya!")
	system.Shutdown()
} 