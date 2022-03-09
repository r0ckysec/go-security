/**
 * @Description
 * @Author r0cky
 * @Date 2021/12/24 16:08
 **/
package secio

import (
	"bytes"
	"io"
	"sync"
)

var Buffer sync.Pool

func init() {
	Buffer = NewBufferPool()
}

func NewBufferPool() sync.Pool {
	return sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, 4096))
		},
	}
}

func ReadAll(r io.Reader) ([]byte, error) {
	buffer := Buffer.Get().(*bytes.Buffer)
	buffer.Reset()
	defer func() {
		if buffer != nil {
			Buffer.Put(buffer)
			buffer = nil
		}
	}()
	_, err := io.Copy(buffer, r)
	return buffer.Bytes(), err
}
