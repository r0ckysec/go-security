package json

import (
	"bytes"
	"encoding/gob"
	"github.com/valyala/fastjson"
)

/**
 * @Description
 * @Author r0cky
 * @Date 2022/2/27 20:25
 */

var JsonPool = &fastjson.ParserPool{}

func GetBytes(key interface{}) ([]byte, error) {
	var buf bytes.Buffer
	defer buf.Reset()
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
