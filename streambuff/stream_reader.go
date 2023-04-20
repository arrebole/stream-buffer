package streambuff

import (
	"bytes"
	"errors"
	"io"
	"sync"
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

var bufferPool = sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{}
	},
}

func NewStreamReader(reader io.Reader) *StreamReader {
	buf := bufferPool.Get().(*bytes.Buffer)
	return &StreamReader{
		buf:    buf,
		reader: reader,
	}
}

// 重置读取偏移量为0， 但是不会清空已经缓存的内容，未缓存的内容需要被读取时，继续从源 reader 中读取
func (c *StreamReader) Reset() {
	c.offset = 0
}

// 清空缓存，并将 buffer 返还给 bufferPool
func (c *StreamReader) Close() {
	c.buf.Reset()
	bufferPool.Put(c.buf)
}

func (c *StreamReader) Read(p []byte) (int, error) {
	// 此次 Read 需要的字节长度
	requireSize := c.offset + int64(len(p))

	// 还需要从源文件读取的长度
	needMoreSize := requireSize - int64(c.buf.Len())

	// 内容还需要从源 reader 读取，并且文件还没有读取完毕，则继续从reader读取
	if needMoreSize > 0 && !c.hasEOF {
		n, err := io.CopyN(c.buf, c.reader, int64(len(p)))
		if err != nil && errors.Is(err, io.EOF) {
			c.hasEOF = true
		}
		if err != nil && !errors.Is(err, io.EOF) {
			return int(n), err
		}
	}

	// 如果已经读取完毕，继续读取，则返回eof错误
	if c.offset >= int64(c.buf.Len()) {
		return 0, io.EOF
	}

	// 将需要的数据拷贝到
	size := copy(p, c.buf.Bytes()[c.offset:])
	c.offset += int64(size)

	return size, nil
}
