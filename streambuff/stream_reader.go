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
	cache *bytes.Buffer

	// 源数据读取器
	originReader io.Reader

	// 包装读取器
	reader io.Reader
}

var bufferPool = sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{}
	},
}

func NewStreamReader(reader io.Reader) *StreamReader {
	if reader == nil {
		panic("reader must not nil")
	}
	cache := bufferPool.Get().(*bytes.Buffer)
	return &StreamReader{
		originReader: reader,
		reader:       io.TeeReader(reader, cache),
		cache:        cache,
	}
}

func (c *StreamReader) Read(p []byte) (int, error) {
	n, err := c.reader.Read(p)
	if err != nil && !errors.Is(err, io.EOF) {
		return n, errors.Join(err, ErrRead)
	}
	return n, err
}

// 重置读取偏移量为0， 但是不会清空已经缓存的内容，未缓存的内容需要被读取时，继续从源 reader 中读取
func (c *StreamReader) Reset() error {
	if c.cache == nil {
		panic("StreamReader cache is nil, cache may have been cleared")
	}

	// 先从已经缓存的数据中再次读取，然后再继续从原始 reader 中读取
	// 原始 reader 读取过的内容会追加到缓存中，再次重置读取后，会优先读取缓存中的内容
	c.reader = io.MultiReader(
		bytes.NewReader(c.cache.Bytes()),
		io.TeeReader(c.originReader, c.cache),
	)
	return nil
}

// 清空缓存，并将 buffer 返还给 bufferPool
func (c *StreamReader) Clean() (err error) {
	if r, ok := c.originReader.(io.ReadCloser); ok {
		err = r.Close()
	}

	c.cache.Reset()
	bufferPool.Put(c.cache)
	c.cache = nil
	return
}
