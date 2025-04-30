package models

import (
	"fmt"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

type AgentServiceDetails struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Protocol  string            `json:"protocol"`
	Host      string            `json:"host"`
	Port      uint16            `json:"port"`
	Metadata  map[string]string `json:"metadata"`
	SubAgents []string          `json:"sub_agents"`
}

func (s *AgentServiceDetails) ToMap() map[string]string {
	mapForm := map[string]string{
		"id":       s.ID,
		"name":     s.Name,
		"protocol": s.Protocol,
		"host":     s.Host,
		"port":     fmt.Sprintf("%d", s.Port),
	}

	// add our metadata to the map
	metadataPrefix := "metadata_"
	for key, value := range s.Metadata {
		mapForm[metadataPrefix+key] = value
	}

	// add our sub agents to the map, comma separated string
	subAgentsStr := ""
	for i, subAgent := range s.SubAgents {
		if i > 0 {
			subAgentsStr += ","
		}
		subAgentsStr += subAgent
	}
	if subAgentsStr != "" {
		mapForm["sub_agents"] = subAgentsStr
	} else {
		mapForm["sub_agents"] = ""
	}

	return mapForm
}

func (s *AgentServiceDetails) FromMap(m map[string]string) {
	s.ID = m["id"]
	s.Name = m["name"]
	s.Protocol = m["protocol"]
	s.Host = m["host"]
	port, err := strconv.Atoi(m["port"])
	if err != nil {
		log.Errorf("Invalid port value: %v", err)
		s.Port = 0
	} else {
		s.Port = uint16(port)
	}

	s.Metadata = make(map[string]string)
	metadataPrefix := "metadata_"
	for key, value := range m {
		if strings.HasPrefix(key, metadataPrefix) {
			metadataKey := strings.TrimPrefix(key, metadataPrefix)
			s.Metadata[metadataKey] = value
		}
	}

	subAgentsStr := m["sub_agents"]
	if subAgentsStr != "" {
		s.SubAgents = strings.Split(subAgentsStr, ",")
	} else {
		s.SubAgents = []string{}
	}
}
