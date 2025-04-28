package discovery

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"

	protocols "ioa-svc-disc-docker/pkg/agent-protocols"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type ServiceInfo struct {
	Name         string `json:"name"`
	Image        string `json:"image"`
	Protocol     string `json:"protocol"`
	Address      string `json:"address"`
	Port         int    `json:"port"`
	OpenapiSpec  string `json:"openapi_spec"`
	DocsEndpoint string `json:"docs_endpoint"`
}

type ServiceDiscovery interface {
	Scan() []*ServiceInfo
}

type DockerDiscovery struct {
	cli *client.Client
}

func NewDockerDiscovery() *DockerDiscovery {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Error creating Docker client: %v", err)
	}
	return &DockerDiscovery{cli: cli}
}

func (d *DockerDiscovery) Scan() []*ServiceInfo {
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

	var discoveredAgents []*ServiceInfo

	// Upload nodes to the graph store
	for _, container := range containers {
		svc, err := d.getContainerServiceInfo(container)
		if err != nil {
			log.Debugf("Failed to get service info for container %s: %v", container.ID, err)
			continue
		}
		discoveredAgents = append(discoveredAgents, svc)
	}

	log.Infof("Discovered %d agents", len(discoveredAgents))
	return discoveredAgents
}

func (d *DockerDiscovery) getContainerServiceInfo(container container.Summary) (*ServiceInfo, error) {
	if len(container.Ports) == 0 {
		log.Debugf("Container %s has no ports exposed and we rely on host networking", container.ID)
		return nil, fmt.Errorf("no ports found for container %s", container.ID)
	}

	// Detect the agent protocol
	protocol, err := protocols.DetectAgentProtocol(container)
	if err != nil {
		log.Errorf("Failed to detect agent protocol: %v", err)
		return nil, err
	}

	// recursively check for subagents
	if protocol == "ACP" {
		log.Info("Detected ACP protocol, checking for subagents...")
		// TODO: Implement subagent discovery
	} else if protocol == "AGP" {
		log.Info("Detected AGP protocol, checking for subagents...")
		// TODO: this is a placeholder for when we have an AGP control plane
	}

	host, port, _ := protocols.GetContainerAddress(container)

	return &ServiceInfo{
		Name:         container.Names[0],
		Image:        container.Image,
		Protocol:     protocol,
		Address:      host,
		Port:         int(port),
		OpenapiSpec:  fmt.Sprintf("http://%s:%d/openapi.json", host, port),
		DocsEndpoint: fmt.Sprintf("http://%s:%d/docs", host, port),
	}, nil
}
