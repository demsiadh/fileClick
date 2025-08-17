package system

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fileClick/config"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

// WalRecord 表示一条WAL记录
type WalRecord struct {
	FileId uint64
	Ts     int64
}

// WalThread 单个WAL线程，负责写入一个WAL文件
type WalThread struct {
	threadId int
	dir      string
	maxSize  int64

	mu      sync.Mutex
	curFile *os.File
	writer  *bufio.Writer
	curSize int64
	seq     int
}

// Wal 多线程WAL管理器
type Wal struct {
	dir        string
	maxSize    int64
	threads    [config.WalThreads]*WalThread
	shutdown   chan struct{}
	wg         sync.WaitGroup
	randomSeed uint64
}

func NewWAL() (*Wal, error) {
	w := &Wal{
		dir:      config.WalPath,
		maxSize:  config.WalMaxSize,
		shutdown: make(chan struct{}),
	}

	// 初始化5个WAL线程
	for i := 0; i < config.WalThreads; i++ {
		thread, err := w.newWalThread(i)
		if err != nil {
			return nil, fmt.Errorf("failed to create WAL thread %d: %w", i, err)
		}
		w.threads[i] = thread
	}

	return w, nil
}

func (w *Wal) newWalThread(threadId int) (*WalThread, error) {
	wt := &WalThread{
		threadId: threadId,
		dir:      w.dir,
		maxSize:  w.maxSize,
	}

	if err := wt.initFromExisting(); err != nil {
		return nil, err
	}
	return wt, nil
}

func (wt *WalThread) initFromExisting() error {
	// 查找该线程对应的WAL文件，格式: wal-{threadId}-*.log
	pattern := fmt.Sprintf("wal-%d-*.log", wt.threadId)
	matches, _ := filepath.Glob(filepath.Join(wt.dir, pattern))
	if len(matches) == 0 {
		return wt.rotateLocked()
	}
	sort.Strings(matches)
	last := matches[len(matches)-1]
	base := filepath.Base(last)
	// 解析文件名 wal-{threadId}-{seq}.log
	parts := strings.Split(strings.TrimSuffix(base, ".log"), "-")
	if len(parts) != 3 {
		return fmt.Errorf("invalid WAL filename format: %s", base)
	}
	seqStr := parts[2]
	seq, _ := strconv.Atoi(seqStr)
	wt.seq = seq

	f, err := os.OpenFile(last, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	wt.curFile = f
	wt.writer = bufio.NewWriter(f)
	fi, _ := f.Stat()
	wt.curSize = fi.Size()
	return nil
}

func (wt *WalThread) rotateLocked() error {
	if wt.curFile != nil {
		_ = wt.writer.Flush()
		_ = wt.curFile.Sync()
		_ = wt.curFile.Close()
	}
	wt.seq++
	name := fmt.Sprintf("wal-%d-%06d.log", wt.threadId, wt.seq)
	path := filepath.Join(wt.dir, name)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	wt.curFile = f
	wt.writer = bufio.NewWriter(f)
	wt.curSize = 0
	return nil
}

func (wt *WalThread) appendRecord(fileId uint64, ts int64) error {
	// payload = fileId(8) + ts(8)
	payload := make([]byte, 16)
	binary.LittleEndian.PutUint64(payload[0:8], fileId)
	binary.LittleEndian.PutUint64(payload[8:16], uint64(ts))

	crc := crc32.ChecksumIEEE(payload)
	var header [8]byte
	binary.LittleEndian.PutUint32(header[0:4], uint32(len(payload)))
	binary.LittleEndian.PutUint32(header[4:8], crc)

	recordSize := int64(len(payload) + len(header))

	wt.mu.Lock()
	defer wt.mu.Unlock()

	if wt.curFile == nil {
		if err := wt.rotateLocked(); err != nil {
			return err
		}
	}
	// 按大小滚动
	if wt.curSize+recordSize > wt.maxSize {
		if err := wt.rotateLocked(); err != nil {
			return err
		}
	}
	// 顺序写入 + always fsync
	if _, err := wt.writer.Write(header[:]); err != nil {
		return err
	}
	if _, err := wt.writer.Write(payload); err != nil {
		return err
	}
	wt.curSize += recordSize

	if err := wt.writer.Flush(); err != nil {
		return err
	}
	return wt.curFile.Sync()
}

func (wt *WalThread) close() error {
	wt.mu.Lock()
	defer wt.mu.Unlock()
	if wt.curFile == nil {
		return nil
	}
	_ = wt.writer.Flush()
	_ = wt.curFile.Sync()
	return wt.curFile.Close()
}

// Append 使用随机分配方式选择线程进行写入，实现负载均衡
func (w *Wal) Append(fileId uint64, ts int64) error {
	// 使用随机分配方式选择线程，实现更好的负载均衡
	// 由于WAL只保存ID和时间戳，同一ID放到不同文件无所谓
	seed := atomic.AddUint64(&w.randomSeed, 1)
	threadId := int(seed % config.WalThreads)
	return w.threads[threadId].appendRecord(fileId, ts)
}

func (w *Wal) Close() error {
	close(w.shutdown)
	w.wg.Wait()

	var lastErr error
	for _, thread := range w.threads {
		if err := thread.close(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// ReplayAll 多线程回放所有WAL文件
func (w *Wal) ReplayAll(minTs int64, apply func(fileId uint64, ts int64) error) error {
	// 收集所有WAL文件
	var allFiles []string
	for i := 0; i < config.WalThreads; i++ {
		pattern := fmt.Sprintf("wal-%d-*.log", i)
		matches, _ := filepath.Glob(filepath.Join(w.dir, pattern))
		allFiles = append(allFiles, matches...)
	}
	sort.Strings(allFiles)

	// 使用工作池进行并发回放
	fileChan := make(chan string, len(allFiles))
	errorChan := make(chan error, config.WalThreads)
	var wg sync.WaitGroup

	// 启动工作协程
	for i := 0; i < config.WalThreads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filePath := range fileChan {
				if err := w.replayOne(filePath, minTs, apply); err != nil {
					errorChan <- fmt.Errorf("failed to replay %s: %w", filePath, err)
					return
				}
			}
		}()
	}

	// 发送文件到工作协程
	for _, filePath := range allFiles {
		fileChan <- filePath
	}
	close(fileChan)

	// 等待所有工作协程完成
	wg.Wait()
	close(errorChan)

	// 检查是否有错误
	for err := range errorChan {
		return err
	}

	return nil
}

func (w *Wal) replayOne(path string, minTs int64, apply func(fileId uint64, ts int64) error) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	r := bufio.NewReader(f)
	for {
		var header [8]byte
		if _, err := io.ReadFull(r, header[:]); err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				return nil
			}
			return err
		}
		length := binary.LittleEndian.Uint32(header[0:4])
		sum := binary.LittleEndian.Uint32(header[4:8])

		if length != 16 {
			return fmt.Errorf("bad wal record length in %s", path)
		}
		data := make([]byte, length)
		if _, err := io.ReadFull(r, data); err != nil {
			if errors.Is(err, io.ErrUnexpectedEOF) {
				return nil // 尾部半条忽略
			}
			return err
		}
		if crc32.ChecksumIEEE(data) != sum {
			return nil // 数据损坏，停止回放该文件
		}

		fileId := binary.LittleEndian.Uint64(data[0:8])
		ts := int64(binary.LittleEndian.Uint64(data[8:16]))

		if ts > minTs {
			if err := apply(fileId, ts); err != nil {
				return err
			}
		}
	}
}
