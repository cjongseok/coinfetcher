coinfetcher
===
Periodically fetch cryptocurrency data from **coinmarketcap.com** using **github.com/miguelmota/go-coinmarketcap**.
The fetcher pulls data every 5 minutes, which is same with the coinmarketcap.com API endpoint update cycle.

Usage
---
Turn on the fetcher
```go
import coin "github.com/cjongseok/coinfetcher"
coin.Start(0)
coin.WaitForFetching()
```
And get desired data anytime
```go
coin.Get("BTC")      // BTC data
coin.All()           // all coin data
coin.TotalMarket()   // market data

```

