package main

import (
	"context"
	"flag"
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

	var serverUrl string
	flag.StringVar(&serverUrl, "server", "localhost:8080", "The server to connect in format <host>:<port>")
	flag.Parse()

	conn, err := grpc.NewClient(serverUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		exitCode = 1
		return
	}
	defer conn.Close()

	// Create a grpc client
	client := api.NewStockServiceClient(conn)

	// get the list of stocks
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

	// offer some stocks with the duplicate to check it is rejected
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

	// track the price of the chosen stock
	avg, err := impl.TrackPrice(ctx, client, stock)
	if err != nil {
		exitCode = 1
		return
	}
	fmt.Printf("The average price is %.2f\n\n", float64(avg)/100)

	// do some trading
	impl.Trade(ctx, client, stock, avg)
}
