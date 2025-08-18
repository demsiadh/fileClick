package config

import (
	"os"
	"time"
)

const (
	WalPath       = "data/system/wal/"
	WalMaxSize    = 64 << 20
	WalThreads    = 4
	RdbMaxFileNum = 3
	RdbPath       = "data/system/rdb/"
	RdbShotEvery  = time.Minute * 5
	FilePath      = "data/files/"
	FileInfoPath  = "data/fileInfo.json"
	FileMaxSize   = 32 << 20
	LogPath       = "data/logs/"
	FileEventMax  = 10000
)

func init() {
	var err error
	err = os.MkdirAll(WalPath, os.ModePerm)
	if err != nil {
		panic("mkdir failed")
	}
	err = os.MkdirAll(RdbPath, os.ModePerm)
	if err != nil {
		panic("mkdir failed")
	}
	err = os.MkdirAll(FilePath, os.ModePerm)
	if err != nil {
		panic("mkdir failed")
	}
}
