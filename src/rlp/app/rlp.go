package app

import (
	"fmt"
	"log"
	"net"
	"rlp/internal/egress"
	"rlp/internal/ingress"
	"trafficcontroller/grpcconnector"

	v2 "plumbing/v2"

	"google.golang.org/grpc"
)

// RLP represents a running reverse log proxy. It connects to various gRPC
// servers to ingress data and opens a gRPC server to egress data.
type RLP struct {
	EgressAddr net.Addr
	egressPort int

	ingressAddrs    []string
	ingressDialOpts []grpc.DialOption

	receiver       *ingress.Receiver
	egressListener net.Listener
	egressServer   *grpc.Server
}

// RLPOption represents a function that can configure a remote log proxy.
type RLPOption func(c *RLP)

// WithEgressPort specifies the port used to bind the egress gRPC server.
func WithEgressPort(port int) RLPOption {
	return func(r *RLP) {
		r.egressPort = port
	}
}

// WithIngressAddrs specifies the addresses used to connect to ingress data.
func WithIngressAddrs(addrs []string) RLPOption {
	return func(r *RLP) {
		r.ingressAddrs = addrs
	}
}

// WithIngressDialOptions specifies the dial options used when connecting to
// the gRPC server to ingress data.
func WithIngressDialOptions(opts ...grpc.DialOption) RLPOption {
	return func(r *RLP) {
		r.ingressDialOpts = opts
	}
}

// Start creates and starts a remote log proxy.
func Start(opts ...RLPOption) *RLP {
	rlp := &RLP{
		ingressAddrs:    []string{"doppler.service.cf.internal"},
		ingressDialOpts: []grpc.DialOption{grpc.WithInsecure()},
	}
	for _, o := range opts {
		o(rlp)
	}
	rlp.setupIngress()
	rlp.setupEgress()
	go rlp.serveEgress()
	return rlp
}

func (r *RLP) setupIngress() {
	finder := ingress.NewFinder(r.ingressAddrs)
	pool := grpcconnector.NewPool(20, r.ingressDialOpts...)
	// TODO: Add real metrics
	batcher := &ingress.NullMetricBatcher{}
	connector := grpcconnector.New(1000, pool, finder, batcher)
	// TODO: user real converter
	converter := ingress.NullConverter{}
	r.receiver = ingress.NewReceiver(converter, connector)
}

func (r *RLP) setupEgress() {
	var err error
	r.egressListener, err = net.Listen("tcp", fmt.Sprintf(":%d", r.egressPort))
	if err != nil {
		log.Fatalf("failed to listen on port: %d: %s", r.egressPort, err)
	}
	r.EgressAddr = r.egressListener.Addr()
	r.egressServer = grpc.NewServer()
	v2.RegisterEgressServer(r.egressServer, egress.NewServer(r.receiver))
}

func (r *RLP) serveEgress() {
	if err := r.egressServer.Serve(r.egressListener); err != nil {
		log.Fatal("failed to serve: ", err)
	}
}
