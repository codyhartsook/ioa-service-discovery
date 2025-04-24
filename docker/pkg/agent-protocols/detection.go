package agentprotocols

import "fmt"

// DetectAgentProtocol uses the provided openapi spec to determine the agent protocol being used if any.
func DetectAgentProtocol(spec map[string]interface{}) (string, error) {
	// check if this is an ACP agent
	if _, ok := spec["openapi"]; ok {
		protocol, err := DetectACP(spec)
		if err == nil {
			return protocol, nil
		}
	}

	return "unknown", nil
}

// TODO: Move the protocol detection functions to their own files

// DetectACP checks if the provided OpenAPI spec is for the ACP protocol.
func DetectACP(spec map[string]interface{}) (string, error) {
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

// DetectAP checks if the provided OpenAPI spec is for the AP protocol.
func DetectAP(spec map[string]interface{}) (string, error) {
	return "", fmt.Errorf("AP protocol detection not implemented")
}

// DetectA2A checks if the provided OpenAPI spec is for the A2A protocol.
func DetectA2A(spec map[string]interface{}) (string, error) {
	return "", fmt.Errorf("A2A protocol detection not implemented")
}
