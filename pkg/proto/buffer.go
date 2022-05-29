package proto

import (
	"bytes"
	"sync"
)

var buf = sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{}
	},
}

func GetBuf() (*bytes.Buffer, func()) {
	var b *bytes.Buffer
	ifc := buf.Get()
	if ifc != nil {
		b = ifc.(*bytes.Buffer)
	}

	b.Reset()

	return b, func() {
		buf.Put(b)
	}
}
