package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/viper"
	"github.com/VeilNet-PTY-LTD/veilnet"
	"github.com/VeilNet-PTY-LTD/veilnet-portal/portal"
)

func main() {
	// Initialize Viper
	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("yaml")   // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(".")      // look for config in the working directory

	// Set environment variable prefix
	viper.SetEnvPrefix("VEILNET")
	viper.AutomaticEnv() // read in environment variables that match

	// Set defaults
	viper.SetDefault("guardian_url", "https://guardian.veilnet.org")
	viper.SetDefault("public", true)

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			veilnet.Logger.Sugar().Warn("Config file not found, using environment variables and defaults")
		} else {
			veilnet.Logger.Sugar().Errorf("Error reading config file: %v", err)
		}
	} else {
		veilnet.Logger.Sugar().Info("Using config file:", viper.ConfigFileUsed())
	}

	// Get configuration values
	guardianURL := viper.GetString("guardian_url")
	anchorToken := viper.GetString("anchor_token")
	anchorName := viper.GetString("anchor_name")
	domainName := viper.GetString("domain_name")
	region := viper.GetString("region")
	public := viper.GetBool("public")

	// Create a new portal
	p := portal.NewPortal()

	// Start the portal
	err := p.Start(guardianURL, anchorToken, anchorName, domainName, region, "eth0", public)
	if err != nil {
		veilnet.Logger.Sugar().Errorf("Failed to start portal: %v", err)
		panic(err)
	}

	// Wait for the portal to stop
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Give the daemon time to clean up
	veilnet.Logger.Sugar().Info("Received shutdown signal, shutting down...")

	// Create a channel to signal when cleanup is done
	shutdownComplete := make(chan bool, 1)

	go func() {
		p.Stop()
		shutdownComplete <- true
	}()

	// Wait for cleanup with timeout
	select {
	case <-shutdownComplete:
		veilnet.Logger.Sugar().Info("Cleanup completed successfully")
	case <-time.After(10 * time.Second):
		veilnet.Logger.Sugar().Warn("Cleanup timeout, forcing exit")
	}
}
