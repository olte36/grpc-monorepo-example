package impl

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"time"

	"github.com/olte36/grpc-monorepo-example/genproto/api"
)

func Trade(ctx context.Context, client api.StockServiceClient, stock *api.Stock, avgPrice int) error {
	// Use cancel context to close the stream when we're done
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Create bidirectional stream
	stream, err := client.Trade(ctx)
	if err != nil {
		return err
	}

	// Send some random number of orders
	numOrdersToPlace := rand.Intn(8) + 3
	fmt.Printf("Making %d orders\n", numOrdersToPlace)
	time.Sleep(1 * time.Second)
	for i := 0; i < numOrdersToPlace; i++ {
		// Prepare the order
		amount := rand.Intn(10) + 1
		// Buy or sell
		if rand.Intn(2) == 1 {
			amount = -amount
		}
		order := &api.TradeRequest{
			Stock: &api.Stock{
				Ticker: stock.Ticker,
			},
			Amount: int32(amount),
		}
		// Randomly limit the price to the avg
		if rand.Intn(2) == 1 {
			limitPrice := int32(avgPrice)
			order.LimitPrice = &limitPrice
		}

		handleOrderToPlace(order)

		// Send the order
		err := stream.Send(order)
		if err != nil {
			return err
		}

		// Wait some time for price change
		time.Sleep(2 * time.Second)
	}
	fmt.Println()

	numOrdersExecuted := 0
	errChan := make(chan error)
	allOrdersDone := make(chan bool)

	// Recieve executed orders in a goroutine because stream.Recv() is blocking
	// We can recieve not all orders because we set the limit price for some of them
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
				handleExecutedOrder(resp)
				numOrdersExecuted++
				if numOrdersExecuted == numOrdersToPlace {
					allOrdersDone <- true
					return
				}
			}
		}
	}()

	// Set the time we wait for orders to complete
	sec := 45
	fmt.Printf("Waiting %d sec for orders to complete\n", sec)
	timer := time.NewTimer(time.Duration(sec) * time.Second)

	// Waiting for the results
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-allOrdersDone:
		fmt.Println("All orders have been completed")
		return nil
	case <-timer.C:
		fmt.Printf("We've waited too long, cancel %d orders\n", numOrdersToPlace-numOrdersExecuted)
		return nil
	case err := <-errChan:
		if err == io.EOF {
			return nil
		}
		return err
	}
}

func handleOrderToPlace(order *api.TradeRequest) {
	amount := order.Amount
	op := "Buying"
	if amount < 0 {
		amount = -amount
		op = "Selling"
	}
	limitPriceStr := ""
	if order.LimitPrice != nil {
		limitPriceStr = fmt.Sprintf("limit price: %.2f", float64(*order.LimitPrice)/100)
	}
	fmt.Printf("%s %d x %s %s\n", op, amount, order.Stock.Ticker, limitPriceStr)
}

func handleExecutedOrder(order *api.TradeResponse) {
	amount := order.Amount
	op := "Bought"
	if amount < 0 {
		amount = -amount
		op = "Sold"
	}
	fmt.Printf("%s %d x %s by %.2f\n", op, amount, order.Stock.Ticker, float64(order.Price)/100)
}
