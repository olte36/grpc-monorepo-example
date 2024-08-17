package impl

import (
	"context"
	"fmt"
	"io"
	"math"
	"math/rand"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/olte36/grpc-monorepo-example/genproto/api"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const exp = 2

type stockServer struct {
	api.UnimplementedStockServiceServer
	stocks map[string]*stock
	mu     sync.Mutex
}

type stock struct {
	ticker      string
	description string
	price       int
}

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

func (s *stockServer) GetPrice(req *api.GetPriceRequest, respStream grpc.ServerStreamingServer[api.GetPriceResponse]) error {
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
			resp := api.GetPriceResponse{
				Stock: &api.Stock{
					Ticker: trackedStock.ticker,
				},
				Price: int32(trackedStock.price),
			}
			err := respStream.Send(&resp)
			if err != nil {
				return err
			}
		}
	}
}

func (s *stockServer) Trade(stream grpc.BidiStreamingServer[api.TradeRequest, api.TradeResponse]) error {
	log.Info().Msg("Starting trading")
	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()
	requestedOrders := make(chan *api.TradeRequest, 1)
	executedOrders := make(chan *api.TradeResponse, 1)
	// execute orders in background
	go s.executeOrders(ctx, requestedOrders, executedOrders)
	for {
		select {
		case <-stream.Context().Done():
			log.Info().Msg("Finishing trading")
			return stream.Context().Err()
		// send executed orders
		case executedOrder := <-executedOrders:
			err := stream.Send(executedOrder)
			// in case of error, left orders are canceled
			if err != nil {
				log.Err(err).Msgf("Cannot send executed order %v to the client, the left orders will be canceled", executedOrder)
				return err
			}
			log.Debug().Msgf("Sent the order %v", executedOrder)
		// accept new orders
		default:
			req, err := stream.Recv()
			// if a client closed the streams, left orders are canceled
			if err == io.EOF {
				return nil
			}
			// in case of error, left orders are canceled
			if err != nil {
				log.Err(err).Msgf("Error while receving a client's order, the left orders will be canceled")
				return err
			}
			log.Debug().Msgf("Recieved the order %v", req)
			requestedOrders <- req
		}
	}
}

func (s *stockServer) executeOrders(ctx context.Context, requestedOrders chan *api.TradeRequest, executedOrders chan *api.TradeResponse) {
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
						Price:  int32(stock.price),
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
				executedOrders <- readyOrder
			}
			if len(invalidOrders) > 0 {
				log.Warn().Msgf("Invalid orders: %v have been ignored", invalidOrders)
			}
		}
	}
}

func (s *stockServer) changePrices() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, stock := range s.stocks {
		sample := rand.Intn(100)
		change := -1
		// 50% chance the price will go up
		if rand.Intn(2) == 1 {
			change = 1
		}
		if sample >= 50 { // 50% chance the price will change by 1%-10%
			change *= rand.Intn(10) + 1
		} else if sample >= 30 { // 20% chance the price will change by 11-30%
			change *= rand.Intn(20) + 11
		} else if sample >= 20 { // 10% chance the price will change by 31-60%
			change *= rand.Intn(30) + 31
		} else if sample >= 18 { // 2% chance the price will change by 61-90%
			change *= rand.Intn(30) + 61
		} else { // 18% chance the price will not change
			change = 0
		}
		stock.price = stock.price * (100 - change) / 100
	}
}

func (s *stockServer) initStocks() {
	stocks := []*stock{
		{
			ticker:      "AAPL",
			description: "Apple Inc. Common Stock",
			price:       randPrice(),
		},
		{
			ticker:      "PEP",
			description: "PepsiCo, Inc. Common Stock",
			price:       randPrice(),
		},
		{
			ticker:      "JNJ",
			description: "Johnson & Johnson Common Stock",
			price:       randPrice(),
		},
		{
			ticker:      "CSCO",
			description: "Cisco Systems, Inc. Common Stock",
			price:       randPrice(),
		},
	}
	s.mu.Lock()
	s.stocks = map[string]*stock{}
	for _, stock := range stocks {
		s.stocks[stock.ticker] = stock
	}
	s.mu.Unlock()
}

func RegisterStockServer(s *grpc.Server) {
	stockServer := stockServer{}
	stockServer.initStocks()
	// change price every 5-7 sec
	go func() {
		for {
			if zerolog.GlobalLevel() == zerolog.DebugLevel {
				var printedStocks []string
				stockServer.mu.Lock()
				for _, v := range stockServer.stocks {
					price := float64(v.price) / math.Pow(10, float64(exp))
					printedStocks = append(printedStocks, fmt.Sprintf("%s %.2f", v.ticker, price))
				}
				stockServer.mu.Unlock()
				slices.Sort(printedStocks)
				log.Debug().Msgf("Current stocks: %s", strings.Join(printedStocks, ", "))
			}
			sleepTime := rand.Intn(6) + 5
			time.Sleep(time.Duration(sleepTime) * time.Second)
			stockServer.changePrices()
		}
	}()
	api.RegisterStockServiceServer(s, &stockServer)
}

func randPrice() int {
	return (rand.Intn(40) + 80) * int(math.Pow(10, 2))
}
