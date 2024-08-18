package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"

	"github.com/olte36/grpc-monorepo-example/client/impl"
	"github.com/olte36/grpc-monorepo-example/genproto/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer cancel()

	conn, err := grpc.NewClient("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		exitCode = 1
		return
	}
	defer conn.Close()

	// Create a grpc client
	client := api.NewStockServiceClient(conn)

	stocks, err := impl.ListStocks(ctx, client)
	if err != nil {
		exitCode = 1
		return
	}
	fmt.Println()

	// choose a random stock
	stock := stocks[rand.Intn(len(stocks))]
	fmt.Printf("We've chosen %s\n", stock.Ticker)
	fmt.Println()

	stocksToOffer := []*api.Stock{
		{
			Ticker: "TSM",
		},
		stock,
		{
			Ticker: "NVO",
		},
	}
	err = impl.Offer(ctx, client, stocksToOffer)
	if err != nil {
		exitCode = 1
		return
	}
	fmt.Println()

	avg, err := impl.TrackPrice(ctx, client, stock)
	if err != nil {
		exitCode = 1
		return
	}
	fmt.Printf("The average price is %.2f\n\n", float64(avg)/100)

	impl.Trade(ctx, client, stock, avg)
}
