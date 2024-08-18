package impl

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/olte36/grpc-monorepo-example/genproto/api"
)

func Offer(ctx context.Context, client api.StockServiceClient, stocks []*api.Stock) error {
	stream, err := client.Offer(ctx)
	if err != nil {
		return err
	}
	for _, stock := range stocks {
		offer := api.OfferRequest{
			Stock:        stock,
			InitialPrice: uint32(rand.Intn(70)),
		}
		fmt.Printf("Offering %s for %.2f\n", offer.Stock.Ticker, float64(offer.InitialPrice)/100)
		err = stream.Send(&offer)
		if err != nil {
			return err
		}
		time.Sleep(1 * time.Second)
	}
	resp, err := stream.CloseAndRecv()
	if err != nil {
		return err
	}
	if len(resp.ListedStocks) == 0 {
		fmt.Println("No stocks have been listed")
	} else {
		var listedTickers []string
		for _, stock := range resp.ListedStocks {
			listedTickers = append(listedTickers, stock.Ticker)
		}
		fmt.Printf("The following stocks have been listed: %s\n", strings.Join(listedTickers, ", "))
	}
	if len(resp.RejectedStocks) != 0 {
		for _, stock := range resp.RejectedStocks {
			desc := ""
			if stock.Description != nil {
				desc = *stock.Description
			}
			fmt.Printf("%s has been rejected: %s\n", stock.Ticker, desc)
		}
	}
	return nil
}
