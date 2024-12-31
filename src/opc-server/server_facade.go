package opcserver

import "time"

type OpcServer interface {
	Start() error
	Stop() error
}

type OpcServerBuildInfo struct {
	Version string
	BuildDate time.Time
}

type OpcServecConfig struct {
	ServerName string
	ServerEndpointUrl string
	BuildInfo OpcServerBuildInfo
}

func CreateNew(c OpcServecConfig) (OpcServer, error) {
	s := &opcServerImpl{Config: c}
	return s, nil
}
