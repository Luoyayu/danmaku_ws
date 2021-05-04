package danmaku_ws

import (
	"bytes"
	"compress/zlib"
	"io"
	"log"
)

func (c *Client) Log(v ...interface{}) {
	if c.Debug {
		log.Println(v...)
	}
}

func (c *Client) Logf(f string, v ...interface{}) {
	if c.Debug {
		log.Printf(f, v...)
	}
}

func zlibDecompress(compressSrc []byte) []byte {
	var o bytes.Buffer
	r, _ := zlib.NewReader(bytes.NewReader(compressSrc))
	io.Copy(&o, r)
	return o.Bytes()
}
