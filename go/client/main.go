package main

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/olte36/grpc-monorepo-example/genproto/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
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

	stocks, err := listStocks(ctx, client)
	if err != nil {
		exitCode = 1
		return
	}
	fmt.Println()

	// choose a random stock
	stock := stocks[rand.Intn(len(stocks))]
	fmt.Printf("We've chosen %s\n", stock.Ticker)

	trackingPriceCtx, stopTrackingPrice := context.WithCancel(ctx)
	avg, err := trackPrice(trackingPriceCtx, client, stock)
	stopTrackingPrice()
	if err != nil {
		exitCode = 1
		return
	}
	fmt.Printf("The average price is %.2f\n\n", float64(avg)/100)

	tardeCtx, stopTrade := context.WithTimeout(ctx, 30*time.Second)
	defer stopTrade()
	trade(tardeCtx, client, stock)
}

func listStocks(ctx context.Context, client api.StockServiceClient) ([]*api.Stock, error) {
	stocksList, err := client.List(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}
	fmt.Println("Available stocks:")
	for _, stock := range stocksList.Stocks {
		time.Sleep(1 * time.Second)
		desc := ""
		if stock.Description != nil {
			desc = *stock.Description
		}
		fmt.Printf("%s - %s\n", stock.Ticker, desc)
	}
	return stocksList.Stocks, nil
}

func trackPrice(ctx context.Context, client api.StockServiceClient, stock *api.Stock) (int, error) {
	priceStream, err := client.GetPrice(ctx, &api.GetPriceRequest{
		Stock: &api.Stock{
			Ticker: stock.Ticker,
		},
		TrackInterval: durationpb.New(3 * time.Second),
	})
	if err != nil {
		return 0, err
	}

	sec := 15
	timer := time.NewTimer(time.Duration(sec) * time.Second)
	fmt.Printf("Tracking the price of %s for %d sec\n", stock.Ticker, sec)
	sum := 0
	count := 0
	for {
		select {
		case <-ctx.Done():
			return sum / count, nil
		case <-timer.C:
			return sum / count, nil
		default:
			resp, err := priceStream.Recv()
			if err == io.EOF {
				return sum / count, nil
			}
			if err != nil {
				return 0, err
			}
			sum += int(resp.Price)
			count++
			fmt.Printf("%s %.2f\n", resp.Stock.Ticker, float64(resp.Price)/100)
		}
	}
}

func trade(ctx context.Context, client api.StockServiceClient, stock *api.Stock) error {
	stream, err := client.Trade(ctx)
	if err != nil {
		return err
	}
	for i := 0; i < rand.Intn(5)+1; i++ {
		order := &api.TradeRequest{
			Stock: &api.Stock{
				Ticker: stock.Ticker,
			},
			Amount: int32(rand.Intn(10) - 5),
		}
		err := stream.Send(order)
		if err != nil {
			return err
		}
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			resp, err := stream.Recv()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return err
			}
			fmt.Printf("%v\n", resp)
		}
	}
}
