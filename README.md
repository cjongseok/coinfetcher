coinfetcher
===
Cache for CoinMarketCap ticks.
It periodically fetch coin ticks from [coinmarketcap.com](https://coinmarketcap.com/api) using [miguelmota/go-coinmarketcap](https://github.com/miguelmota/go-coinmarketcap).

Usage
---
```go
// Turn on the fetcher
coinfetcher.Start()         // default fetching delay is 5 minutes.
coinfetcher.WaitForFetching()

// Get ticks
coinfetcher.Get("BTC")      // recent BTC tick
coinfetcher.All()           // all the recent ticks
coinfetcher.TotalMarket()   // market data

// Retrieve ticks
exactMatches, partialMatches := coinfetcher.Search("neo eos")
```

See Also
---
* [cjongseok/fetch-bittrex](https://github.com/cjongseok/fetch-bittrex)
