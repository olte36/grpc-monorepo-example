package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/olte36/grpc-monorepo-example/server/impl"
	"google.golang.org/grpc"
)

func main() {
	exitCode := 0
	var err error
	defer func() {
		if err != nil {
			fmt.Println(err)
		}
		os.Exit(exitCode)
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer cancel()

	var port uint
	var logLevelStr string
	flag.UintVar(&port, "port", 8080, "")
	flag.StringVar(&logLevelStr, "loglevel", "info", "")

	logLevel, err := zerolog.ParseLevel(logLevelStr)
	if err != nil {
		exitCode = 1
		return
	}
	zerolog.SetGlobalLevel(logLevel)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		exitCode = 1
		return
	}

	grpcServer := grpc.NewServer()
	impl.RegisterStockServer(grpcServer)

	go func() {
		log.Info().Msgf("Listening on port %d", port)
		err := grpcServer.Serve(lis)
		if err != nil {
			exitCode = 1
			cancel()
		}
	}()

	<-ctx.Done()
	log.Info().Msg("Stopping")
	grpcServer.Stop()
}
