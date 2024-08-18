package impl

import (
	"fmt"
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

func RegisterStockServer(s *grpc.Server) {
	stockServer := stockServer{}
	stockServer.initStocks()
	// change price every 3-5 sec
	go func() {
		for {
			if zerolog.GlobalLevel() == zerolog.DebugLevel {
				stockServer.logStocks()
			}
			sleepTime := rand.Intn(3) + 3
			time.Sleep(time.Duration(sleepTime) * time.Second)
			stockServer.changePrices()
		}
	}()
	api.RegisterStockServiceServer(s, &stockServer)
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
		if sample >= 50 { // 50% chance the price will change by 2%-10%
			change *= rand.Intn(10) + 2
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

func (s *stockServer) logStocks() {
	var printedStocks []string
	s.mu.Lock()
	for _, v := range s.stocks {
		price := float64(v.price) / math.Pow(10, float64(exp))
		printedStocks = append(printedStocks, fmt.Sprintf("%s %.2f", v.ticker, price))
	}
	s.mu.Unlock()
	slices.Sort(printedStocks)
	log.Debug().Msgf("Current stocks: %s", strings.Join(printedStocks, ", "))
}

func randPrice() int {
	return (rand.Intn(40) + 80) * int(math.Pow(10, 2))
}
