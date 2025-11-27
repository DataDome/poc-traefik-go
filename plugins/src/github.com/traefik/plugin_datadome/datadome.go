package plugin_datadome

import (
	"context"
	"fmt"
	"net/http"
	"os"

	modulego "github.com/traefik/plugin_datadome/modulego"
)

type Config struct {
	Endpoint string `json:"endpoint"`
}

func CreateConfig() *Config {
	return &Config{}
}

type DataDomePlugin struct {
	next           http.Handler
	name           string
	datadomeClient *modulego.Client
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	serverSideKey := os.Getenv("DATADOME_SERVER_SIDE_KEY")
	if serverSideKey == "" {
		return nil, fmt.Errorf("DATADOME_SERVER_SIDE_KEY environment variable is not set")
	}
	ddClient, err := modulego.NewClient(
		serverSideKey,
		modulego.WithEndpoint(config.Endpoint),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create DataDome client: %w", err)
	}

	return &DataDomePlugin{
		next:           next,
		name:           name,
		datadomeClient: ddClient,
	}, nil
}

func (m *DataDomePlugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// Example of payload, values should come from the req
	isBlocked, err := m.datadomeClient.DatadomeProtect(rw, req)

	if err != nil {
		fmt.Println("error when requesting DataDome", err)
	}

	if isBlocked {
		fmt.Println("request blocked by DataDome")
		return
	}

	m.next.ServeHTTP(rw, req)
}
