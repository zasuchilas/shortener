package grpc_server

import (
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"github.com/zasuchilas/shortener/internal/app/api/grpc_api"
	"github.com/zasuchilas/shortener/internal/app/config"
	"github.com/zasuchilas/shortener/internal/app/logger"
	"github.com/zasuchilas/shortener/internal/app/service"
	desc "github.com/zasuchilas/shortener/pkg/shortener_v1"
)

type Server struct {
	server           *grpc.Server
	shortenerService service.ShortenerService
}

func NewServer(shortenerService service.ShortenerService) *Server {

	grpcServer := grpc.NewServer(
		grpc.Creds(insecure.NewCredentials()),
	)

	reflection.Register(grpcServer)

	desc.RegisterShortenerV1Server(grpcServer, grpc_api.NewImplementation(shortenerService))

	return &Server{
		server:           grpcServer,
		shortenerService: shortenerService,
	}
}

func (s *Server) Run() {
	logger.Log.Info("gRPC server is running", zap.String("address", config.GRPCServerAddress))

	list, err := net.Listen("tcp", config.GRPCServerAddress)
	if err != nil {
		logger.Log.Panic("gRPC listen")
	}

	err = s.server.Serve(list)
	if err != nil {
		logger.Log.Panic("gRPC serve")
	}
}
