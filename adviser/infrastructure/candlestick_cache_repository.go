package infrastructure

import (
	"context"
	"sync"
	"time"

	"github.com/websmee/example_of_my_code/adviser/domain/candlestick"
	"github.com/websmee/example_of_my_code/adviser/domain/quote"
)

type candlestickCacheRepository struct {
	quoteRepository       quote.Repository
	candlestickRepository candlestick.Repository
	cache                 map[string]map[candlestick.Interval][]candlestick.Candlestick
	lock                  sync.RWMutex
}

func NewCandlestickCacheRepository(
	ctx context.Context,
	quoteRepository quote.Repository,
	candlestickRepository candlestick.Repository,
	intervals []candlestick.Interval,
	from, to time.Time,
) (candlestick.Repository, error) {
	cr := &candlestickCacheRepository{
		quoteRepository:       quoteRepository,
		candlestickRepository: candlestickRepository,
	}
	if err := cr.makeCache(ctx, intervals, from, to); err != nil {
		return nil, err
	}
	return cr, nil
}

func (r *candlestickCacheRepository) makeCache(ctx context.Context, intervals []candlestick.Interval, from, to time.Time) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	quotes, err := r.quoteRepository.GetQuotes(ctx)
	if err != nil {
		return err
	}

	symbols := make([]string, len(quotes))
	for i := range quotes {
		symbols[i] = quotes[i].Symbol
	}

	r.cache = make(map[string]map[candlestick.Interval][]candlestick.Candlestick)
	for i := range symbols {
		r.cache[symbols[i]] = make(map[candlestick.Interval][]candlestick.Candlestick)
		for j := range intervals {
			cs, err := r.candlestickRepository.GetCandlesticks(ctx, symbols[i], intervals[j], from, to)
			if err != nil {
				return err
			}

			r.cache[symbols[i]][intervals[j]] = cs
		}
	}

	return nil
}

func (r *candlestickCacheRepository) GetCandlesticks(_ context.Context, symbol string, interval candlestick.Interval, from, to time.Time) ([]candlestick.Candlestick, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	var c []candlestick.Candlestick
	if _, ok := r.cache[symbol]; !ok {
		return c, nil
	}
	if _, ok := r.cache[symbol][interval]; !ok {
		return c, nil
	}
	for i := range r.cache[symbol][interval] {
		if r.cache[symbol][interval][i].Timestamp.Unix() >= from.Unix() &&
			r.cache[symbol][interval][i].Timestamp.Unix() <= to.Unix() {
			c = append(c, r.cache[symbol][interval][i])
		}
	}

	return c, nil
}

func (r *candlestickCacheRepository) GetCandlesticksByCount(
	ctx context.Context,
	symbol string,
	interval candlestick.Interval,
	start time.Time,
	direction candlestick.GetterDirection,
	count int,
) ([]candlestick.Candlestick, error) {
	return candlestick.NewGreedyGetter(r).GetCandlesticksByCount(ctx, symbol, interval, start, direction, count)
}
