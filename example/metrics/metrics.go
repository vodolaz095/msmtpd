package main

import (
	"log"
	_ "net/http/pprof" // it is ok to make default HTTP server to publish debug information

	"github.com/vodolaz095/msmtpd"
)

// Example with metrics and profiling.

// Profiling info will be published on http://localhost:3000/debug/pprof/
// Metrics will be published on http://localhost:3000/metrics

// Good read:
// https://pkg.go.dev/net/http/pprof
// https://www.influxdata.com/products/data-collection/scrapers/
// https://prometheus.io/docs/instrumenting/exposition_formats/

func main() {
	var err error
	server := msmtpd.Server{
		Hostname:       "localhost",
		WelcomeMessage: "Do you believe in our God?",
	}
	go func() {
		err = server.StartPrometheusScrapperEndpoint(":3000", "/metrics")
		if err != nil {
			log.Fatalf("%s : while starting metrics scrapper endpoint", err)
		}
	}()
	err = server.ListenAndServe(":1025")
	if err != nil {
		log.Fatalf("%s : while starting server on 0.0.0.0:1025", err)
	}
}
