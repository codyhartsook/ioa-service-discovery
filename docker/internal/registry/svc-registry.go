package registry

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/codyhartsook/ioa-service-discovery/docker/pkg/models"

	"github.com/hashicorp/consul/api"
)

const (
	defaultConsulAddress = "localhost:8500"
)

type ConsulServiceMonitor struct {
	client *api.Client
}

func NewConsulMonitor() (*ConsulServiceMonitor, error) {
	// check env var for Consul address or use default
	consulAddress := os.Getenv("CONSUL_HTTP_ADDR")
	if consulAddress == "" {
		consulAddress = defaultConsulAddress
	}

	config := api.DefaultConfig()
	config.Address = consulAddress
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}
	return &ConsulServiceMonitor{client: client}, nil
}

func sanitizeServiceName(name string) string {
	rawName := strings.TrimPrefix(name, "/")
	dnsSafe := regexp.MustCompile("[^a-zA-Z0-9-]+")
	safeName := dnsSafe.ReplaceAllString(rawName, "-")
	if len(safeName) > 64 {
		safeName = safeName[:64]
	}
	return safeName
}

func agentDetailsToServiceRegistration(svc *models.AgentServiceDetails) *api.AgentServiceRegistration {
	record := &api.AgentServiceRegistration{
		ID:      svc.ID,
		Name:    sanitizeServiceName(svc.Name),
		Address: svc.Host,
		Port:    int(svc.Port),
		Tags:    []string{"agent"},
		Meta:    svc.ToMap(), // Add the entire svc as metadata
	}

	// TODO: Add health check if we have an http endpoint
	/*Check: &api.AgentServiceCheck{
		HTTP:     fmt.Sprintf("http://%s:%d/docs", info.Address, info.Port),
		Interval: "10s",
		Timeout:  "2s",
	},*/

	return record
}

func (c *ConsulServiceMonitor) RegisterService(svc *models.AgentServiceDetails) error {
	record := agentDetailsToServiceRegistration(svc)

	err := c.client.Agent().ServiceRegister(record)
	if err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}
	fmt.Printf("Registered service with Consul: %s\n", svc.Name)
	return nil
}
