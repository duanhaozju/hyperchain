package threadsafe

type WeightItem interface {
	Value() interface{}
	Weight() int
}
