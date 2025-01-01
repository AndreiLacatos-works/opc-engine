package opcserver

import (
	"time"

	nodeengine "github.com/AndreiLacatos/opc-engine/node-engine"
	"github.com/AndreiLacatos/opc-engine/node-engine/models/opc"
	"go.uber.org/zap"
)

type OpcServer interface {
	Setup() error
	SetNodeStructure(opc.OpcStructure) error
	SetNodeValues() error
	Start() error
	Subscribe(chan nodeengine.NodeValueChange)
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

func CreateNew(c OpcServerConfig, l *zap.Logger) (OpcServer, error) {
	s := &opcServerImpl{Config: c, Logger: l.Named("OPCSRV")}
	return s, nil
}
