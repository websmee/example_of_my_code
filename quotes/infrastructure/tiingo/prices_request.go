package tiingo

import "time"

type ResponseResampleFreq string

const ResponseResampleFreqHour ResponseResampleFreq = "1hour"

type PricesRequest struct {
	Ticker       string
	StartDate    time.Time
	EndDate      time.Time
	ResampleFreq ResponseResampleFreq
}

func (r PricesRequest) GetPath() string {
	path := "/iex/" + r.Ticker + "/prices"
	path = path + "?columns=open,high,low,close,volume"
	path = path + "&startDate=" + r.StartDate.Format("2006-01-02")
	path = path + "&endDate=" + r.EndDate.Format("2006-01-02")
	path = path + "&resampleFreq=" + string(r.ResampleFreq)

	return path
}
