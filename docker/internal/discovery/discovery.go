package discovery

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"

	protocols "github.com/codyhartsook/ioa-service-discovery/docker/pkg/agent-protocols"
	"github.com/codyhartsook/ioa-service-discovery/docker/pkg/models"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type ServiceDiscovery interface {
	Scan() map[string]*models.AgentServiceDetails
}

type DockerDiscovery struct {
	cli       *client.Client
	cache     map[string]*models.AgentServiceDetails
	detectors []protocols.ProtocolSniffer
}

func NewDockerDiscovery(detectors []protocols.ProtocolSniffer) *DockerDiscovery {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Error creating Docker client: %v", err)
	}

	return &DockerDiscovery{
		cache:     make(map[string]*models.AgentServiceDetails),
		cli:       cli,
		detectors: detectors}
}

func (d *DockerDiscovery) Scan() map[string]*models.AgentServiceDetails {
	// List only running containers
	containers, err := d.cli.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		log.Errorf("Error listing containers: %v", err)
		return nil
	}

	// If no running containers found
	if len(containers) == 0 {
		fmt.Println("No running containers found")
		return nil
	}

	discoveredAgents := make(map[string]*models.AgentServiceDetails)

	for _, container := range containers {
		// check if we have cached this container
		if _, ok := d.cache[container.ID]; ok {
			discoveredAgents[container.ID] = d.cache[container.ID]
			log.Debugf("Using cached service info for container %s", container.ID)
			continue
		}
		agentSvc, err := d.getAgentProtocolDetails(container)
		if err != nil {
			log.Debugf("Failed to get service info for container %s: %v", container.ID, err)
			continue
		}
		discoveredAgents[container.ID] = agentSvc
		d.cache[container.ID] = agentSvc
	}

	log.Infof("Discovered %d agents", len(discoveredAgents))
	return discoveredAgents
}

func (d *DockerDiscovery) getAgentProtocolDetails(container container.Summary) (*models.AgentServiceDetails, error) {
	if len(container.Ports) == 0 {
		log.Debugf("Container %s has no ports exposed and we rely on host networking", container.ID)
		return nil, fmt.Errorf("no ports found for container %s", container.ID)
	}

	// TODO: optimize this process
	for _, detector := range d.detectors {
		svc, err := detector.SniffProtocol(container)
		if err != nil {
			log.Debugf("Failed to sniff protocol for container %s: %v", container.ID, err)
			continue
		}

		log.Infof("Detected protocol %s for container %s", svc.Protocol, container.ID)

		svc.Metadata["container_id"] = container.ID
		svc.Metadata["image"] = container.Image
		return svc, nil
	}

	return nil, fmt.Errorf("no protocol sniffer found for container %s", container.ID)
}
