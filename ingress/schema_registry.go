// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ingress

//go:generate counterfeiter -o ../mocks/schema_registry.go --fake-name SchemaRegistry . SchemaRegistry

// SchemaRegistry get the ID for the given subject and schema.
type SchemaRegistry interface {
	SchemaId(subject string, schema string) (uint32, error)
}
