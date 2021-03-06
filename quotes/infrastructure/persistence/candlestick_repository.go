package persistence

import (
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/pkg/errors"

	"github.com/websmee/example_of_my_code/quotes/domain/candlestick"
	"github.com/websmee/example_of_my_code/quotes/domain/quote"
)

type CandlestickRepository struct {
	db *pg.DB
}

func NewCandlestickRepository(db *pg.DB) *CandlestickRepository {
	return &CandlestickRepository{db}
}

func (r CandlestickRepository) SaveCandlestick(candlestick *candlestick.Candlestick) error {
	_, err := r.db.Model(candlestick).
		OnConflict("(quote_id, interval, timestamp) DO UPDATE").
		Set("open = EXCLUDED.open").
		Set("low = EXCLUDED.low").
		Set("high = EXCLUDED.high").
		Set("close = EXCLUDED.close").
		Set("adj_close = EXCLUDED.adj_close").
		Set("volume = EXCLUDED.volume").
		Insert()

	return errors.Wrap(err, "SaveCandlestick failed")
}

func (r CandlestickRepository) GetCandlesticks(quote *quote.Quote, interval candlestick.Interval, from, to time.Time) ([]candlestick.Candlestick, error) {
	var candlesticks []candlestick.Candlestick

	err := r.db.Model(&candlestick.Candlestick{}).
		Where("quote_id = ?", quote.ID).
		Where("interval = ?", interval).
		Where("timestamp >= ?", from).
		Where("timestamp <= ?", to).
		Order("timestamp ASC").
		Select(&candlesticks)

	if err != nil && err != pg.ErrNoRows {
		return nil, errors.Wrap(err, "GetCandlesticks failed")
	}

	return candlesticks, nil
}

func (r CandlestickRepository) GetLastCandlestickTimestamp(quote *quote.Quote, interval candlestick.Interval) (time.Time, error) {
	var toReturn struct {
		Timestamp time.Time
	}

	err := r.db.Model().
		ColumnExpr("timestamp").
		TableExpr("candlesticks").
		Where("quote_id = ?", quote.ID).
		Where("interval = ?", interval).
		Order("timestamp ASC").
		Limit(1).
		Select(&toReturn)

	if err != nil {
		if err == pg.ErrNoRows {
			return time.Now(), nil
		}
		return time.Now(), errors.Wrap(err, "GetLastCandlestickTimestamp failed")
	}

	return toReturn.Timestamp, nil
}
