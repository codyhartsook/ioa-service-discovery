package registry

import (
	"fmt"
	protocols "ioa-svc-disc-docker/pkg/agent-protocols"
	"os"
	"regexp"
	"strings"

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

func (c *ConsulServiceMonitor) RegisterService(svc *protocols.AgentServiceDetails) error {
	// Sanitize the service name to be DNS-compliant
	svc.Name = sanitizeServiceName(svc.Name)

	reg := &api.AgentServiceRegistration{
		Name:    svc.Name,
		ID:      svc.ID,
		Address: svc.Host,
		Port:    int(svc.Port),
		Meta:    svc.Metadata,
		/*Check: &api.AgentServiceCheck{
			HTTP:     fmt.Sprintf("http://%s:%d/docs", info.Address, info.Port),
			Interval: "10s",
			Timeout:  "2s",
		},*/
	}

	err := c.client.Agent().ServiceRegister(reg)
	if err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}
	fmt.Printf("Registered service with Consul: %s\n", svc.Name)
	return nil
}
