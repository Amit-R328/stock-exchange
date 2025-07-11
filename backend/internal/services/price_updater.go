package services

import (
	"log"
	"math/rand/v2"
	"time"
)

type PriceUpdater struct {
	exchange *Exchange
	ticker   *time.Ticker
	done     chan bool
}

func NewPriceUpdater(exchange *Exchange, interval time.Duration) *PriceUpdater {
	return &PriceUpdater{
		exchange: exchange,
		ticker:   time.NewTicker(interval),
		done:     make(chan bool),
	}
}

func (pu *PriceUpdater) Start() {
	go func() {
		for {
			select {
			case <-pu.ticker.C:
				pu.updatePrices()
			case <-pu.done:
				return
			}
		}
	}()
}

func (pu *PriceUpdater) Stop() {
	pu.ticker.Stop()
	pu.done <- true
}

func (pu *PriceUpdater) updatePrices() {
	pu.exchange.mu.Lock()
	defer pu.exchange.mu.Unlock()

	for _, stock := range pu.exchange.Stocks {
		// Random change between -2% and +2%
		change := (rand.Float64() - 0.5) * 0.04
		oldPrice := stock.CurrentPrice
		stock.CurrentPrice = stock.CurrentPrice * (1 + change)

		// Round to 2 decimal places
		stock.CurrentPrice = float64(int(stock.CurrentPrice*100+0.5)) / 100

		if oldPrice != stock.CurrentPrice {
			log.Printf("Updated %s: $%.2f -> $%.2f", stock.Name, oldPrice, stock.CurrentPrice)
		}
	}
}
