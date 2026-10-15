package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	binance "github.com/YuanJey/goex-binance"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	client, err := binance.New(binance.Config{})
	if err != nil {
		log.Fatal(err)
	}
	sub, err := client.Stream.Market(ctx, "btcusdt@bookTicker")
	if err != nil {
		log.Fatal(err)
	}
	defer sub.Close()
	for {
		select {
		case raw, ok := <-sub.Events():
			if !ok {
				return
			}
			if event, ok := raw.(*binance.BookTickerEvent); ok {
				fmt.Println(event.BidPrice, event.AskPrice)
			}
		case err := <-sub.Errors():
			log.Println(err)
		case <-ctx.Done():
			return
		}
	}
}
