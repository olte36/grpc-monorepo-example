package impl

import (
	"io"

	"github.com/olte36/grpc-monorepo-example/genproto/api"
	"google.golang.org/grpc"
)

func (s *stockServer) Offer(stream grpc.ClientStreamingServer[api.OfferRequest, api.OfferResponse]) error {
	var listed []*api.Stock
	var rejected []*api.Stock
	for {
		select {
		case <-stream.Context().Done():
			return nil
		default:
			req, err := stream.Recv()
			if err == io.EOF {
				return stream.SendAndClose(&api.OfferResponse{
					ListedStocks:   listed,
					RejectedStocks: rejected,
				})
			}
			if err != nil {
				return err
			}
			s.mu.Lock()
			_, ok := s.stocks[req.Stock.Ticker]
			if ok {
				desc := "The stock with such ticker is already listed"
				rejected = append(rejected, &api.Stock{
					Ticker:      req.Stock.Ticker,
					Description: &desc,
				})
			} else if req.InitialPrice == 0 {
				desc := "Cannot list the stock with zero price"
				rejected = append(rejected, &api.Stock{
					Ticker:      req.Stock.Ticker,
					Description: &desc,
				})
			} else {
				desc := ""
				if req.Stock.Description != nil {
					desc = *req.Stock.Description
				}
				s.stocks[req.Stock.Ticker] = &stock{
					ticker:      req.Stock.Ticker,
					description: desc,
					price:       int(req.InitialPrice),
				}
				listed = append(listed, req.Stock)
			}
			s.mu.Unlock()
		}
	}
}
