package impl

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/olte36/grpc-monorepo-example/genproto/api"
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
	s.mu.Lock()
	trackedStock, ok := s.stocks[req.Stock.Ticker]
	s.mu.Unlock()
	if !ok {
		return status.Errorf(codes.NotFound, "the stock %s has not been found", req.Stock.Ticker)
	}
	timer := time.NewTicker(req.TrackInterval.AsDuration())
	for {
		select {
		case <-respStream.Context().Done():
			return nil
		case <-timer.C:
			s.mu.Lock()
			resp := api.GetPriceResponse{
				Stock: &api.Stock{
					Ticker: trackedStock.ticker,
				},
				Exp: exp,
			}
			s.mu.Unlock()
			err := respStream.Send(&resp)
			if err != nil {
				return status.Error(codes.Internal, err.Error())
			}
		}
	}
}

func (*stockServer) Trade(stream grpc.BidiStreamingServer[api.TradeRequest, api.TradeResponse]) error {
	return status.Errorf(codes.Unimplemented, "method Trade not implemented")
}

func (s *stockServer) changePrices() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, stock := range s.stocks {
		sample := rand.Intn(100)
		change := -1
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

		}
		stock.price *= (1 - change)
	}
}

func (s *stockServer) initStocks() {
	stocks := []*stock{
		{
			ticker:      "AAPL",
			description: "Apple Inc. Common Stock",
			price:       rand.Intn(40) + 80,
		},
		{
			ticker:      "PEP",
			description: "PepsiCo, Inc. Common Stock",
			price:       rand.Intn(40) + 80,
		},
		{
			ticker:      "JNJ",
			description: "Johnson & Johnson Common Stock",
			price:       rand.Intn(40) + 80,
		},
		{
			ticker:      "CSCO",
			description: "Cisco Systems, Inc. Common Stock",
			price:       rand.Intn(40) + 80,
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
		sleepTime := rand.Intn(6) + 5
		time.Sleep(time.Duration(sleepTime) * time.Second)
		stockServer.changePrices()
	}()
	api.RegisterStockServiceServer(s, &stockServer)
}
