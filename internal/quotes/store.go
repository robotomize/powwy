package quotes

import "github.com/robotomize/powwy/pkg/hashcash"

type Store interface {
	Lookup(name string) (hashcash.Header, bool)
	Set(name string, object hashcash.Header) error
}
