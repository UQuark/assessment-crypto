package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/aiviaio/go-binance/v2"
	"log"
	"sync"
)

func getFirstFiveSymbols(cl *binance.Client) ([]binance.Symbol, error) {
	// Retrieve the first five trading pairs (symbols) through the endpoint api/v3/exchangeInfo
	ei, err := cl.NewExchangeInfoService().Do(context.Background())
	if err != nil {
		return nil, err
	}
	if len(ei.Symbols) < 5 {
		return nil, errors.New("response does not contain enough symbols")
	}
	firstFive := make([]binance.Symbol, 5)
	copy(firstFive, ei.Symbols[:5])
	return firstFive, nil
}

func fetchLastPrice(cl *binance.Client, symbol binance.Symbol, ch chan<- *binance.SymbolPrice) {
	// Obtain the latest price of each symbol
	prices, err := cl.NewListPricesService().Symbol(symbol.Symbol).Do(context.Background())
	if err != nil {
		log.Println(err)
		return
	}
	if len(prices) == 0 {
		log.Println("response does not contain prices")
		return
	}
	// Transmit the symbol and price
	ch <- prices[0]
}

func printSymbolPrices(ch <-chan *binance.SymbolPrice, wg *sync.WaitGroup) {
	// Print all symbol-price pairs
	for price := range ch {
		fmt.Printf("%s %s\n", price.Symbol, price.Price)
		wg.Done()
	}
}

func main() {
	cl := binance.NewClient("", "")
	symbols, err := getFirstFiveSymbols(cl)
	if err != nil {
		log.Fatal(err)
	}

	ch := make(chan *binance.SymbolPrice)
	wg := &sync.WaitGroup{}

	go printSymbolPrices(ch, wg)

	for _, s := range symbols {
		// Each symbol adds 1 to the WaitGroup
		wg.Add(1)
		// Create goroutines and pass each symbol as a parameter to a separate goroutine
		go fetchLastPrice(cl, s, ch)
	}
	// Each symbol-price pair print will subtract 1 from the WaitGroup
	wg.Wait()
	close(ch)
}
