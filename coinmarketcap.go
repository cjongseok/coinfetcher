package coinfetcher

import (
	"sync"
	"time"
	api "github.com/cjongseok/go-coinmarketcap"
)

const (
	//normalDelay  = 5 * time.Minute // see https://coinmarketcap.com/api/#Limits
	retryDelay   = 5 * time.Second
	allCoinLimit = 10000
)

var normalDelay = 5 * time.Minute
var limit = allCoinLimit
var m sync.Mutex
var wg sync.WaitGroup
var coinFetchWg sync.WaitGroup
var marketFetchWg sync.WaitGroup
var started bool
var coinFetched bool
var marketFetched bool
var coinFetchInterrupt chan struct{}
var marketFetchInterrupt chan struct{}
var all map[string]api.Coin
var total api.GlobalMarketData
var coinFetchedTime time.Time
var marketFetchedTime time.Time
var nextCoinFetchTime time.Time
var nextMarketFetchTime time.Time
// 0 means all the coins
func Start(period time.Duration) chan api.Coin {
	return StartLimit(period, 0)
}
func StartLimit(period time.Duration, coinLimit int) chan api.Coin {
	m.Lock()
	defer m.Unlock()
	if started {
		return nil
	}
	normalDelay = period
	if coinLimit != 0 {
		limit = coinLimit
	}
	coinFetchInterrupt = make(chan struct{})
	marketFetchInterrupt = make(chan struct{})
	wg = sync.WaitGroup{}
	coinFetchWg = sync.WaitGroup{}
	marketFetchWg = sync.WaitGroup{}
	wg.Add(2)
	coinFetchWg.Add(1)
	marketFetchWg.Add(1)
	fetchMarket()
	coinUpdates := fetchCoin()
	started = true
	return coinUpdates
}
func WaitForFetching() {
	if !started {
		return
	}
	coinFetchWg.Wait()
	marketFetchWg.Wait()
}
func fetchMarket() {
	go func() {
		defer wg.Done()
		var fetchDelay time.Duration
		fetch := func() bool {
			market, err := api.GetMarketData()
			if err != nil {
				fetchDelay = retryDelay
				return false
			}
			total = market
			if !marketFetched {
				marketFetched = true
				marketFetchWg.Done()
			}
			fetchDelay = normalDelay
			marketFetchedTime = time.Now()
			return true
		}
		fetch()
		for {
			nextMarketFetchTime = marketFetchedTime.Add(fetchDelay)
			select {
			case <-marketFetchInterrupt:
				return
			case <-time.After(fetchDelay):
				fetch()
			}
		}
	}()
}
func fetchCoin() chan api.Coin {
	out := make(chan api.Coin)
	go func() {
		defer wg.Done()
		var fetchDelay time.Duration
		var stopStreaming bool
		fetch := func() (changed map[string]api.Coin) {
			changed = make(map[string]api.Coin)
			coins, err := api.GetAllCoinData(limit)
			if err != nil {
				fetchDelay = retryDelay
				return
			}
			for k, new := range coins {
				old, ok := all[k]
				if !ok || (ok && old != new) {
					changed[k] = new
				}
			}
			all = coins
			if !coinFetched {
				coinFetched = true
				coinFetchWg.Done()
			}
			fetchDelay = normalDelay
			coinFetchedTime = time.Now()
			return
		}
		stream := func(coins map[string]api.Coin, to chan api.Coin) {
			defer func() {
				// handle panics on pushing to closed channel
				recover()
			}()
			for _, c := range coins {
				if !stopStreaming {
					to <- c
				}
			}
		}
		newCoins := fetch()
		go stream(newCoins, out)
		for {
			nextCoinFetchTime = coinFetchedTime.Add(fetchDelay)
			select {
			case <-coinFetchInterrupt:
				stopStreaming = true
				close(out)
				return
			case <-time.After(fetchDelay):
				updated := fetch()
				go stream(updated, out)
			}
		}
	}()
	return out
}
func Fetched() bool {
	return coinFetched && marketFetched
}
func Size() int {
	if coinFetched {
		return len(all)
	}
	return 0
}
func All() map[string]api.Coin {
	if coinFetched {
		return all
	}
	return nil
}
func Get(coinsymbol string) api.Coin {
	if coinFetched {
		return all[coinsymbol]
	}
	return api.Coin{}
}
func TotalMarket() api.GlobalMarketData {
	return total
}
func NextCoinFetchTime() time.Time {
	return nextCoinFetchTime
}
func NextMarketFetchTime() time.Time {
	return nextMarketFetchTime
}
func Close() {
	close(coinFetchInterrupt)
	close(marketFetchInterrupt)
	wg.Wait()
}
