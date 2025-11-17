package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	config "172.21.5.249/air-trans/at-drone/internal/config"
	gapi "172.21.5.249/air-trans/at-drone/internal/gapi"
	gclient "172.21.5.249/air-trans/at-drone/internal/gapi/client"
	hapi "172.21.5.249/air-trans/at-drone/internal/hapi"
	router "172.21.5.249/air-trans/at-drone/internal/hapi/router"
	service "172.21.5.249/air-trans/at-drone/internal/service"

	"github.com/nats-io/nats.go"
	"github.com/qiniu/qmgo"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const probeFlag string = "probe"

var serverCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the server",
	Long:  "Starts server",
	Run: func(cmd *cobra.Command, args []string) {
		runServer(args)
	},
}

func init() {
	serverCmd.Flags().BoolP(probeFlag, "p", false, "Probe readiness before startup.")

	rootCmd.AddCommand(serverCmd)
}

func runServer(args []string) {
	baseCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx := log.Logger.WithContext(baseCtx)

	/**
	* Load config file
	 */
	cfgFile := "."

	if len(args) != 0 {
		cfgFile = args[0]

		config.PrintDebugLog(ctx, "Use config file by argument: %+v", cfgFile)
	}

	config.PrintDebugLog(ctx, "Load config file: %s", cfgFile)

	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		config.PrintFatalLog(ctx, err, "Failed to load config file: %s", cfgFile)

		os.Exit(1)
	}

	config.PrintDebugLog(ctx, "Config file content: %+v", cfg)

	/**
	* Setting logger
	 */
	if cfg.OtherConfig.Environment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05 02-01-2006"})
	}

	// /**
	// * Setting tracer
	//  */
	// tp, err := initTracer(cfg.OtherConfig)
	// if err != nil {
	// 	config.PrintErrorLog(ctx, err, "Failed to init tracer")
	// }
	// defer func() {
	// 	if err := tp.Shutdown(context.Background()); err != nil {
	// 		config.PrintErrorLog(ctx, err, "Failed to shutdown tracer provider")
	// 	}
	// }()

	/**
	* Start mongoDB client connection
	 */
	addr := fmt.Sprintf("mongodb://%s:%d/?replicaset=%s", cfg.DbConfig.DBHost, cfg.DbConfig.DBPort, cfg.DbConfig.DBReplica)
	qmgoClient, err := qmgo.NewClient(ctx, &qmgo.Config{Uri: addr})
	if err != nil {
		config.PrintFatalLog(ctx, err, "Failed to connect to MongoDB: %s", addr)

		os.Exit(1)
	} else {
		config.PrintDebugLog(ctx, "Connected to connect to MongoDB: %s", addr)
	}

	// /**
	// * Start RabbitMQ client connection
	//  */
	// uri := fmt.Sprintf("%s://%s:%s@%s:%d/%s",
	// 	cfg.RabbitmqConfig.Schema,
	// 	cfg.RabbitmqConfig.Username,
	// 	cfg.RabbitmqConfig.Password,
	// 	cfg.RabbitmqConfig.Host,
	// 	cfg.RabbitmqConfig.Port,
	// 	cfg.RabbitmqConfig.Vhost)
	// conn := mq.Connection(uri)
	// publisher := publisher.NewEventPublisher(conn, cfg.RabbitmqConfig.EventExchange)

	/**
	* Start NATS client connection
	 */
	natsClient, err := nats.Connect(cfg.NATSConfig.Server)
	if err != nil {
		config.PrintFatalLog(ctx, err, "Failed to connect NATs server: %s", cfg.NATSConfig.Server)
	} else {
		config.PrintDebugLog(ctx, "Connected to connect NATs server: %s", cfg.NATSConfig.Server)
	}
	defer natsClient.Drain()

	/**
	* Start GRPC client connection
	 */
	grpcClient := gclient.New(cfg.GrpcConfig.GrpcChannels)

	svc := service.New(qmgoClient, cfg, grpcClient, natsClient)
	svc.StartFlightContainmentMonitor(ctx, time.Second)

	errs := make(chan error, 2)

	/**
	* Start GRPC server
	 */
	config.PrintDebugLog(ctx, "Starting HTTP server...")

	httpServer := hapi.NewServer(svc, cfg)
	httpServer.InitI18n()
	router.Init(httpServer)
	go httpServer.Start(errs)

	/**
	* Start HTTP server
	 */
	config.PrintDebugLog(ctx, "Starting GRPC server...")

	grpcServer := gapi.NewServer(cfg, svc)
	go grpcServer.Start(errs)

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT)

		errs <- fmt.Errorf("%s", <-c)
	}()
	// time.Sleep(3 * time.Second)
	// config.PrintDebugLog(ctx, "Starting schedu...")
	// svc.StartScheduler(svc)

	err = <-errs

	config.PrintFatalLog(ctx, err, "Services terminate")
}
