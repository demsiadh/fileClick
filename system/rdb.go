package system

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fileClick/config"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Rdb 快照
type Rdb struct {
	dir     string
	version uint16
}

func NewRDB() *Rdb {
	dir := config.RdbPath
	_ = os.MkdirAll(dir, 0755)
	return &Rdb{dir: dir, version: 1}
}

// Save 保存RDB
func (r *Rdb) Save(files []*File) (snapshotTs int64, path string, err error) {
	tmpPath := filepath.Join(config.RdbPath, fmt.Sprintf("dump-%d.rdb.tmp", time.Now().UnixNano()))
	finalTs := time.Now().Unix()
	finalPath := filepath.Join(config.RdbPath, fmt.Sprintf("dump-%d.rdb", finalTs))

	f, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return 0, "", err
	}
	defer func() {
		if err != nil {
			_ = f.Close()
			_ = os.Remove(tmpPath)
		}
	}()

	w := bufio.NewWriter(f)
	var body bytes.Buffer

	// 头部 magic(4) + version(2) + snapshotTs(8)
	body.WriteString("RDB1")
	_ = binary.Write(&body, binary.LittleEndian, r.version)
	_ = binary.Write(&body, binary.LittleEndian, finalTs)

	// Count
	n := uint32(len(files))
	_ = binary.Write(&body, binary.LittleEndian, n)

	// Entries: [id(8) | count(8) | nameLen(2) | name(n)]
	for _, f := range files {
		_ = binary.Write(&body, binary.LittleEndian, f.Id)
		_ = binary.Write(&body, binary.LittleEndian, f.Count)
		nameBytes := []byte(f.FileName)
		_ = binary.Write(&body, binary.LittleEndian, uint16(len(nameBytes)))
		_, _ = body.Write(nameBytes)
	}

	// CRC
	crc := crc32.ChecksumIEEE(body.Bytes())
	if _, err = w.Write(body.Bytes()); err != nil {
		return 0, "", err
	}
	var crcBuf [4]byte
	binary.LittleEndian.PutUint32(crcBuf[:], crc)
	if _, err = w.Write(crcBuf[:]); err != nil {
		return 0, "", err
	}
	if err = w.Flush(); err != nil {
		return 0, "", err
	}
	if err = f.Sync(); err != nil {
		return 0, "", err
	}
	if err = f.Close(); err != nil {
		return 0, "", err
	}
	if err = os.Rename(tmpPath, finalPath); err != nil {
		return 0, "", err
	}

	matches, _ := filepath.Glob(filepath.Join(config.RdbPath, "dump-*.rdb"))
	if len(matches) > 3 {
		sort.Strings(matches)
		for _, old := range matches[:len(matches)-3] {
			_ = os.Remove(old)
		}
	}

	return finalTs, finalPath, nil
}

type RdbLoadResult struct {
	SnapshotTs int64
	Files      []*File
	Path       string
}

// LoadLatest 加载RDB文件
func (r *Rdb) LoadLatest() (*RdbLoadResult, error) {
	matches, _ := filepath.Glob(filepath.Join(config.RdbPath, "dump-*.rdb"))
	if len(matches) == 0 {
		return &RdbLoadResult{SnapshotTs: 0, Files: []*File{}, Path: ""}, nil
	}
	sort.Strings(matches)
	path := matches[len(matches)-1]

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(b) < 4+2+8+4+4 { // 最小长度保护
		return nil, fmt.Errorf("rdb too small: %s", path)
	}
	if string(b[0:4]) != "RDB1" {
		return nil, fmt.Errorf("bad rdb magic: %s", path)
	}
	// 校验 CRC
	crcStored := binary.LittleEndian.Uint32(b[len(b)-4:])
	if crc32.ChecksumIEEE(b[:len(b)-4]) != crcStored {
		return nil, fmt.Errorf("rdb crc mismatch: %s", path)
	}

	reader := bytes.NewReader(b[4 : len(b)-4])
	var ver uint16
	var ts int64
	_ = binary.Read(reader, binary.LittleEndian, &ver)
	_ = binary.Read(reader, binary.LittleEndian, &ts)
	var n uint32
	_ = binary.Read(reader, binary.LittleEndian, &n)
	out := make([]*File, n)
	for i := uint32(0); i < n; i++ {
		var id, cnt uint64
		var nameLen uint16
		_ = binary.Read(reader, binary.LittleEndian, &id)
		_ = binary.Read(reader, binary.LittleEndian, &cnt)
		_ = binary.Read(reader, binary.LittleEndian, &nameLen)
		name := make([]byte, nameLen)
		if _, err := io.ReadFull(reader, name); err != nil {
			return nil, err
		}
		out[i] = &File{
			Id:       id,
			Count:    cnt,
			FileName: string(name),
		}
	}
	return &RdbLoadResult{
		SnapshotTs: ts,
		Files:      out,
		Path:       path,
	}, nil
}
