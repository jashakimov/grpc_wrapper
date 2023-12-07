package grpc_server

import (
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"net"
	"net/http"
	"time"
)

type Server interface {
	Start()
	GracefulStop()
	Stop()
	Serve() error
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	GetServiceInfo() map[string]grpc.ServiceInfo
	RegisterService(sd *grpc.ServiceDesc, ss interface{})
}

type grpcServer[T any] struct {
	cfg    *ServerConfig[T]
	lis    net.Listener
	server *grpc.Server
}

type ServerConfig[T any] struct {
	Addr            string
	PanicHandler    grpc_recovery.RecoveryHandlerFunc
	ServerOptions   []grpc.ServerOption
	GrpcServiceDesc *grpc.ServiceDesc
	GrpcServiceImpl T
	ShutdownTimeSec time.Duration
}

func New[T any](cfg *ServerConfig[T]) Server {
	s := new(grpcServer[T])
	s.cfg = cfg

	// set listener
	lis, err := net.Listen("tcp", s.cfg.Addr)
	if err != nil {
		log.Fatal(err)
	}
	s.lis = lis

	// set default PanicHandler if not exist
	if s.cfg.PanicHandler == nil {
		s.cfg.PanicHandler = func(p any) (err error) {
			return status.Errorf(codes.Internal, "%s", p)
		}
	}
	s.cfg.ServerOptions = append(s.cfg.ServerOptions, grpc.ChainUnaryInterceptor(
		grpc_recovery.UnaryServerInterceptor(grpc_recovery.WithRecoveryHandler(s.cfg.PanicHandler)),
	))

	// set server
	s.server = grpc.NewServer(s.cfg.ServerOptions...)

	// register grpc service
	s.server.RegisterService(s.cfg.GrpcServiceDesc, s.cfg.GrpcServiceImpl)

	return s
}

func (s *grpcServer[T]) Start() {
	log.Printf("gRPC server '%s' listening at %s\n", s.cfg.GrpcServiceDesc.ServiceName, s.lis.Addr())
	if err := s.Serve(); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (s *grpcServer[T]) GracefulStop() {
	s.server.GracefulStop()
	time.Sleep(s.cfg.ShutdownTimeSec)
}

func (s *grpcServer[T]) Stop() {
	s.server.Stop()
}

func (s *grpcServer[T]) Serve() error {
	return s.server.Serve(s.lis)
}

func (s *grpcServer[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.server.ServeHTTP(w, r)
}

func (s *grpcServer[T]) GetServiceInfo() map[string]grpc.ServiceInfo {
	return s.server.GetServiceInfo()
}

func (s *grpcServer[T]) RegisterService(sd *grpc.ServiceDesc, ss interface{}) {
	s.server.RegisterService(sd, ss)
}
