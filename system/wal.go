package system

import (
	"bufio"
	"fileClick/config"
	"os"
	"sync"
)

var walOnce sync.Once
var wal *Wal

type Wal struct {
	file   *os.File
	writer *bufio.Writer
}

func GetWal() *Wal {
	walOnce.Do(func() {
		f, err := os.OpenFile(config.WalPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
		wal = &Wal{
			file:   f,
			writer: bufio.NewWriter(f),
		}
	})
	return wal
}

func (w *Wal) Append(event HitEvent) error {
	panic("implement me")
}
