package lineprotocol

func NewMockedBuilder(p Pooler) *Builder {
	return &Builder{pool: p}
}
