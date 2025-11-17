package gapi

import (
	"context"
	"fmt"
	"net"

	config "172.21.5.249/air-trans/at-drone/internal/config"
	drone "172.21.5.249/air-trans/at-drone/internal/gapi/drone"
	droneTrack "172.21.5.249/air-trans/at-drone/internal/gapi/track_history"
	logger "172.21.5.249/air-trans/at-drone/internal/gapi/middleware"
	service "172.21.5.249/air-trans/at-drone/internal/service"
	pb "172.21.5.249/air-trans/at-drone/pkg/pb"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	Config      config.ServiceConfig
	MainService *service.MainService
}

func NewServer(cfg config.ServiceConfig, svc *service.MainService) *Server {
	return &Server{
		Config:      cfg,
		MainService: svc,
	}
}

func (s *Server) Start(errs chan error) {
	ctx := log.Logger.WithContext(context.Background())

	grpcLogger := grpc.UnaryInterceptor(logger.LoggerMiddleware)

	// embedded logger to grpc server
	grpcServer := grpc.NewServer(grpcLogger)

	droneHandler := drone.NewDroneHandler(s.MainService)
	droneTrackHandler := droneTrack.NewTrackHistoryHandler(s.MainService)
	pb.RegisterDroneServiceServer(grpcServer, droneHandler)
	pb.RegisterTrackHistoryServiceServer(grpcServer, droneTrackHandler)

	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", fmt.Sprintf("%v:%d", s.Config.GrpcConfig.GrpcHost, s.Config.GrpcConfig.GrpcPort))
	if err != nil {
		config.PrintFatalLog(ctx, err, "Cannot create grpc listener")

		errs <- err
	}

	config.PrintDebugLog(ctx, "Start GRPC server on: %s", listener.Addr().String())

	err = grpcServer.Serve(listener)
	if err != nil {
		config.PrintFatalLog(ctx, err, "Cannot start grpc server")

		errs <- err
	}
}
