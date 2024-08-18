package impl

import (
	"context"
	"fmt"
	"time"

	"github.com/olte36/grpc-monorepo-example/genproto/api"
	"google.golang.org/protobuf/types/known/emptypb"
)

func ListStocks(ctx context.Context, client api.StockServiceClient) ([]*api.Stock, error) {
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
