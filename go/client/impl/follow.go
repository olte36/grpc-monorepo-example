package impl

import (
	"context"
	"fmt"
	"time"

	"github.com/olte36/grpc-monorepo-example/genproto/api"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TrackPrice(ctx context.Context, client api.StockServiceClient, stock *api.Stock) (int, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	stream, err := client.Follow(ctx, &api.FollowRequest{
		Stock: &api.Stock{
			Ticker: stock.Ticker,
		},
		TrackInterval: durationpb.New(3 * time.Second),
	})
	if err != nil {
		return 0, err
	}

	errChan := make(chan error)
	priceChan := make(chan int)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				resp, err := stream.Recv()
				if err != nil {
					errChan <- err
					return
				}
				priceChan <- int(resp.Price)
				fmt.Printf("%s %.2f\n", resp.Stock.Ticker, float64(resp.Price)/100)
			}
		}
	}()

	sum := 0
	count := 0
	sec := 15
	timer := time.NewTimer(time.Duration(sec) * time.Second)
	stop := false
	fmt.Printf("Tracking the price of %s for %d sec\n", stock.Ticker, sec)
	for !stop {
		select {
		case <-ctx.Done():
			if ctx.Err() != nil {
				return 0, ctx.Err()
			}
			stop = true
		case <-timer.C:
			stop = true
		case price := <-priceChan:
			sum += price
			count++
		case err := <-errChan:
			return 0, err
		}
	}
	avg := 0
	if count != 0 {
		avg = sum / count
	}
	return avg, nil
}
