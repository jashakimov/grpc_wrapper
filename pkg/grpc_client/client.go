package grpc_client

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/resolver/manual"
	"log"
)

type Client[T any] interface {
	CloseConn() error
	Client() T
}

type grpcClient[T any] struct {
	conn   *grpc.ClientConn
	client T
}

type ClientConfig[T any] struct {
	Name        string
	Addrs       []string
	InitFunc    func(cc grpc.ClientConnInterface) T
	DialOptions []grpc.DialOption
}

func New[T any](cfg *ClientConfig[T]) Client[T] {
	gClient := new(grpcClient[T])
	// init resolver
	r := manual.NewBuilderWithScheme(cfg.Name)

	// init addresses
	addrs := make([]resolver.Address, len(cfg.Addrs))
	for i := range cfg.Addrs {
		addrs[i] = resolver.Address{Addr: cfg.Addrs[i]}
	}
	r.InitialState(resolver.State{
		Addresses: addrs,
	})

	// init dial options, cfg.DialOption must rewrite default options
	dialOptions := []grpc.DialOption{
		grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [{"round_robin":{}}]}`),
		grpc.WithResolvers(r),
	}
	dialOptions = append(dialOptions, cfg.DialOptions...)

	// init dial
	conn, err := grpc.Dial(r.Scheme()+":///", dialOptions...)
	if err != nil {
		log.Fatal(err)
	}
	gClient.conn = conn

	// init grpc gClient
	gClient.client = cfg.InitFunc(gClient.conn)

	return gClient
}

func (g *grpcClient[T]) CloseConn() error {
	g.conn.ResetConnectBackoff()
	return g.conn.Close()
}

func (g *grpcClient[T]) Client() T {
	return g.client
}
