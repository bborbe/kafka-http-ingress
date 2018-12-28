// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ingress_test

import (
	"github.com/bborbe/kafka-http-ingress/ingress"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Webhook App", func() {
	var app *ingress.App
	BeforeEach(func() {
		app = &ingress.App{
			Port:                   1337,
			KafkaBrokers:           "kafka:9092",
			KafkaTopic:             "my-topic",
			KafkaSchemaRegistryUrl: "http://localhost:8081",
		}
	})
	It("Validate without error", func() {
		Expect(app.Validate()).NotTo(HaveOccurred())
	})
	It("Validate returns error if port is 0", func() {
		app.Port = 0
		Expect(app.Validate()).To(HaveOccurred())
	})
	It("Validate returns error KafkaBrokers is empty", func() {
		app.KafkaBrokers = ""
		Expect(app.Validate()).To(HaveOccurred())
	})
	It("Validate returns error KafkaTopic is empty", func() {
		app.KafkaTopic = ""
		Expect(app.Validate()).To(HaveOccurred())
	})
	It("Validate returns error KafkaSchemaRegistryUrl is empty", func() {
		app.KafkaSchemaRegistryUrl = ""
		Expect(app.Validate()).To(HaveOccurred())
	})
})
