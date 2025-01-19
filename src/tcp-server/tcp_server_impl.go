package tcpserver

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/AndreiLacatos/opc-engine/node-engine/models/opc"
	"github.com/AndreiLacatos/opc-engine/tcp-server/serialization"
	"go.uber.org/zap"
)

type TcpServerImpl struct {
	Host       string
	Port       uint16
	Logger     *zap.Logger
	Listener   *net.Listener
	Done       chan bool
	Command    chan opc.OpcStructure
	Response   chan error
	CommandMap map[string]func(any) error
}

func (s *TcpServerImpl) Setup() {
	s.Done = make(chan bool, 1)
	s.Command = make(chan opc.OpcStructure, 1)
	s.Response = make(chan error, 1)
	s.CommandMap = map[string]func(any) error{
		"configure nodes": s.handleConfigureNodes,
	}
}

func (s *TcpServerImpl) Start() error {
	s.Logger.Info("starting TCP server")
	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("failed to start TCP server on %s, reason: %v", addr, err))
		return err
	}
	s.Listener = &listener
	s.Logger.Info(fmt.Sprintf("started TCP server on %s", listener.Addr()))

	for {
		c, err := listener.Accept()
		if err != nil {
			select {
			case <-s.Done:
				return nil
			default:
				s.Logger.Error(fmt.Sprintf("error accepting client connection: %v", err))
				continue
			}
		}

		s.handleConnection(c)
	}
}

func (s *TcpServerImpl) GetCommandChannel() chan opc.OpcStructure {
	return s.Command
}
func (s *TcpServerImpl) GetResponseChannel() chan error {
	return s.Response
}

func (s *TcpServerImpl) Stop() error {
	if s.Listener == nil {
		return fmt.Errorf("server was never started")
	}
	s.Logger.Info("stopping TCP server")
	s.Done <- true
	if err := (*s.Listener).Close(); err != nil {
		s.Logger.Error(fmt.Sprintf("failed to stop TCP server, reason: %v", err))
		return err
	}
	s.Logger.Info("stopped TCP server")
	return nil
}

func (s *TcpServerImpl) handleConnection(c net.Conn) error {
	s.Logger.Debug(fmt.Sprintf("new connection from %s", c.RemoteAddr()))
	defer c.Close()

	jsonStr, err := s.readMessage(c)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("error whilst reading message from %s, reason: %v", c.RemoteAddr(), err))
		return err
	}
	s.Logger.Debug(fmt.Sprintf("from %s received %s", c.RemoteAddr(), jsonStr))

	command, err := s.parseCommand(jsonStr)
	if err != nil {
		s.Logger.Error("error parsing client message")
		msg := "invalid message"
		res, _ := json.Marshal(serialization.Respose{
			Status: "failure",
			Reason: &msg,
		})
		c.Write(res)
		return err
	}

	if handler, found := s.CommandMap[strings.ToLower(command.Command)]; !found {
		s.Logger.Warn(fmt.Sprintf("unrecognized client command %s", command.Command))
		msg := "unrecognized client"
		res, _ := json.Marshal(serialization.Respose{
			Status: "failure",
			Reason: &msg,
		})
		c.Write(res)
		return nil
	} else {
		err = handler(command.Payload.ToDomain(s.Logger))
		var res serialization.Respose
		if err != nil {
			msg := err.Error()
			res.Status = "failure"
			res.Reason = &msg
		} else {
			res.Status = "success"
		}
		resJson, _ := json.Marshal(res)
		c.Write(resJson)
		return err
	}
}

func (s *TcpServerImpl) readMessage(c net.Conn) (string, error) {
	reader := bufio.NewReader(c)
	var buffer bytes.Buffer
	msg, err := reader.ReadBytes(0x0A)
	if err != nil && err != io.EOF {
		s.Logger.Error(fmt.Sprintf("error reading connection payload: %v", err))
		return "", err
	}

	buffer.Write(msg)
	return strings.Trim(strings.ReplaceAll(buffer.String(), "\n", ""), "\n\t "), nil
}

func (s *TcpServerImpl) parseCommand(m string) (*serialization.Command, error) {
	var command serialization.Command
	err := json.Unmarshal([]byte(m), &command)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("error unmarshaling JSON: %v", err))
		return nil, err
	}
	return &command, nil
}

func (s *TcpServerImpl) handleConfigureNodes(p any) (err error) {
	defer func() {
		if r := recover(); r != nil {
			s.Logger.Error("input is not OPC structure")
			err = fmt.Errorf("invalid input")
		}
	}()
	structure := p.(opc.OpcStructure)

	s.Command <- structure
	if err = <-s.Response; err != nil {
		s.Logger.Error(fmt.Sprintf("failed to apply new OPC node structure, reason: %v", err))
		return err
	} else {
		return nil
	}
}
