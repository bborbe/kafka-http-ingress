// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ingress

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Shopify/sarama"
	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/seibert-media/go-kafka/schema"
)

// App for transform HTTP requests to Kafka messages.
type App struct {
	KafkaBrokers           string
	KafkaTopic             string
	KafkaSchemaRegistryUrl string
	Port                   int
	BodySizeLimit          int
}

// Validate all required parameters are set.
func (a *App) Validate() error {
	if a.Port <= 0 {
		return errors.New("Port invalid")
	}
	if a.KafkaBrokers == "" {
		return errors.New("KafkaBrokers missing")
	}
	if a.KafkaSchemaRegistryUrl == "" {
		return errors.New("KafkaSchemaRegistryUrl missing")
	}
	if a.KafkaTopic == "" {
		return errors.New("KafkaTopic missing")
	}
	return nil
}

// Run HTTP Server.
func (a *App) Run(ctx context.Context) error {

	config := sarama.NewConfig()
	config.Version = sarama.V2_0_0_0
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 10
	config.Producer.Return.Successes = true

	client, err := sarama.NewClient(strings.Split(a.KafkaBrokers, ","), config)
	if err != nil {
		return errors.Wrap(err, "create client failed")
	}
	defer client.Close()

	producer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		return errors.Wrap(err, "create sync producer failed")
	}
	defer producer.Close()

	router := mux.NewRouter()
	router.Path("/healthz").HandlerFunc(a.check)
	router.Path("/readiness").HandlerFunc(a.check)
	router.Path("/metrics").Handler(promhttp.Handler())
	router.NotFoundHandler = &HttpHandler{
		KafkaTopic: a.KafkaTopic,
		SchemaRegistry: &schema.Registry{
			HttpClient:        http.DefaultClient,
			SchemaRegistryUrl: a.KafkaSchemaRegistryUrl,
		},
		Producer: producer,
	}
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", a.Port),
		Handler: router,
	}
	go func() {
		select {
		case <-ctx.Done():
			if err := server.Shutdown(ctx); err != nil {
				glog.Warningf("shutdown failed: %v", err)
			}
		}
	}()
	return server.ListenAndServe()
}

func (a *App) check(resp http.ResponseWriter, req *http.Request) {
	resp.WriteHeader(200)
	fmt.Fprintf(resp, "ok")
}
