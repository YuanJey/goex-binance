package main

import (
	"context"
	"fmt"
	"log"

	binance "github.com/YuanJey/goex-binance"
)

func main() {
	client, err := binance.New(binance.Config{})
	if err != nil {
		log.Fatal(err)
	}

	mark, err := client.Market.MarkPrice(context.Background(), "BTCUSDT")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s mark price: %s\n", mark.Symbol, mark.MarkPrice)
}
