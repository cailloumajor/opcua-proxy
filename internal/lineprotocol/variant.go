package lineprotocol

import "github.com/gopcua/opcua/ua"

//go:generate moq -out variant_mocks_test.go . VariantProvider

// VariantProvider is a consumer contract modelling an OPC-UA variant provider.
type VariantProvider interface {
	Bool() bool
	ByteString() []byte
	Float() float64
	Int() int64
	String() string
	Type() ua.TypeID
	Uint() uint64
}
