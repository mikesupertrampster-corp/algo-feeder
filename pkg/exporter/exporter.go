package exporter

import (
	"github.com/mikesupertrampster/algo-api/services/alphavantage"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
)

type StocksExporter struct {
	av      alphavantage.Client
	price   *prometheus.Desc
	symbols []string
}

func NewStocksExporter(av alphavantage.Client, symbols []string) *StocksExporter {
	return &StocksExporter{
		av:      av,
		price:   prometheus.NewDesc("price", "Stock Price", []string{"symbol"}, nil),
		symbols: symbols,
	}
}

func (e *StocksExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.price
}

func (e *StocksExporter) Collect(ch chan<- prometheus.Metric) {
	for _, symbol := range e.symbols {
		q, err := e.av.GetQuote(symbol)
		if err != nil {
			return
		}

		price, err := strconv.ParseFloat(q.GlobalQuote.Price, 64)
		if err != nil {
			return
		}

		ch <- prometheus.MustNewConstMetric(
			e.price, prometheus.GaugeValue, price, symbol,
		)
	}
}
