package query

type Queryable[DB any, Domain any] interface {
	Converter() func(*DB) (*Domain, error)
}
