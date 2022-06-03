package quotes

import (
	"strings"
	"sync"
)

var buf = sync.Pool{
	New: func() interface{} {
		return &strings.Builder{}
	},
}

func GetBuilder() (*strings.Builder, func()) {
	var b *strings.Builder
	ifc := buf.Get()
	if ifc != nil {
		b = ifc.(*strings.Builder)
	}

	b.Reset()

	return b, func() {
		buf.Put(b)
	}
}
