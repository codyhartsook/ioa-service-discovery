package registry

import (
	"fmt"
	disc "ioa-svc-disc-docker/internal/discovery"
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

func (c *ConsulServiceMonitor) RegisterService(info *disc.ServiceInfo) error {
	// Sanitize the service name to be DNS-compliant
	info.Name = sanitizeServiceName(info.Name)

	reg := &api.AgentServiceRegistration{
		Name:    info.Name,
		Address: info.Address,
		Port:    info.Port,
		Meta: map[string]string{
			"image":            info.Image,
			"protocol":         info.Protocol,
			"openapi_spec_url": info.OpenapiSpec,
			"openapi_docs_url": info.DocsEndpoint,
		},
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
	fmt.Printf("Registered service with Consul: %s\n", info.Name)
	return nil
}
