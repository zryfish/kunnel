package proxy

type Domainer interface {
	Next() string

	Invalidate(domain string)
}

type Domain struct {
}
