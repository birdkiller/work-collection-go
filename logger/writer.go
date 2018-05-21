package logger

import (
	"bytes"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// 通过文件写入
type FileWriter struct {
	Filepath     string
	cache        *bytes.Buffer
	lock         sync.Mutex
	flushflag    int
	Cachesize    int // 缓冲区大小
	AutoFlushLag int // 自动清空缓冲区延迟 0未不设置
}

func (self FileWriter) writeFile(b []byte) (int, error) {
	dir := filepath.Dir(self.Filepath)
	err := ensurePath(dir)
	if err != nil {
		return 0, err
	}
	file, err := os.OpenFile(self.Filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	return file.Write(b)
}

// Flush 立即清空缓冲区写入文件
func (self FileWriter) Flush() {
	self.lock.Lock()
	defer self.lock.Unlock()
	if self.flushflag == 0 {
		return
	}
	if self.cache == nil {
		return
	}
	data := make([]byte, self.cache.Len())
	self.cache.Read(data)
	self.cache.Reset()
	self.writeFile(data)
	self.flushflag = 0

}

func (self FileWriter) Write(b []byte) (int, error) {
	if self.Cachesize == 0 {
		// 不启用缓存，直接写入文件
		return self.writeFile(b)
	}
	if self.cache == nil {
		self.cache = new(bytes.Buffer)
	}

	size, err := self.cache.Write(b)
	self.flushflag = 1
	if self.cache.Len() >= self.Cachesize {
		self.Flush()
	}

	go func() {
		time.Sleep(time.Second * time.Duration(self.AutoFlushLag))
		self.Flush()
	}()
	return size, err
}
