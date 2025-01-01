package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/AndreiLacatos/opc-engine/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// set from build flags
var (
	version   string
	buildTime string
)

var l *zap.Logger

type Config struct {
	LogLevel      zapcore.Level
	Version       string
	BuildTime     time.Time
	ProjectPath   string
	ServerAddress string
	ServerPort    uint16
}

func GetConfig() Config {
	level := getLogLevel()
	l = logging.MakeLogger(level).Named("config")
	build, _ := time.Parse(time.DateTime, buildTime)

	return Config{
		LogLevel:      level,
		Version:       version,
		BuildTime:     build,
		ProjectPath:   getProjectPath(),
		ServerPort:    getPort(),
		ServerAddress: getIpAddress(),
	}
}

func getLogLevel() zapcore.Level {
	l := getTrimmedEnvVar("OPC_ENGINE_SIMULATOR_LOG_LEVEL")

	switch strings.ToLower(l) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	}

	return zapcore.InfoLevel
}

func getProjectPath() string {
	p := getTrimmedEnvVar("OPC_ENGINE_SIMULATOR_PROJECT_PATH")
	if p != "" {
		l.Debug(fmt.Sprintf("using project path %s from environment", p))
		return p
	}
	if len(os.Args) >= 2 {
		l.Debug(fmt.Sprintf("using project path %s from command line arguments", os.Args[1]))
		return os.Args[1]
	}
	l.Warn("project path not set in environment and no command line argument")
	return ""
}

func getPort() uint16 {
	s := getTrimmedEnvVar("OPC_ENGINE_SIMULATOR_SERVER_PORT")
	if p, err := strconv.ParseUint(s, 10, 16); err != nil {
		defaultPort := uint16(39056)
		l.Warn(fmt.Sprintf("invalid port %s, defaulting to %d", s, defaultPort))
		return defaultPort
	} else {
		l.Debug(fmt.Sprintf("got server port %d from environment", p))
		return uint16(p)
	}
}

func getTrimmedEnvVar(s string) string {
	return strings.Trim(os.Getenv(s), " \t")
}

func getIpAddress() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		l.Fatal(fmt.Sprintf("error retrieving network interfaces: %v", err))
	}

	target := strings.ToLower(getTrimmedEnvVar("OPC_ENGINE_SIMULATOR_NETWORK_INTERFACE"))
	if target != "" {
		l.Debug(fmt.Sprintf("targeting network interface %s", target))
	} else {
		l.Warn("network interface not defined, defaulting to first")
	}

	candidates := make([]string, 0)
	for _, iface := range interfaces {
		// skip down or loopback interfaces
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			l.Debug(fmt.Sprintf("skipping %s (loopback or interface is down)", iface.Name))
			continue
		}

		if target != "" && strings.ToLower(iface.Name) != target {
			l.Debug(fmt.Sprintf("%s is not the target, skipping", iface.Name))
			continue
		}

		// get addresses associated with the interface
		addrs, err := iface.Addrs()
		if err != nil {
			l.Debug(fmt.Sprintf("error getting addresses for interface %s: %v", iface.Name, err))
			continue
		}

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
				l.Debug(fmt.Sprintf("no IP (or is loopback/link-local unicast) for interface %s address %s, skipping",
					iface.Name, addr.String()))
				continue
			}

			// only consider IPv4
			if ip.To4() != nil {
				l.Debug(fmt.Sprintf("found candidate interface %s address %s IP %s", iface.Name, addr.String(), ip.String()))
				candidates = append(candidates, ip.String())
			}
		}
	}

	if len(candidates) > 0 {
		l.Info(fmt.Sprintf("determined IP address %s", candidates[0]))
		return candidates[0]
	} else {
		l.Warn("no candidate interfaces found, defaulting to localhost")
		return "localhost"
	}
}
