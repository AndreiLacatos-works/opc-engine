package opcserver

import "time"

type OpcServer interface {
	Setup() error
	Start() error
	Stop() error
}

type OpcServerBuildInfo struct {
	Version   string
	BuildDate time.Time
}

type OpcServerConfig struct {
	ServerName        string
	ServerEndpointUrl string
	Port              int32
	BuildInfo         OpcServerBuildInfo
}

func CreateNew(c OpcServerConfig) (OpcServer, error) {
	s := &opcServerImpl{Config: c}
	return s, nil
}
