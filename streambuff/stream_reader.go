package streambuff

import (
	"bytes"
	"errors"
	"io"
	"sync/atomic"
)

// 带有缓存的高级 Reader
// 只会在第一次 read 的时候从源 reader 调用 read
// 读取过的数据会被缓存下来，之后的重试操作可以使用缓存中的数据，通过 Reset 重置读取进度
type StreamReader struct {
	buf *bytes.Buffer

	// 源文件是否已经读完
	hasEOF bool

	// 对外读取到的偏移量
	offset int64

	// 源数据读取器的 bufio 包装
	reader io.Reader
}

func NewStreamReader(reader io.Reader) *StreamReader {
	return &StreamReader{
		buf:    &bytes.Buffer{},
		reader: reader,
	}
}

func (c *StreamReader) Reset() {
	atomic.SwapInt64(&c.offset, 0)
}

func (c *StreamReader) CachedSize() int64 {
	return int64(len(c.buf.Bytes()))
}

func (c *StreamReader) Read(p []byte) (int, error) {
	// 此次 Read 需要的字节长度
	requireSize := c.offset + int64(len(p))

	// 还需要从源文件读取的长度
	needMoreSize := requireSize - c.CachedSize()

	// 内容还需要从源 reader 读取，并且文件还没有读取完毕，则继续从reader读取
	if needMoreSize > 0 && !c.hasEOF {
		n, err := c.reader.Read(p)
		if n > 0 {
			c.buf.Write(p[0:n])
		}
		if err != nil && errors.Is(err, io.EOF) {
			c.hasEOF = true
		}
		if err != nil && !errors.Is(err, io.EOF) {
			return n, err
		}
	}

	// 如果已经读取完毕，继续读取，则返回eof错误
	if c.hasEOF && c.offset >= c.CachedSize() {
		return 0, io.EOF
	}

	size := 0
	// 如果未读取完毕，则返回需要读取的内容
	// - 需要读取的长度大于源文件的长度
	// - 需要读取的长度小于源文件的长度
	if requireSize > c.CachedSize() {
		size = copy(p, c.buf.Bytes()[c.offset:])
		atomic.SwapInt64(&c.offset, c.CachedSize())
	} else {
		size = copy(p, c.buf.Bytes()[c.offset:requireSize])
		atomic.AddInt64(&c.offset, int64(size))
	}
	return size, nil
}
