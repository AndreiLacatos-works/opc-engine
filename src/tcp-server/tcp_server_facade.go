package tcpserver

import (
	"github.com/AndreiLacatos/opc-engine/node-engine/models/opc"
	"go.uber.org/zap"
)

type TcpServer interface {
	Setup()
	Start() error
	GetCommandChannel() chan opc.OpcStructure
	GetResponseChannel() chan error
	Stop() error
}

type TcpServerConfig struct {
	Host string
	Port uint16
}

func CreateNew(c TcpServerConfig, l *zap.Logger) TcpServer {
	return &TcpServerImpl{
		Host:   c.Host,
		Port:   c.Port,
		Logger: l.Named("TCPSRV"),
	}
}
