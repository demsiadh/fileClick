package config

import (
	"os"
	"time"
)

const (
	WalPath      = "data/system/wal/"
	WalMaxSize   = 64 << 20
	RdbPath      = "data/system/rdb/"
	RdbShotEvery = time.Minute * 5
	FilePath     = "data/files/"
	FileInfoPath = "data/fileInfo.json"
	WalThreads   = 5
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
