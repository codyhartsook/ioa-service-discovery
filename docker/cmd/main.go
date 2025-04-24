package main

import (
	"ioa-svc-disc-docker/internal/discovery"
	"ioa-svc-disc-docker/internal/registry"
	"os"
	"strconv"
	"time"

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

	svcDiscovery := discovery.NewDockerDiscovery()
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
				log.Errorf("Failed to register service %s: %v", info.Name, err)
			}
		}
	}
}
