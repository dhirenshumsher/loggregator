package component_test

import (
	"net"
	"plumbing"
	v2 "plumbing/v2"

	"testservers"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type DopplerIngestorServer interface {
	plumbing.DopplerIngestorServer
}

type Server struct {
	port     int
	server   *grpc.Server
	listener net.Listener
	V1       *mockDopplerIngestorServerV1
	V2       *mockDopplerIngressServerV2
}

func NewServer() (*Server, error) {
	tlsConfig, err := plumbing.NewMutualTLSConfig(
		testservers.DopplerCertPath(),
		testservers.DopplerKeyPath(),
		testservers.CACertPath(),
		"",
	)
	if err != nil {
		return nil, err
	}
	transportCreds := credentials.NewTLS(tlsConfig)
	mockDopplerV1 := newMockDopplerIngestorServerV1()
	mockDopplerV2 := newMockDopplerIngressServerV2()

	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, err
	}

	s := grpc.NewServer(grpc.Creds(transportCreds))
	plumbing.RegisterDopplerIngestorServer(s, mockDopplerV1)
	v2.RegisterDopplerIngressServer(s, mockDopplerV2)

	go s.Serve(lis)

	return &Server{
		port:     lis.Addr().(*net.TCPAddr).Port,
		server:   s,
		listener: lis,
		V1:       mockDopplerV1,
		V2:       mockDopplerV2,
	}, nil
}

func (s *Server) URI() string {
	return s.listener.Addr().String()
}

func (s *Server) Port() int {
	return s.port
}

func (s *Server) Stop() error {
	err := s.listener.Close()
	s.server.Stop()
	return err
}
