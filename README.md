
## stream-buffer

带有缓存的高级 Reader, 只会在第一次 read 的时候从源 reader 调用 read()
读取过的数据会被缓存下来，之后的重试操作可以使用缓存中的数据

- 通过 Reset() 重置读取进度, Reset() 不会清除已经缓存的内容
- 通过 Clean() 关闭 Reader, Clean() 会清理已经缓存到的内容
### 使用方法

```golang

func sample() {
	origin := []byte(".................xxxx...aaa.aa")

	reader := streambuff.NewStreamReader(
		bytes.NewReader(origin),
	)
	data1, err := io.ReadAll(reader)

	reader.Reset()
	data2, err := io.ReadAll(reader)
	
	reader.Clean()
	// data1 == data2 == origin
}

```

```go
func sample() {
	origin := []byte(".................xxxx...aaa.aa")

	reader := streambuff.NewStreamReader(
		bytes.NewReader(origin),
	)
	
	var data []byte
	for {
		buffer := make([]byte, 128)
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
		data = append(data, buffer[0:n]...)
	}

	reader.Clean()
	// data == origon
}
```