package streambuff

import (
	"bytes"
	"errors"
	"io"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func randomInt(n int) int {
	return rand.New(rand.NewSource(time.Now().Unix())).Intn(n)
}

func randomString(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// 测试基本的读
func TestStreamReader(t *testing.T) {
	origin := []byte(randomString(randomInt(2000)))
	reader := NewStreamReader(
		bytes.NewReader(origin),
	)

	var data []byte
	size := 3
	for {
		buffer := make([]byte, size)
		n, err := reader.Read(buffer)
		if n == 0 {
			break
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else {
				assert.NoError(t, err)
			}
		}
		assert.Equal(t, buffer[0:n], origin[len(data):len(data)+n])
		data = append(data, buffer[0:n]...)
	}

	reader.Close()
	assert.Equal(t, data, origin)
}

// 测试通过 io.ReadAll 读取
func TestStreamReaderByReadAll(t *testing.T) {
	origin := []byte(randomString(randomInt(2000)))
	reader := NewStreamReader(
		bytes.NewReader(origin),
	)

	data, err := io.ReadAll(reader)
	assert.NoError(t, err)

	reader.Close()
	assert.Equal(t, data, origin)
}

// 测试通过 Reset 重置后，重复读取
func TestStreamReaderReset(t *testing.T) {
	origin := []byte(randomString(randomInt(2000)))
	reader := NewStreamReader(
		bytes.NewReader(origin),
	)

	data1, err := io.ReadAll(reader)
	assert.NoError(t, err)

	reader.Reset()
	data2, err := io.ReadAll(reader)
	assert.NoError(t, err)

	reader.Close()
	assert.Equal(t, origin, data1)
	assert.Equal(t, origin, data2)
}

// 测试使用 io.ReadAll 和 Read 组合读取
func TestStreamReaderCombin(t *testing.T) {
	origin := []byte(randomString(randomInt(2000)))
	reader := NewStreamReader(
		bytes.NewReader(origin),
	)

	var data1 []byte
	times := 4
	timesSize := 3
	for i := 0; i < times; i++ {
		buffer := make([]byte, timesSize)
		n, err := reader.Read(buffer)
		if n == 0 {
			break
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else {
				assert.NoError(t, err)
			}
		}
		assert.Equal(t, buffer[0:n], origin[len(data1):len(data1)+n])
		data1 = append(data1, buffer[0:n]...)
	}

	data2, err := io.ReadAll(reader)
	assert.NoError(t, err)

	reader.Close()
	assert.Equal(t, origin, append(data1, data2...))
}

// 测试读取失败后，通过Reset重置之后再次读取
func TestStreamReaderBreak(t *testing.T) {
	origin := []byte(randomString(randomInt(2000)))
	reader := NewStreamReader(
		bytes.NewReader(origin),
	)

	for i := 0; i < 10; i++ {
		buffer := make([]byte, 10)
		n, err := reader.Read(buffer)
		if n == 0 {
			break
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else {
				assert.NoError(t, err)
			}
		}
	}
	reader.Reset()

	data, err := io.ReadAll(reader)
	assert.NoError(t, err)

	reader.Close()
	assert.Equal(t, origin, data)
}

func BenchmarkStreamReaderReadAll(b *testing.B) {
	for n := 0; n < b.N; n++ {
		b.StopTimer()
		origin := []byte(randomString(randomInt(2000)))
		reader := NewStreamReader(
			bytes.NewReader(origin),
		)
		b.StartTimer()

		data, err := io.ReadAll(reader)
		assert.NoError(b, err)

		reader.Close()
		assert.Equal(b, origin, data)
	}
}

func BenchmarkStreamReaderRead(b *testing.B) {
	for n := 0; n < b.N; n++ {
		b.StopTimer()
		origin := []byte(randomString(randomInt(2000)))
		reader := NewStreamReader(
			bytes.NewReader(origin),
		)
		b.StartTimer()

		var data []byte
		size := 128
		for {
			buffer := make([]byte, size)
			n, err := reader.Read(buffer)
			if n == 0 {
				break
			}
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				} else {
					assert.NoError(b, err)
				}
			}
			assert.Equal(b, buffer[0:n], origin[len(data):len(data)+n])
			data = append(data, buffer[0:n]...)
		}
		reader.Close()
		assert.Equal(b, data, origin)
	}
}
