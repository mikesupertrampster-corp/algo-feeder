package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/mikesupertrampster/algo-api/services/alphavantage"
	"github.com/mikesupertrampster/algo-feeder/pkg/exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
	"log"
	"net/http"
	"os"
	"time"
)

type cfg struct {
	Port string `envconfig:"default=8080"`

	SymbolsAPI struct {
		Url string `envconfig:"default=http://localhost:8000"`
	}

	AlphaVantage struct {
		ApiKey string `envconfig:"default=KEY"`
	}
}

type Symbols []string

var logger = logrus.New()

func init() {
	logger.SetOutput(os.Stdout)
	logger.SetReportCaller(true)
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		DisableColors: false,
		FullTimestamp: false,
	})
}

func main() {
	config := new(cfg)
	if err := envconfig.Init(config); err != nil {
		log.Fatal(err)
	}

	collect(config, symbols(config))
}

func symbols(config *cfg) []string {
	req, err := http.NewRequest(http.MethodPost, config.SymbolsAPI.Url, bytes.NewBuffer([]byte(`{"target":"symbols"}`)))
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer func(resp *http.Response) {
		err := resp.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(resp)

	var symbols Symbols
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&symbols)
	if err != nil {
		log.Fatal(err)
	}

	return symbols
}

func collect(config *cfg, symbols []string) {
	av := alphavantage.New(logger, config.AlphaVantage.ApiKey)

	stockExporter := exporter.NewStocksExporter(av, symbols)
	prometheus.MustRegister(stockExporter)

	if err := createHttpServer(config.Port).ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func createHttpServer(port string) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	return &http.Server{
		Handler:      mux,
		Addr:         fmt.Sprintf(":%s", port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
}
