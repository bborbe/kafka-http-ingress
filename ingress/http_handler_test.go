// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ingress_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"

	sarama_mocks "github.com/Shopify/sarama/mocks"
	"github.com/bborbe/kafka-http-ingress/ingress"
	"github.com/bborbe/kafka-http-ingress/mocks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pkg/errors"
)

var _ = Describe("Http Handler", func() {
	var handler *ingress.HttpHandler
	var responseRecorder *httptest.ResponseRecorder
	var schemaRegistry *mocks.SchemaRegistry
	var producer *sarama_mocks.SyncProducer
	BeforeEach(func() {
		responseRecorder = &httptest.ResponseRecorder{}
		schemaRegistry = &mocks.SchemaRegistry{}
		producer = sarama_mocks.NewSyncProducer(GinkgoT(), nil)
		handler = &ingress.HttpHandler{
			KafkaTopic:     "my-topic",
			SchemaRegistry: schemaRegistry,
			Producer:       producer,
		}
	})
	It("returns no error", func() {
		producer.ExpectSendMessageAndSucceed()
		handler.ServeHTTP(responseRecorder, httptest.NewRequest(http.MethodGet, "http://localhost:8080", nil))
		Expect(responseRecorder.Code).To(Equal(http.StatusOK))
	})
	It("return error if message delivery failed", func() {
		producer.ExpectSendMessageAndFail(errors.New("banana"))
		handler.ServeHTTP(responseRecorder, httptest.NewRequest(http.MethodGet, "http://localhost:8080", nil))
		Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError))
	})
	It("contains error message if message delivery failed", func() {
		responseRecorder.Body = &bytes.Buffer{}
		producer.ExpectSendMessageAndFail(errors.New("banana"))
		handler.ServeHTTP(responseRecorder, httptest.NewRequest(http.MethodGet, "http://localhost:8080", nil))
		Expect(gbytes.BufferWithBytes(responseRecorder.Body.Bytes())).To(gbytes.Say("handle request failed"))
	})
	It("have message content", func() {
		producer.ExpectSendMessageWithCheckerFunctionAndSucceed(func(val []byte) error {
			Expect(len(val)).To(BeNumerically(">", 0))
			return nil
		})
		handler.ServeHTTP(responseRecorder, httptest.NewRequest(http.MethodGet, "http://localhost:8080", nil))
	})
	It("returns error if body size limit is exeeded", func() {
		handler.BodySizeLimit = 10
		body := bytes.NewBuffer(make([]byte, 11))
		handler.ServeHTTP(responseRecorder, httptest.NewRequest(http.MethodGet, "http://localhost:8080", body))
		Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError))
	})
	It("returns not error if body size limit is limit", func() {
		producer.ExpectSendMessageAndSucceed()
		handler.BodySizeLimit = 10
		body := bytes.NewBuffer(make([]byte, 10))
		handler.ServeHTTP(responseRecorder, httptest.NewRequest(http.MethodGet, "http://localhost:8080", body))
		Expect(responseRecorder.Code).To(Equal(http.StatusOK))
	})
})
