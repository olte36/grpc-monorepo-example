package impl

import (
	"context"
	"fmt"
	"io"

	"github.com/olte36/grpc-monorepo-example/genproto/api"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

func (s *stockServer) Trade(stream grpc.BidiStreamingServer[api.TradeRequest, api.TradeResponse]) error {
	log.Info().Msg("Starting trading")
	ctx := stream.Context()
	requestedOrders := make(chan *api.TradeRequest, 1)
	errChan := make(chan error)
	// execute orders in background
	go s.executeOrders(ctx, stream, requestedOrders, errChan)
	go s.requestOrders(ctx, stream, requestedOrders, errChan)

	select {
	case <-ctx.Done():
		log.Info().Msg("Finishing trading")
		return ctx.Err()
	case err := <-errChan:
		// if a client closed the streams, left orders are canceled
		if err == io.EOF {
			return nil
		}
		// in case of error, left orders are canceled
		if err != nil {
			log.Err(err).Msg("Stopping trading because of an error")
			return err
		}
		return nil
	}
}

func (s *stockServer) requestOrders(
	ctx context.Context,
	stream grpc.BidiStreamingServer[api.TradeRequest, api.TradeResponse],
	requestedOrders chan *api.TradeRequest,
	errChan chan error) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			req, err := stream.Recv()
			if err != nil {
				errChan <- fmt.Errorf("unable to recieve an order: %w", err)
				return
			}
			log.Debug().Msgf("Recieved the order %v", req)
			requestedOrders <- req
		}
	}
}

func (s *stockServer) executeOrders(
	ctx context.Context,
	stream grpc.BidiStreamingServer[api.TradeRequest, api.TradeResponse],
	requestedOrders chan *api.TradeRequest,
	errChan chan error) {
	ordersMap := map[string][]*api.TradeRequest{}
	for {
		select {
		case <-ctx.Done():
			return
		case order := <-requestedOrders:
			orders, ok := ordersMap[order.Stock.Ticker]
			if !ok {
				orders = []*api.TradeRequest{}
			}
			ordersMap[order.Stock.Ticker] = append(orders, order)
		default:
			var invalidOrders []*api.TradeRequest
			var readyOrders []*api.TradeResponse
			for ticker, orders := range ordersMap {
				s.mu.Lock()
				stock, ok := s.stocks[ticker]
				// ignore invalid orders
				if !ok {
					invalidOrders = append(invalidOrders, orders...)
					continue
				}
				var leftOrders []*api.TradeRequest
				for _, order := range orders {
					readyOrder := api.TradeResponse{
						Stock: &api.Stock{
							Ticker: stock.ticker,
						},
						Amount: order.Amount,
						Price:  uint32(stock.price),
					}
					// if no limit price, execute immediatly
					if order.LimitPrice == nil {
						readyOrders = append(readyOrders, &readyOrder)
					} else if order.Amount > 0 { // buy
						// if the current price <= the limit price, buy it, otherwise keep for the future
						if stock.price <= int(*order.LimitPrice) {
							readyOrders = append(readyOrders, &readyOrder)
						} else {
							leftOrders = append(leftOrders, order)
						}
					} else if order.Amount < 0 { // sell
						// if the current price >= the limit price, sell it, otherwise keep for the future
						if stock.price >= int(*order.LimitPrice) {
							readyOrders = append(readyOrders, &readyOrder)
						} else {
							leftOrders = append(leftOrders, order)
						}
					} else { // orders with zero amount are ignored
						invalidOrders = append(invalidOrders, order)
					}
				}
				s.mu.Unlock()
				ordersMap[ticker] = leftOrders
			}
			for _, readyOrder := range readyOrders {
				err := stream.Send(readyOrder)
				if err != nil {
					errChan <- fmt.Errorf("unable to send the order: %w", err)
					return
				}
			}
			if len(invalidOrders) > 0 {
				log.Warn().Msgf("Invalid orders: %v have been ignored", invalidOrders)
			}
		}
	}
}
