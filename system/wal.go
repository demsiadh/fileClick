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
)

// Wal wal文件，类比AOF，格式[4]len | [4]crc32 | [8]fileId | [8]ts | [2]nameLen | [n]name
type Wal struct {
	dir     string
	maxSize int64

	mu      sync.Mutex
	curFile *os.File
	writer  *bufio.Writer
	curSize int64
	seq     int
}

func NewWAL() (*Wal, error) {
	w := &Wal{dir: config.WalPath, maxSize: config.WalMaxSize}
	if err := w.initFromExisting(); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *Wal) initFromExisting() error {
	// 找已经存在的 wal-*.log，确定起始 seq；打开最后一个以追加
	matches, _ := filepath.Glob(filepath.Join(w.dir, "wal-*.log"))
	if len(matches) == 0 {
		return w.rotateLocked()
	}
	sort.Strings(matches)
	last := matches[len(matches)-1]
	base := filepath.Base(last)
	seqStr := strings.TrimSuffix(strings.TrimPrefix(base, "wal-"), ".log")
	seq, _ := strconv.Atoi(seqStr)
	w.seq = seq

	f, err := os.OpenFile(last, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	w.curFile = f
	w.writer = bufio.NewWriter(f)
	fi, _ := f.Stat()
	w.curSize = fi.Size()
	return nil
}

func (w *Wal) rotateLocked() error {
	if w.curFile != nil {
		_ = w.writer.Flush()
		_ = w.curFile.Sync()
		_ = w.curFile.Close()
	}
	w.seq++
	name := fmt.Sprintf("wal-%06d.log", w.seq)
	path := filepath.Join(w.dir, name)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	w.curFile = f
	w.writer = bufio.NewWriter(f)
	w.curSize = 0
	return nil
}

func (w *Wal) Append(fileId uint64, ts int64, fileName string) error {
	nameBytes := []byte(fileName)
	payload := make([]byte, 8+8+2+len(nameBytes)) // id(8) + ts(8) + namelen(2) + name
	binary.LittleEndian.PutUint64(payload[0:8], fileId)
	binary.LittleEndian.PutUint64(payload[8:16], uint64(ts))
	binary.LittleEndian.PutUint16(payload[16:18], uint16(len(nameBytes)))
	copy(payload[18:], nameBytes)

	crc := crc32.ChecksumIEEE(payload)
	var header [8]byte
	binary.LittleEndian.PutUint32(header[0:4], uint32(len(payload)))
	binary.LittleEndian.PutUint32(header[4:8], crc)

	recordSize := int64(len(payload) + len(header))

	w.mu.Lock()
	defer w.mu.Unlock()

	if w.curFile == nil {
		if err := w.rotateLocked(); err != nil {
			return err
		}
	}
	// 按大小滚动
	if w.curSize+recordSize > w.maxSize {
		if err := w.rotateLocked(); err != nil {
			return err
		}
	}
	// 顺序写入 + always fsync
	if _, err := w.writer.Write(header[:]); err != nil {
		return err
	}
	if _, err := w.writer.Write(payload); err != nil {
		return err
	}
	w.curSize += recordSize

	if err := w.writer.Flush(); err != nil {
		return err
	}
	return w.curFile.Sync()
}

func (w *Wal) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.curFile == nil {
		return nil
	}
	_ = w.writer.Flush()
	_ = w.curFile.Sync()
	return w.curFile.Close()
}

// ReplayAll 从目录中按顺序回放全部 wal（遇到损坏或尾部半条自动停止该文件），apply 只会收到 ts>minTs 的记录
func (w *Wal) ReplayAll(minTs int64, apply func(fileId uint64, ts int64, fileName string) error) error {
	matches, _ := filepath.Glob(filepath.Join(w.dir, "wal-*.log"))
	sort.Strings(matches)
	for _, p := range matches {
		if err := w.replayOne(p, minTs, apply); err != nil {
			return err
		}
	}
	return nil
}

func (w *Wal) replayOne(path string, minTs int64, apply func(fileId uint64, ts int64, fileName string) error) error {
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

		if length == 0 || length > 1<<26 { // 粗略防御：单条 >64MB 视为异常
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
			// 数据损坏：停止该文件后续回放（与 Redis 类似）
			return nil
		}
		fileId := binary.LittleEndian.Uint64(data[0:8])
		ts := int64(binary.LittleEndian.Uint64(data[8:16]))
		nameLen := binary.LittleEndian.Uint16(data[16:18])
		if int(18+nameLen) > len(data) {
			return nil // 保护：异常长度当作截断
		}
		fileName := string(data[18 : 18+nameLen])

		if ts > minTs {
			if err := apply(fileId, ts, fileName); err != nil {
				return err
			}
		}
	}
}
