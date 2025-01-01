package opcserver

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/url"
	"os"
	"reflect"
	"time"

	nodeengine "github.com/AndreiLacatos/opc-engine/node-engine"
	"github.com/AndreiLacatos/opc-engine/node-engine/models/opc"
	opcnode "github.com/AndreiLacatos/opc-engine/node-engine/models/opc/opc_node"
	waveformvalue "github.com/AndreiLacatos/opc-engine/node-engine/models/waveform/waveform_value"
	"github.com/awcullen/opcua/server"
	"github.com/awcullen/opcua/ua"
	"github.com/pkg/errors"
)

type opcServerImpl struct {
	Config    OpcServerConfig
	OpcServer *server.Server
}

func (s *opcServerImpl) Setup() error {
	if s.OpcServer != nil {
		return fmt.Errorf("server already set up")
	}

	log.Println("setting up OPC server")
	configJson, err := json.MarshalIndent(s.Config, "", "  ")
	if err != nil {
		log.Printf("failed to convert config to JSON: %v\n", err)
		return err
	}
	log.Printf("config: %s\n", string(configJson))

	if err = createServerCertificate(s.Config); err != nil {
		log.Printf("failed to create server certificate: %v\n", err)
		return err
	}

	endpointURL := fmt.Sprintf("opc.tcp://%s:%d", s.Config.ServerEndpointUrl, s.Config.Port)
	srv, err := server.New(
		ua.ApplicationDescription{
			ApplicationURI: fmt.Sprintf("urn:%s:%s", s.Config.ServerEndpointUrl, s.Config.ServerName),
			ProductURI:     "https://github.com/AndreiLacatos-works/opc-engine",
			ApplicationName: ua.LocalizedText{
				Text:   fmt.Sprintf("%s@%s", s.Config.ServerName, s.Config.ServerEndpointUrl),
				Locale: "en",
			},
			ApplicationType:     ua.ApplicationTypeServer,
			GatewayServerURI:    "",
			DiscoveryProfileURI: "",
			DiscoveryURLs:       []string{endpointURL},
		},
		"./pki/server.crt",
		"./pki/server.key",
		endpointURL,
		server.WithBuildInfo(
			ua.BuildInfo{
				ProductURI:       "https://github.com/AndreiLacatos-works/opc-engine",
				ManufacturerName: "AndreiLacatos",
				ProductName:      s.Config.ServerName,
				SoftwareVersion:  s.Config.BuildInfo.Version,
				BuildDate:        s.Config.BuildInfo.BuildDate,
			}),
		server.WithAnonymousIdentity(true),
		server.WithSecurityPolicyNone(true),
		server.WithInsecureSkipVerify(),
		server.WithServerDiagnostics(true),
	)
	if err != nil {
		fmt.Printf("error setting up server: %v\n", err)
		return err
	}

	s.OpcServer = srv
	return nil
}

func (s *opcServerImpl) Start() error {
	if s.OpcServer == nil {
		return fmt.Errorf("server never set up")
	}

	log.Printf("starter listening as %s on %s\n", s.Config.ServerName, s.OpcServer.EndpointURL())
	if err := s.OpcServer.ListenAndServe(); err != nil {
		return errors.Wrap(err, "error starting server")
	}

	return nil
}

func (s *opcServerImpl) Subscribe(c chan nodeengine.NodeValueChange) {
	for {
		p, ok := <-c
		if ok {
			s.updateNodeValue(p)
		} else {
			log.Println("event channel closed")
			return
		}
	}
}

func (s *opcServerImpl) Stop() error {
	if s.OpcServer == nil {
		return fmt.Errorf("server never set up")
	}
	log.Println("stopping opc server")
	return s.OpcServer.Close()
}

func (s *opcServerImpl) SetNodeStructure(o opc.OpcStructure) error {
	if s.OpcServer == nil {
		return fmt.Errorf("server never set up")
	}
	applicationObjects := ua.NewNodeIDNumeric(0, 85)
	r := opcnode.OpcContainerNode(o.Root)
	return addNodesRecursively(&r, applicationObjects, s.OpcServer.NamespaceManager(), s.OpcServer)
}

func (s *opcServerImpl) SetNodeValues() error {
	return nil
}

func addNodesRecursively(r opcnode.OpcStructureNode, p ua.NodeID, m *server.NamespaceManager, s *server.Server) error {
	n, err := makeNode(r, p, s)
	if err != nil {
		return err
	}
	m.AddNode(n)
	if t, ok := r.(*opcnode.OpcContainerNode); ok {
		for _, c := range t.Children {
			if err := addNodesRecursively(c, ua.NewNodeIDGUID(2, t.GetId()), m, s); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *opcServerImpl) updateNodeValue(c nodeengine.NodeValueChange) {
	log.Printf("change: %f on %s (%s)\n", c.NewValue.GetValue(), c.Node.Label, c.Node.Id)
	m := s.OpcServer.NamespaceManager()
	if node, ok := m.FindVariable(ua.NewNodeIDGUID(2, c.Node.Id)); !ok {
		log.Printf("node %s (%s) not found\n", c.Node.Label, c.Node.Id)
	} else {
		var v ua.Variant
		switch n := c.NewValue.(type) {
		case *waveformvalue.Transition:
			v = !node.Value().Value.(bool)
		case *waveformvalue.DoubleValue:
			v = n.GetValue()
		default:
			log.Printf("value type %T not recognized, skipping\n", reflect.TypeOf(c.NewValue))
			return
		}
		node.SetValue(ua.NewDataValue(v, ua.Good, time.Now(), 0, time.Now(), 0))
	}
}

func createServerCertificate(c OpcServerConfig) error {
	// do nothing if certificate already exists
	if _, err := os.Stat("./pki"); !os.IsNotExist(err) {
		return nil
	}

	// make a directory called pki for certificate files
	if err := os.MkdirAll("./pki", os.ModeDir|0755); err != nil {
		return err
	}

	// create a server cert in ./pki/server.crt
	if err := createNewCertificate(c, "./pki/server.crt", "./pki/server.key"); err != nil {
		return err
	}

	return nil
}

func createNewCertificate(c OpcServerConfig, certFile, keyFile string) error {
	// create a keypair
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return ua.BadCertificateInvalid
	}

	// get local hostname
	host, _ := os.Hostname()

	// create a certificate.
	applicationURI, _ := url.Parse(fmt.Sprintf("urn:%s:%s", host, c.ServerName))
	serialNumber, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	subjectKeyHash := sha1.New()
	subjectKeyHash.Write(key.PublicKey.N.Bytes())
	subjectKeyId := subjectKeyHash.Sum(nil)
	oidDC := asn1.ObjectIdentifier([]int{0, 9, 2342, 19200300, 100, 1, 25})

	template := x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               pkix.Name{CommonName: c.ServerName, ExtraNames: []pkix.AttributeTypeAndValue{{Type: oidDC, Value: host}}},
		SubjectKeyId:          subjectKeyId,
		AuthorityKeyId:        subjectKeyId,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageContentCommitment | x509.KeyUsageKeyEncipherment | x509.KeyUsageDataEncipherment | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{host, "localhost"},
		IPAddresses:           []net.IP{net.ParseIP(c.ServerEndpointUrl), []byte{127, 0, 0, 1}},
		URIs:                  []*url.URL{applicationURI},
	}

	rawcrt, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return ua.BadCertificateInvalid
	}

	if f, err := os.Create(certFile); err == nil {
		block := &pem.Block{Type: "CERTIFICATE", Bytes: rawcrt}
		if err := pem.Encode(f, block); err != nil {
			f.Close()
			return err
		}
		f.Close()
	} else {
		return err
	}

	if f, err := os.Create(keyFile); err == nil {
		block := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}
		if err := pem.Encode(f, block); err != nil {
			f.Close()
			return err
		}
		f.Close()
	} else {
		return err
	}

	return nil
}
