package main

import (
	coin "github.com/cjongseok/coinfetcher"
	"fmt"
	"time"
)

func main() {
	fmt.Println("Start")
	coin.Start(0)
	fmt.Printf("%s: Fetched? %v\n", time.Now(), coin.Fetched())
	coin.WaitForFetching()
	fmt.Printf("%s: Fetched? %v\n", time.Now(), coin.Fetched())
	fmt.Println("Coins:", coin.All())
	fmt.Println("Market:", coin.TotalMarket())
	fmt.Println("Size:", coin.Size())
	fmt.Println("Get LTC:", coin.Get("LTC"))
	fmt.Println("Now:", time.Now())
	fmt.Println("Next coin fetching time:", coin.NextCoinFetchTime())
	fmt.Println("Next market fetching time:", coin.NextMarketFetchTime())
	coin.Close()
	fmt.Println("Closed")
}



