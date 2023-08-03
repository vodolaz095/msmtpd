package main

import (
	"log"

	"github.com/vodolaz095/msmtpd"
)

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
