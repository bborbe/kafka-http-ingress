// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ingress

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/Shopify/sarama"
	"github.com/bborbe/kafka-http-ingress/avro"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/seibert-media/go-kafka/schema"
)

// HttpHandler that send a message to Kafka for each HTTP request.
type HttpHandler struct {
	KafkaTopic     string
	Producer       sarama.SyncProducer
	SchemaRegistry SchemaRegistry
	BodySizeLimit  int
}

// ServeHTTP transform each HTTP request to a Kafka message.
func (h *HttpHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if err := h.sendMessage(req); err != nil {
		glog.Warningf("handle request failed: %v", err)
		http.Error(resp, "handle request failed", http.StatusInternalServerError)
		return
	}
	glog.V(2).Infof("handle request successful")
	resp.WriteHeader(http.StatusOK)
	fmt.Fprint(resp, "ok")
}

func (h *HttpHandler) sendMessage(req *http.Request) error {
	valueEncoder, err := h.createValueEncoder(req)
	if err != nil {
		return err
	}
	partition, offset, err := h.Producer.SendMessage(&sarama.ProducerMessage{
		Topic: h.KafkaTopic,
		Key:   sarama.StringEncoder(req.RemoteAddr),
		Value: valueEncoder,
	})
	if err != nil {
		return errors.Wrap(err, "send message to kafka failed")
	}
	glog.V(3).Infof("send message successful to %s with partition %d offset %d", h.KafkaTopic, partition, offset)
	return nil
}

func (h *HttpHandler) createValueEncoder(req *http.Request) (sarama.Encoder, error) {
	httpRequest, err := h.createHttpRequest(req)
	if err != nil {
		return nil, err
	}
	schemaId, err := h.SchemaRegistry.SchemaId(fmt.Sprintf("%s-value", h.KafkaTopic), httpRequest.Schema())
	if err != nil {
		return nil, errors.Wrap(err, "get schema id failed")
	}
	buf := &bytes.Buffer{}
	if err := httpRequest.Serialize(buf); err != nil {
		return nil, errors.Wrap(err, "serialize httpRequest failed")
	}
	return &schema.AvroEncoder{SchemaId: schemaId, Content: buf.Bytes()}, nil
}

func (h *HttpHandler) createHttpRequest(req *http.Request) (*avro.HttpRequest, error) {
	var err error
	httpRequest := avro.NewHttpRequest()
	httpRequest.Body, err = h.readBody(req.Body)
	if err != nil {
		return nil, errors.Wrap(err, "read request failed")
	}
	for k, v := range req.Header {
		httpRequest.Header[k] = v
	}
	return httpRequest, nil
}

func (h *HttpHandler) readBody(r io.ReadCloser) ([]byte, error) {
	defer r.Close()
	if h.BodySizeLimit == 0 {
		return ioutil.ReadAll(r)
	}
	buf := make([]byte, h.BodySizeLimit+1)
	n, err := r.Read(buf)
	if err != nil {
		return nil, err
	}
	if n > h.BodySizeLimit {
		return nil, errors.New("max body size exeeded")
	}
	return buf[:n], nil
}
