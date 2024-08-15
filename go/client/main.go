package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/olte36/grpc-monorepo-example/genproto/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/durationpb"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer cancel()

	conn, err := grpc.NewClient("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Println(err)
		return
	}

	defer conn.Close()

	client := api.NewStockServiceClient(conn)
	stream, err := client.GetPrice(ctx, &api.GetPriceRequest{
		Stock: &api.Stock{
			Ticker: "AAPL",
		},
		TrackInterval: durationpb.New(3 * time.Second),
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			resp, err := stream.Recv()
			if err != nil {
				cancel()
			}
			price := float64(resp.Price) / math.Pow(10, float64(resp.Exp))
			fmt.Printf("%s %f\n", resp.Stock.Ticker, price)
		}
	}
}
