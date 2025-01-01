package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AndreiLacatos/opc-engine/logging"
	nodeengine "github.com/AndreiLacatos/opc-engine/node-engine"
	opcnode "github.com/AndreiLacatos/opc-engine/node-engine/models/opc/opc_node"
	"github.com/AndreiLacatos/opc-engine/node-engine/serialization"
	opcserver "github.com/AndreiLacatos/opc-engine/opc-server"
	"go.uber.org/zap"
)

var l *zap.Logger

func main() {
	l = logging.MakeLogger()
	defer l.Sync()

	if len(os.Args) != 2 {
		l.Error("Wrong number of arguments")
		l.Error("Usage: opc-engine-simulator /path/to/engine/config.opcroj")
		os.Exit(1)
	}

	input := os.Args[1]
	content, err := os.ReadFile(input)
	if err != nil {
		l.Fatal(fmt.Sprintf("error reading file: %v", err))
	}
	jsonString := string(content)

	var structureModel serialization.OpcStructureModel

	err = json.Unmarshal([]byte(jsonString), &structureModel)
	if err != nil {
		l.Fatal(fmt.Sprintf("error decoding JSON: %v", err))
	}
	structure := structureModel.ToDomain()

	e := nodeengine.CreateNew(extractValueNodes(structure.Root), l)

	s, err := opcserver.CreateNew(opcserver.OpcServerConfig{
		ServerName:        "test-server",
		ServerEndpointUrl: getIpAddress(),
		Port:              39056,
		BuildInfo: opcserver.OpcServerBuildInfo{
			Version:   "0.0.1",
			BuildDate: time.Now().UTC(),
		},
	}, l)

	if err != nil {
		l.Fatal(fmt.Sprintf("could not create OPC server: %v", err))
	}
	if err = s.Setup(); err != nil {
		l.Fatal(fmt.Sprintf("could not set up OPC server: %v", err))
	}
	if err = s.SetNodeStructure(structure); err != nil {
		l.Fatal(fmt.Sprintf("some nodes might not have been added correctly: %v", err))
	}

	stop := make(chan interface{})
	go func() {
		l.Info("starting opc server")
		s.Start()
		l.Info("opc server stopped")
		stop <- ""
	}()

	go s.Subscribe(e.EventChannel())
	e.Start()

	go func() {
		waitTerminationSignal()

		e.Stop()
		time.Sleep(1 * time.Second)

		if err = s.Stop(); err != nil {
			l.Fatal(fmt.Sprintf("could not stop OPC server: %v", err))
		}
	}()
	<-stop
	l.Info("program terminated")
}

func getIpAddress() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		l.Fatal(fmt.Sprintf("error retrieving network interfaces: %v", err))
	}

	candidates := make([]string, 0)
	for _, iface := range interfaces {
		// skip down or loopback interfaces
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		// get addresses associated with the interface
		addrs, err := iface.Addrs()
		if err != nil {
			fmt.Printf("error getting addresses for interface %s: %v\n", iface.Name, err)
			continue
		}

		// print each address
		for _, addr := range addrs {
			var ip net.IP

			// extract the IP address from the address
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// skip IPv6 or link-local addresses
			if ip == nil || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
				continue
			}

			// only consider IPv4
			if ip.To4() != nil {
				candidates = append(candidates, ip.String())
			}
		}
	}

	return candidates[0]
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
