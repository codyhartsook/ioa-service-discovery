package agentprotocols

import (
	"context"
	"fmt"
	"time"

	"ioa-svc-disc-docker/pkg/models"

	"github.com/docker/docker/api/types/container"
	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	log "github.com/sirupsen/logrus"
)

type ProtocolSniffer interface {
	SniffProtocol(container container.Summary) (*models.AgentServiceDetails, error)
}

type ACPSniffer struct{}

func (s *ACPSniffer) SniffProtocol(container container.Summary) (*models.AgentServiceDetails, error) {
	host, port, err := GetContainerAddress(container)
	if err != nil {
		log.Infof("Error getting container address for %s: %v", container.Names[0], err)
		return nil, err
	}
	spec, err := FetchOpenAPISpec(host, port)
	if err != nil {
		log.Infof("Error fetching OpenAPI spec for container %s: %v - ports: %v", container.Names[0], err, container.Ports)
		return nil, err
	}

	if _, ok := spec["openapi"]; !ok {
		return nil, fmt.Errorf("invalid OpenAPI spec: 'openapi' field is missing")
	}

	paths, ok := spec["paths"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid OpenAPI spec: 'paths' field is missing or not a map")
	}

	// Define required paths
	requiredPaths := []string{
		"/agents/search",
		"/agents/{agent_id}/descriptor",
		"/agents/{agent_id}/openapi",
		"/runs/wait",
	}

	// Check if each required path exists
	for _, path := range requiredPaths {
		if _, ok := paths[path]; !ok {
			return nil, fmt.Errorf("invalid OpenAPI spec: missing path %s", path)
		}
	}

	return &models.AgentServiceDetails{
		ID:       container.ID,
		Name:     container.Names[0], // could try to get name from the spec but will likely need an api key
		Protocol: "ACP",
		Host:     host,
		Port:     port,
		Metadata: map[string]string{
			"name":    container.Names[0],
			"image":   container.Image,
			"version": container.ImageID,
			"openapi": fmt.Sprintf("http://%s:%d/openapi.json", host, port),
			"docs":    fmt.Sprintf("http://%s:%d/docs", host, port),
		},
	}, nil
}

type MCPSniffer struct{}

func (s *MCPSniffer) SniffProtocol(container container.Summary) (*models.AgentServiceDetails, error) {
	host, port, err := GetContainerAddress(container)
	if err != nil {
		return nil, fmt.Errorf("no ports found for container %s", container.ID)
	}

	serverUrl := fmt.Sprintf("http://%s:%d/sse", host, port)

	client, err := mcpclient.NewSSEMCPClient(serverUrl)
	if err != nil {
		// dont log anything, likely just not an mcp server
		return nil, fmt.Errorf("failed to create MCP client: %v", err)
	}

	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start the client
	if err := client.Start(ctx); err != nil {
		log.Errorf("Failed to start MCP client: %v", err)
		return nil, fmt.Errorf("failed to start MCP client: %v", err)
	}

	// Initialize
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "ioa-service-discovery",
		Version: "1.0.0",
	}

	result, err := client.Initialize(ctx, initRequest)
	if err != nil {
		log.Errorf("Failed to initialize MCP client: %v", err)
		return nil, fmt.Errorf("failed to initialize MCP client: %v", err)
	}

	// get server metadata
	// result.ServerInfo.Name

	// get server tools
	// result.ServerInfo.Tools

	// get server prompts

	// get server resources

	log.Infof("MCP server info: %v", result.ServerInfo)

	return &models.AgentServiceDetails{
		Protocol: "MCP",
		ID:       container.ID,
		Name:     result.ServerInfo.Name,
		Host:     host,
		Port:     port,
		Metadata: map[string]string{
			"image":   container.Image,
			"version": result.ProtocolVersion,
		},
	}, nil
}

type APSniffer struct{}

func (s *APSniffer) SniffProtocol(container container.Summary) (*models.AgentServiceDetails, error) {
	return nil, fmt.Errorf("AP protocol detection not implemented")
}

type A2ASniffer struct{}

func (s *A2ASniffer) SniffProtocol(container container.Summary) (*models.AgentServiceDetails, error) {
	return nil, fmt.Errorf("A2A protocol detection not implemented")
}
