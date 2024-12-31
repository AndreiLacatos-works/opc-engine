package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/AndreiLacatos/opc-engine/node-engine/serialization"
	opcserver "github.com/AndreiLacatos/opc-engine/opc-server"
)

func main() {
	input := os.Args[1]
	content, err := os.ReadFile(input)
	if err != nil {
		log.Fatalf("error reading file: %v", err)
	}
	jsonString := string(content)

	var structure serialization.OpcStructureModel

	err = json.Unmarshal([]byte(jsonString), &structure)
	if err != nil {
		log.Fatalf("error decoding JSON: %v", err)
	}

	fmt.Printf("%v\n", structure)
	r := structure.Root.ToDomain()
	fmt.Printf("%v\n", r)
	s, err := opcserver.CreateNew(opcserver.OpcServecConfig{
		ServerName: "test-server",
		ServerEndpointUrl: getIpAddress(),
		BuildInfo: opcserver.OpcServerBuildInfo{
			Version: "0.0.1",
			BuildDate: time.Now().UTC(),
		},
	})
	if err != nil {
		log.Fatalf("could not create OPC server: %v", err)
	}
	if err = s.Start(); err != nil {
		log.Fatalf("could not start OPC server: %v", err)
	}
	if err = s.Stop(); err != nil {
		log.Fatalf("could not stop OPC server: %v", err)
	}
}

func getIpAddress() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Fatalf("error retrieving network interfaces: %v\n", err)
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
