package main

import (
	"os"
	"strconv"
	"time"

	"github.com/codyhartsook/ioa-service-discovery/docker/internal/discovery"
	"github.com/codyhartsook/ioa-service-discovery/docker/internal/registry"
	agentprotocols "github.com/codyhartsook/ioa-service-discovery/docker/pkg/agent-protocols"

	log "github.com/sirupsen/logrus"
)

const (
	defaultDiscoveryInterval = 10 // default discovery interval in seconds
)

var (
	discoveryInterval int
	err               error
)

func main() {
	if os.Getenv("DISCOVERY_INTERVAL") == "" {
		discoveryInterval = defaultDiscoveryInterval
	} else {
		discoveryInterval, err = strconv.Atoi(os.Getenv("DISCOVERY_INTERVAL"))
		if err != nil {
			log.Fatalf("Invalid DISCOVERY_INTERVAL value: %v", err)
		}
	}

	log.Infof("Discovery interval set to %d seconds", discoveryInterval)

	// add all protocol sniffers here
	// TODO: use flags or env vars to configure the protocol sniffers
	protocolSniffers := []agentprotocols.ProtocolSniffer{
		&agentprotocols.ACPSniffer{},
		&agentprotocols.MCPSniffer{},
	}

	svcDiscovery := discovery.NewDockerDiscovery(protocolSniffers)
	svcRegistry, err := registry.NewConsulMonitor()
	if err != nil {
		log.Fatalf("Error creating Consul client: %v", err)
	}

	ticker := time.NewTicker(time.Duration(discoveryInterval) * time.Second)
	for range ticker.C {
		log.Info("Starting service discovery round...")
		services := svcDiscovery.Scan()

		for _, info := range services {
			err := svcRegistry.RegisterService(info)
			if err != nil {
				log.Errorf("Failed to register service %s: %v", info.ID, err)
			}
		}
	}
}
