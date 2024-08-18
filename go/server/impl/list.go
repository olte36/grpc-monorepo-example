package impl

import (
	"context"

	"github.com/olte36/grpc-monorepo-example/genproto/api"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *stockServer) List(context.Context, *emptypb.Empty) (*api.ListResponse, error) {
	log.Info().Msg("Giving the list of stocks")
	resp := api.ListResponse{}
	for _, stock := range s.stocks {
		resp.Stocks = append(resp.Stocks, &api.Stock{
			Ticker:      stock.ticker,
			Description: &stock.description,
		})
	}
	return &resp, nil
}
