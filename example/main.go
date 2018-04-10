package main

import (
	coin "github.com/cjongseok/fetch-coinmarketcap"
	"fmt"
	"time"
	"github.com/cjongseok/slog"
)

func main() {
	fmt.Println("Start")
	coin.Start()
	fmt.Printf("%s: Fetched? %v\n", time.Now(), coin.Fetched())
	coin.WaitForFetching()
	fmt.Printf("%s: Fetched? %v\n", time.Now(), coin.Fetched())
	fmt.Println("Coins:", slog.Stringify(coin.All()))
	fmt.Println("Market:", slog.Stringify(coin.TotalMarket()))
	fmt.Println("Size:", coin.Size())
	fmt.Println("Get BAT:", slog.Stringify(coin.Get("BAT")))
	//c := coin.Get("BAT")
	//str := strconv.FormatFloat(c.PriceBtc, 'f', -1, 64)
	//fmt.Println("BAT prince in BTC:", c.PriceBtc)
	//fmt.Printf("BAT prince in BTC: %v\n", c.PriceBtc)
	//fmt.Printf("BAT prince in BTC: %s\n", str)
	//fmt.Printf("BAT prince in BTC: %.10f\n", c.PriceBtc)
	fmt.Println("Now:", time.Now())
	fmt.Println("Next coin fetching time:", coin.NextCoinFetchTime())
	fmt.Println("Next market fetching time:", coin.NextMarketFetchTime())
	coin.Close()
	fmt.Println("Closed")
}



