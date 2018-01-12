coinfetcher
===
Periodically fetch cryptocurrency data from [coinmarketcap.com](https://coinmarketcap.com/api) using [miguelmota/go-coinmarketcap](https://github.com/miguelmota/go-coinmarketcap).
The fetcher pulls data every 5 minutes, which is same with the coinmarketcap.com API endpoint update cycle.

Usage
---
Turn on the fetcher
```go
coinfetcher.Start()
coinfetcher.WaitForFetching()
```
And get desired data anytime
```go
coinfetcher.Get("BTC")      // BTC data
coinfetcher.All()           // all coin data
coinfetcher.TotalMarket()   // market data

```
