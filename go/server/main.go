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
	defer func() {
		os.Exit(exitCode)
	}()

	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer cancel()

	var port uint
	flag.UintVar(&port, "port", 8080, "")

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Err(err).Msgf("cannot listen on the port %d", port)
		exitCode = 1
		return
	}

	grpcServer := grpc.NewServer()
	impl.RegisterStockServer(grpcServer)

	go func() {
		log.Info().Msgf("Listening on port %d", port)
		err := grpcServer.Serve(lis)
		if err != nil {
			log.Err(err).Msg("cannot serve")
			exitCode = 1
			cancel()
		}
	}()

	<-ctx.Done()
	log.Info().Msg("Stopping")
	grpcServer.Stop()
}
