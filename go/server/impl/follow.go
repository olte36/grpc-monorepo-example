package impl

import (
	"time"

	"github.com/olte36/grpc-monorepo-example/genproto/api"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *stockServer) Follow(req *api.FollowRequest, respStream grpc.ServerStreamingServer[api.FollowResponse]) error {
	log.Info().Msgf("Streaming the price of %s", req.Stock.Ticker)
	s.mu.Lock()
	trackedStock, ok := s.stocks[req.Stock.Ticker]
	s.mu.Unlock()
	if !ok {
		log.Error().Msgf("We don't have the stock %s", req.Stock.Ticker)
		return status.Errorf(codes.NotFound, "the stock %s has not been found", req.Stock.Ticker)
	}
	ticker := time.NewTicker(req.TrackInterval.AsDuration())
	for {
		select {
		case <-respStream.Context().Done():
			log.Info().Msgf("Finished streaming the price of %s", req.Stock.Ticker)
			return respStream.Context().Err()
		case <-ticker.C:
			// suppose stocks cannot be deleted, so we don't sync them
			resp := api.FollowResponse{
				Stock: &api.Stock{
					Ticker: trackedStock.ticker,
				},
				Price: uint32(trackedStock.price),
			}
			err := respStream.Send(&resp)
			if err != nil {
				return err
			}
		}
	}
}
