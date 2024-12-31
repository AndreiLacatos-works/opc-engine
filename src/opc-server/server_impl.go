package opcserver

import (
	"encoding/json"
	"log"
)


type opcServerImpl struct {
	Config OpcServecConfig
}

func (s *opcServerImpl) Start() error {
	log.Println("starting OPC server")
	configJson, err :=json.MarshalIndent(s.Config, "", "  ")
	if err != nil {
		log.Printf("failed to convet config to JSON: %v", err)
		return err
	}
	log.Printf("config: %s\n", string(configJson))
	return nil
}

func (s *opcServerImpl) Stop() error {
	log.Printf("stopping OPC server")
	return nil
}
