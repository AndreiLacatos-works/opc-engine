package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AndreiLacatos/opc-engine/config"
	"github.com/AndreiLacatos/opc-engine/logging"
	nodeengine "github.com/AndreiLacatos/opc-engine/node-engine"
	opcnode "github.com/AndreiLacatos/opc-engine/node-engine/models/opc/opc_node"
	"github.com/AndreiLacatos/opc-engine/node-engine/serialization"
	opcserver "github.com/AndreiLacatos/opc-engine/opc-server"
	tcpserver "github.com/AndreiLacatos/opc-engine/tcp-server"
	"go.uber.org/zap"
)

var l *zap.Logger

func main() {
	c := config.GetConfig()
	l = logging.MakeLogger(c.LogLevel)
	defer l.Sync()
	l.Info(fmt.Sprintf("OPC Engine Simulator %s (built on %v)", c.Version, c.BuildTime))

	if c.ProjectPath == "" {
		l.Error("Missing project path")
		l.Error("Usage:")
		l.Error("\topc-engine-simulator /path/to/engine/config.opcroj")
		l.Error("\t\tOR")
		l.Error("\tset OPC_ENGINE_SIMULATOR_PROJECT_PATH environment variable")
		os.Exit(1)
	}

	content, err := os.ReadFile(c.ProjectPath)
	if err != nil {
		l.Fatal(fmt.Sprintf("error reading project file: %v", err))
	}
	jsonString := string(content)

	var structureModel serialization.OpcStructureModel

	err = json.Unmarshal([]byte(jsonString), &structureModel)
	if err != nil {
		l.Fatal(fmt.Sprintf("error decoding JSON: %v", err))
	}
	structure := structureModel.ToDomain(l)

	e := nodeengine.CreateNew(extractValueNodes(structure.Root), l, c.EngineDebugEnabled)

	opcServer := opcserver.CreateNew(opcserver.OpcServerConfig{
		ServerName:        "test-server",
		ServerEndpointUrl: c.ServerAddress,
		Port:              c.OpcServerPort,
		BuildInfo: opcserver.OpcServerBuildInfo{
			Version:   c.Version,
			BuildDate: c.BuildTime,
		},
	}, l)

	configServer := tcpserver.CreateNew(tcpserver.TcpServerConfig{
		Host: c.ServerAddress,
		Port: c.TcpServerPort,
	}, l)

	configServer.Setup()

	if err = opcServer.Setup(); err != nil {
		l.Fatal(fmt.Sprintf("could not set up OPC server: %v", err))
	}
	if err = opcServer.SetNodeStructure(structure); err != nil {
		l.Fatal(fmt.Sprintf("some nodes might not have been added correctly: %v", err))
	}

	stop := make(chan interface{})
	go func() {
		l.Info("starting opc server")
		opcServer.Start()
		l.Info("opc server stopped")
		stop <- ""
	}()

	go func() {
		l.Info("starting config tcp server")
		configServer.Start()
		l.Info("config tcp server stopped")
	}()
	go opcServer.Subscribe(e.EventChannel())
	go e.Start()

	go func() {
		commands := configServer.GetCommandChannel()
		response := configServer.GetResponseChannel()
		for {
			select {
			case <-stop:
				return
			case <-commands:
				response <- nil
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	go func() {
		waitTerminationSignal()

		configServer.Stop()
		e.Stop()
		time.Sleep(1 * time.Second)

		if err = opcServer.Stop(); err != nil {
			l.Fatal(fmt.Sprintf("could not stop OPC server: %v", err))
		}
	}()
	<-stop
	l.Info("program terminated")
}

func waitTerminationSignal() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)
	<-sigs
}

func extractValueNodes(r opcnode.OpcContainerNode) []opcnode.OpcValueNode {
	res := make([]opcnode.OpcValueNode, 0)

	for _, n := range r.Children {
		switch t := n.(type) {
		case *opcnode.OpcContainerNode:
			res = append(res, extractValueNodes(*t)...)
		case *opcnode.OpcValueNode:
			res = append(res, *t)
		}
	}
	return res
}
