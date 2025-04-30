package agentprotocols

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/docker/docker/api/types/container"
	log "github.com/sirupsen/logrus"
)

func GetContainerAddress(container container.Summary) (string, uint16, error) {
	if len(container.Ports) == 0 {
		log.Debugf("Container %s has no ports exposed and we rely on host networking", container.ID)
		return "", 0, fmt.Errorf("no ports found for container %s", container.ID)
	}

	host := "localhost"
	port := container.Ports[0].PublicPort
	return host, port, nil
}

func FetchOpenAPISpec(host string, port uint16) (map[string]interface{}, error) {
	// TODO: get protocol schema from the container: ie http or https
	openapiURL := fmt.Sprintf("http://%s:%d/openapi.json", host, port)
	httpClient := &http.Client{Timeout: 3 * time.Second}
	resp, err := httpClient.Get(openapiURL)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get OpenAPI spec from %s: %v", openapiURL, resp.Status)
	}

	defer resp.Body.Close()

	// Process the OpenAPI spec here
	spec := make(map[string]interface{})
	if err := json.NewDecoder(resp.Body).Decode(&spec); err != nil {
		log.Errorf("Failed to decode OpenAPI spec: %v", err)
		return nil, err
	}

	return spec, nil
}
