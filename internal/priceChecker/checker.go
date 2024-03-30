package priceChecker

import (
	"context"
	"fmt"
	"github.com/aiviaio/go-binance/v2"
	"log"
	"sync"
)

type Client struct {
	symbols         []string
	client          *binance.Client
	exchangeService *binance.ExchangeInfoService
	tickerService   *binance.ListSymbolTickerService
}

func NewClient(key, secret string) *Client {
	binanceClient := binance.NewClient(key, secret)
	c := &Client{
		client:          binanceClient,
		exchangeService: binanceClient.NewExchangeInfoService(),
		tickerService:   binanceClient.NewListSymbolTickerService(),
	}
	if err := c.loadDefaultSymbols(); err != nil {
		log.Fatal(err)
	}
	return c
}

func (c *Client) CheckPrices(ctx context.Context) error {
	resultCh := make(chan map[string]string, len(c.symbols))
	errCh := make(chan error, len(c.symbols))
	wg := &sync.WaitGroup{}

	for _, symbol := range c.symbols {
		wg.Add(1)
		go func(ctx context.Context, symbol string) {
			defer wg.Done()
			tickerResult, err := c.tickerService.Symbol(symbol).Do(ctx)
			if err != nil {
				errCh <- err
				return
			}
			if len(tickerResult) < 1 {
				errCh <- fmt.Errorf("no ticker data for %s", symbol)
				return
			}
			resultCh <- map[string]string{symbol: tickerResult[0].LastPrice}
		}(ctx, symbol)
	}

	go func() {
		wg.Wait()
		close(errCh)
		close(resultCh)
	}()

	for {
		select {
		case <-ctx.Done():
			log.Println("Task cancelled. Exiting...")
			return nil
		case result := <-resultCh:
			for r := range result {
				log.Printf("%s: %s", r, result[r])
			}
		case err, ok := <-errCh:
			if !ok {
				return nil
			}
			return err
		}
	}
}

func (c *Client) loadDefaultSymbols() error {
	exResult, err := c.exchangeService.Do(context.Background())
	if err != nil {
		return err
	}
	if len(exResult.Symbols) < 5 {
		return fmt.Errorf("expected at least 5 symbols, got %d", len(exResult.Symbols))
	}
	for _, s := range exResult.Symbols[:5] {
		c.symbols = append(c.symbols, s.Symbol)
	}
	return nil
}
