package infrastructure

import (
	"context"
	"sync"
	"time"

	"github.com/websmee/example_of_my_code/adviser/domain/candlestick"
)

type candlestickCacheRepository struct {
	repository candlestick.Repository
	cache      map[string]map[candlestick.Interval][]candlestick.Candlestick
	lock       sync.RWMutex
}

func NewCandlestickCacheRepository(
	ctx context.Context,
	repository candlestick.Repository,
	symbols []string,
	intervals []candlestick.Interval,
	from, to time.Time,
) (candlestick.Repository, error) {
	cr := &candlestickCacheRepository{
		repository: repository,
	}
	if err := cr.makeCache(ctx, symbols, intervals, from, to); err != nil {
		return nil, err
	}
	return cr, nil
}

func (r *candlestickCacheRepository) makeCache(ctx context.Context, symbols []string, intervals []candlestick.Interval, from, to time.Time) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.cache = make(map[string]map[candlestick.Interval][]candlestick.Candlestick)
	for i := range symbols {
		r.cache[symbols[i]] = make(map[candlestick.Interval][]candlestick.Candlestick)
		for j := range intervals {
			cs, err := r.repository.GetCandlesticks(ctx, symbols[i], intervals[j], from, to)
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
