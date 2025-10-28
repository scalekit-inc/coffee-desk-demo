package handlers

import (
	"fmt"
	"log"
	"sync"

	"github.com/scalekit-inc/scalekit-sdk-go/v2"
)

var (
	globalClient scalekit.Scalekit
	clientOnce   sync.Once
	clientErr    error
)

// GetScaleKitClient returns the global ScaleKit client instance, initializing it if necessary
func GetScaleKitClient() (scalekit.Scalekit, error) {
	clientOnce.Do(func() {
		config, err := getScaleKitConfig()
		if err != nil {
			clientErr = fmt.Errorf("failed to get ScaleKit config: %w", err)
			return
		}

		log.Printf("Initializing global ScaleKit client with environment: %s", config.EnvURL)
		globalClient = scalekit.NewScalekitClient(config.EnvURL, config.ClientID, config.ClientSecret)
		clientErr = nil
	})

	return globalClient, clientErr
}

// InitializeScaleKitClient explicitly initializes the global ScaleKit client
func InitializeScaleKitClient() error {
	_, err := GetScaleKitClient()
	return err
}
