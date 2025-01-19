package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/AndreiLacatos/opc-engine/config"
	"github.com/AndreiLacatos/opc-engine/logging"
	nodeengine "github.com/AndreiLacatos/opc-engine/node-engine"
	"github.com/AndreiLacatos/opc-engine/node-engine/models/opc"
	"github.com/AndreiLacatos/opc-engine/node-engine/serialization"
	opcserver "github.com/AndreiLacatos/opc-engine/opc-server"
	tcpserver "github.com/AndreiLacatos/opc-engine/tcp-server"
	"go.uber.org/zap"
)

var l *zap.Logger
var configServer tcpserver.TcpServer
var opcServer opcserver.OpcServer = nil
var nodeEngine nodeengine.ValueChangeEngine = nil

func main() {
	c := config.GetConfig()
	l = logging.MakeLogger(c.LogLevel)
	defer l.Sync()
	l.Info(fmt.Sprintf("OPC Engine Simulator %s (built on %v)", c.Version, c.BuildTime))

	configServer = tcpserver.CreateNew(tcpserver.TcpServerConfig{
		Host: c.ServerAddress,
		Port: c.TcpServerPort,
	}, l)

	configServer.Setup()
	go func() {
		if err := configServer.Start(); err != nil {
			l.Error(fmt.Sprintf("failed to start config server, reason: %v", err))
			os.Exit(1)
		}
	}()

	initialStructure := getInitialOpcStructure(c)
	if initialStructure != nil {
		setupOpc(c, initialStructure)
	}

	go func() {
		commands := configServer.GetCommandChannel()
		response := configServer.GetResponseChannel()
		for {
			s := <-commands
			if err := teardownOpc(); err != nil {
				l.Error(fmt.Sprintf("error tearing down OPC server, reason: %v", err))
				response <- err
			}
			if err := setupOpc(c, &s); err != nil {
				l.Error(fmt.Sprintf("error setting up OPC server, reason: %v", err))
				response <- err
			}
			response <- nil
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		waitTerminationSignal()

		teardownOpc()
		configServer.Stop()
		wg.Done()
	}()
	wg.Wait()
	l.Info("program terminated")
}

func waitTerminationSignal() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)
	<-sigs
}

func setupOpc(c config.Config, s *opc.OpcStructure) error {
	if s == nil {
		l.Warn("missing OPC structure, aborting setup")
		return nil
	}

	if opcServer != nil || nodeEngine != nil {
		l.Warn("OPC server or node engine not properly disposed, abortint setup")
		return nil
	}

	opcServer = opcserver.CreateNew(opcserver.OpcServerConfig{
		ServerName:        "test-server",
		ServerEndpointUrl: c.ServerAddress,
		Port:              c.OpcServerPort,
		BuildInfo: opcserver.OpcServerBuildInfo{
			Version:   c.Version,
			BuildDate: c.BuildTime,
		},
	}, l)

	if err := opcServer.Setup(); err != nil {
		l.Error(fmt.Sprintf("could not set up OPC server: %v", err))
		return err
	}
	if err := opcServer.SetNodeStructure(*s); err != nil {
		l.Error(fmt.Sprintf("some nodes might not have been added correctly: %v", err))
		return err
	}

	var opcStartErr error = nil
	go func() {
		if err := opcServer.Start(); err != nil {
			opcStartErr = err
		}
	}()
	time.Sleep(2 * time.Second)
	if opcStartErr != nil {
		l.Error(fmt.Sprintf("could not start OPC server, reason: %v", opcStartErr))
		return opcStartErr
	} else {
		l.Info("started OPC server")
	}

	nodeEngine = nodeengine.CreateNew(*s, l, c.EngineDebugEnabled)
	go opcServer.Subscribe(nodeEngine.EventChannel())
	go nodeEngine.Start()

	return nil
}

func teardownOpc() error {
	if opcServer == nil || nodeEngine == nil {
		l.Warn("OPC server or node engine not initialized, aborting teardown")
		return nil
	}

	nodeEngine.Stop()
	if err := opcServer.Stop(); err != nil {
		l.Warn(fmt.Sprintf("could not stop OPC server, reason: %v", err))
		return err
	}

	opcServer = nil
	nodeEngine = nil
	return nil
}

func getInitialOpcStructure(c config.Config) *opc.OpcStructure {
	if c.ProjectPath == "" {
		return nil
	}

	content, err := os.ReadFile(c.ProjectPath)
	if err != nil {
		l.Error(fmt.Sprintf("error reading project file: %v", err))
		return nil
	}

	jsonString := string(content)
	var structureModel serialization.OpcStructureModel
	err = json.Unmarshal([]byte(jsonString), &structureModel)
	if err != nil {
		l.Error(fmt.Sprintf("error decoding JSON: %v", err))
		return nil
	}
	structure := structureModel.ToDomain(l)
	return &structure
}
