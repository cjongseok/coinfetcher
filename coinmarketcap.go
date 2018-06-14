package fetcher

import (
	"sync"
	"time"
	"github.com/cjongseok/slog"
	api "github.com/cjongseok/go-coinmarketcap"
	"strings"
	"sort"
)

//type SearchRank struct {
//	keyIndex int
//	matchRank int
//}

const (
	logTag = "[CoinFetcher]"
	//normalDelay  = 5 * time.Minute // see https://coinmarketcap.com/api/#Limits
	defaultDelay = time.Minute
	retryDelay   = 5 * time.Second
	allCoinLimit = 10000
)

var normalDelay time.Duration
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
var symbolMap map[string]*api.Coin
var nameMap map[string]*api.Coin
var total api.GlobalMarketData
var coinFetchedTime time.Time
var marketFetchedTime time.Time
var nextCoinFetchTime time.Time
var nextMarketFetchTime time.Time

// 0 means all the coins
func Start() chan []api.Coin {
	return StartLimit(defaultDelay, 0)
}
func StartLimit(delay time.Duration, coinLimit int) chan []api.Coin {
	m.Lock()
	defer m.Unlock()
	if started {
		return nil
	}
	normalDelay = delay
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
				slog.Logf(logTag, "market fetch failure: %s\n", err)
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

func fetchCoin() chan []api.Coin {
	out := make(chan []api.Coin)
	go func() {
		defer wg.Done()
		var fetchDelay time.Duration
		streaming := true
		streamMutex := sync.Mutex{}
		isStreaming := func() bool {
			streamMutex.Lock()
			defer streamMutex.Unlock()
			return streaming
		}
		stopStreaming := func() {
			streamMutex.Lock()
			defer streamMutex.Unlock()
			streaming = false
		}
		fetch := func() (changed map[string]*api.Coin) {
			changed = make(map[string]*api.Coin)
			sMap, nMap, err := api.GetAllCoinData(limit)
			if err != nil {
				slog.Logf(logTag, "coin fetch failure: %s\n", err)
				fetchDelay = retryDelay
				return
			}
			for k, new := range sMap {
				old, ok := symbolMap[k]
				if !ok || (ok && old != new) {
					changed[k] = new
				}
			}
			symbolMap = sMap
			nameMap = nMap
			if !coinFetched {
				coinFetched = true
				coinFetchWg.Done()
			}
			fetchDelay = normalDelay
			coinFetchedTime = time.Now()
			return
		}
		stream := func(coins map[string]*api.Coin, to chan []api.Coin) {
			if len(coins) < 1 {
				return
			}
			defer func() {
				// handle panics on pushing to closed channel
				recover()
			}()
			slice := make([]api.Coin, len(coins))
			i := 0
			for _, c := range coins {
				slice[i] = *c
				i++
			}
			if isStreaming() {
				to <- slice
			}
		}
		newCoins := fetch()
		go stream(newCoins, out)
		for {
			nextCoinFetchTime = coinFetchedTime.Add(fetchDelay)
			select {
			case <-coinFetchInterrupt:
				stopStreaming()
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
		return len(symbolMap)
	}
	return 0
}
func All() map[string]api.Coin {
	if coinFetched {
		// copy
		hardcopy := make(map[string]api.Coin)
		for k, v := range symbolMap {
			hardcopy[k] = *v
		}
		return hardcopy
	}
	return nil
}
func Get(coinsymbol string) api.Coin {
	symbol := strings.ToUpper(coinsymbol)
	if coinFetched {
		if coin, ok := symbolMap[symbol]; ok && coin != nil {
			return *coin
		}
	}
	return api.Coin{}
}
func CoinByName(name string) api.Coin {
	if coinFetched {
		return *nameMap[strings.ToLower(name)]
	}
	return api.Coin{}
}
// Search finds coins by keywords which is a sentence allowing spaces.
// It tokenizes keywords by spaces.
func Search(keywords string) (exact, partial []api.Coin) {
	//exactRanks = make(map[string]SearchRank)
	//partialRanks = make(map[string]SearchRank)
	matched := make(map[string]bool)

	for _, key := range strings.Split(strings.ToLower(keywords), " ") {
		//var exactMatches, partialMatches []api.Coin

		// exact match
		if c, ok := symbolMap[strings.ToUpper(key)]; ok && !matched[c.Symbol] {
			exact = append(exact, *c)
			matched[c.Symbol] = true
		} else if c, ok := nameMap[strings.ToLower(key)]; ok && !matched[c.Symbol] {
			exact = append(exact, *c)
			matched[c.Symbol] = true
		}

		// partial match
		var symbolContainsKey, keyContainsSymbol []api.Coin
		for s, c := range symbolMap {
			symbol := strings.ToLower(s)
			if strings.Contains(symbol, key) && !matched[c.Symbol] {
				symbolContainsKey = append(symbolContainsKey, *c)
				matched[c.Symbol] = true
			} else if strings.Contains(key, symbol) && !matched[c.Symbol] {
				keyContainsSymbol = append(keyContainsSymbol, *c)
				matched[c.Symbol] = true
			}
		}
		var nameContainsKey, keyContainsName []api.Coin
		for n, c := range nameMap {
			name := strings.ToLower(n)
			if strings.Contains(name, key) && !matched[c.Symbol] {
				nameContainsKey = append(nameContainsKey, *c)
				matched[c.Symbol] = true
			} else if strings.Contains(key, name) && !matched[c.Symbol] {
				keyContainsName = append(keyContainsName, *c)
				matched[c.Symbol] = true
			}
		}

		// sort partial matches by symbol or name length
		lessSymbolLen := func(coins []api.Coin) (func(i, j int) bool) {
			return func(i, j int) bool {
				if len(coins[i].Symbol) == len(coins[j].Symbol) {
					return coins[i].Rank < coins[j].Rank
				}
				return len(coins[i].Symbol) < len(coins[j].Symbol)
			}
		}
		lessNameLen := func(coins []api.Coin) (func(i, j int) bool) {
			return func(i, j int) bool {
				if len(coins[i].Name) == len(coins[j].Name) {
					return coins[i].Rank < coins[j].Rank
				}
				return len(coins[i].Name) < len(coins[j].Name)
			}
		}
		sort.Slice(symbolContainsKey, lessSymbolLen(symbolContainsKey))
		sort.Slice(keyContainsSymbol, lessSymbolLen(keyContainsSymbol))
		sort.Slice(nameContainsKey, lessNameLen(nameContainsKey))
		sort.Slice(keyContainsName, lessNameLen(keyContainsName))
		partial = append(partial,
			append(symbolContainsKey,
				append(nameContainsKey,
					append(keyContainsSymbol, keyContainsName...)...)...)...)


		// rank matches
		//for matchRank, c := range exactMatches {
		//	if _, ok := exactRanks[c.Symbol]; !ok {
		//		exactRanks[c.Symbol] = SearchRank{keyIndex, matchRank}
		//	}
		//}
		//for matchRank, c := range partialMatches {
		//	if _, ok := partialRanks[c.Symbol]; !ok {
		//		partialRanks[c.Symbol] = SearchRank{keyIndex, matchRank}
		//	}
		//}
	}
	return exact, partial
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
	if coinFetchInterrupt != nil {
		close(coinFetchInterrupt)
	}
	if marketFetchInterrupt != nil {
		close(marketFetchInterrupt)
	}
	wg.Wait()
}
