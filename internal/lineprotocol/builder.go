package lineprotocol

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/influxdata/line-protocol/v2/lineprotocol"
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

//go:generate moq -out builder_mocks_test.go . Pooler Encoder

// Pooler is a consumer contract modelling an encoder pool.
type Pooler interface {
	Get() any
	Put(x any)
}

// Encoder is a consumer contract modelling a line protocol encoder.
type Encoder interface {
	AddField(key string, value lineprotocol.Value)
	AddTag(key, value string)
	Bytes() []byte
	EndLine(t time.Time)
	Err() error
	Reset()
	StartLine(measurement string)
}

func sortedKeys[K constraints.Ordered, V any](m map[K]V) []K {
	sk := maps.Keys(m)
	slices.Sort(sk)
	return sk
}

// Builder represents an InfluxDB line protocol builder.
type Builder struct {
	pool Pooler
}

// NewBuilder returns a line protocol builder initialized with an encoder.
func NewBuilder() *Builder {
	p := &sync.Pool{
		New: func() any {
			return &lineprotocol.Encoder{}
		},
	}
	return &Builder{pool: p}
}

// Build builds a point in line protocol.
func (b *Builder) Build(w io.Writer, measurement string, tags map[string]string, fields map[string]VariantProvider, ts time.Time) error {
	enc := b.pool.Get().(Encoder)

	enc.Reset()

	enc.StartLine(measurement)

	for _, tk := range sortedKeys(tags) {
		enc.AddTag(tk, tags[tk])
	}

	for _, fk := range sortedKeys(fields) {
		val, err := NewValueFromVariant(fields[fk])
		if err != nil {
			return fmt.Errorf("error converting variant to line protocol value: %w", err)
		}
		enc.AddField(fk, val)
	}

	enc.EndLine(ts.UTC())

	if err := enc.Err(); err != nil {
		return fmt.Errorf("error encoding line protocol: %w", err)
	}

	if _, err := w.Write(enc.Bytes()); err != nil {
		return fmt.Errorf("error writing data: %w", err)
	}

	b.pool.Put(enc)

	return nil
}
