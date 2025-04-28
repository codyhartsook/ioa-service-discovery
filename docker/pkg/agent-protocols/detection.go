package agentprotocols

import (
	"fmt"
	"net/http"

	"github.com/docker/docker/api/types/container"
	log "github.com/sirupsen/logrus"
)

// DetectAgentProtocol uses the provided openapi spec to determine the agent protocol being used if any.
func DetectAgentProtocol(container container.Summary) (string, error) {
	// TODO: optimze the detection process by checking the container's labels or environment variables first
	// then check concurrently

	// detect acp
	protocol, err := DetectACP(container)
	if err == nil {
		log.Infof("Detected ACP protocol for container %s", container.Names[0])
		return protocol, nil
	}
	// detect mcp
	protocol, err = DetectMCP(container)
	if err == nil {
		log.Infof("Detected MCP protocol for container %s", container.Names[0])
		return protocol, nil
	}
	// detect ap
	protocol, err = DetectAP(container)
	if err == nil {
		log.Infof("Detected AP protocol for container %s", container.Names[0])
		return protocol, nil
	}
	// detect a2a
	protocol, err = DetectA2A(container)
	if err == nil {
		log.Infof("Detected A2A protocol for container %s", container.Names[0])
		return protocol, nil
	}

	return "unknown", fmt.Errorf("failed to detect agent protocol for container %s: %v", container.Names[0], err)
}

// TODO: Move the protocol detection functions to their own files

// DetectACP checks if the provided container exposes an ACP api.
func DetectACP(container container.Summary) (string, error) {
	spec, err := FetchOpenAPISpec(container)
	if err != nil {
		log.Infof("Error fetching OpenAPI spec for container %s: %v - ports: %v", container.Names[0], err, container.Ports)
		return "", err
	}

	if _, ok := spec["openapi"]; !ok {
		return "", fmt.Errorf("invalid OpenAPI spec: 'openapi' field is missing")
	}

	paths, ok := spec["paths"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid OpenAPI spec: 'paths' field is missing or not a map")
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
			return "", fmt.Errorf("invalid OpenAPI spec: missing path %s", path)
		}
	}

	return "ACP", nil
}

func DetectMCP(container container.Summary) (string, error) {
	host, port, err := GetContainerAddress(container)
	if err != nil {
		return "", fmt.Errorf("no ports found for container %s", container.ID)
	}

	url := fmt.Sprintf("http://%s:%d/sse", host, port)

	log.Debugf("Checking for MCP protocol at %s", url)
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to connect to %s: %v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get SSE stream from %s: %v", url, resp.Status)
	}

	log.Debugf("Successfully connected to %s", url)

	// Check for the presence of the "event" field in the SSE stream
	return "MCP", nil
}

// DetectAP checks if the provided OpenAPI spec is for the AP protocol.
func DetectAP(container container.Summary) (string, error) {
	return "", fmt.Errorf("AP protocol detection not implemented")
}

// DetectA2A checks if the provided OpenAPI spec is for the A2A protocol.
func DetectA2A(container container.Summary) (string, error) {
	return "", fmt.Errorf("A2A protocol detection not implemented")
}
