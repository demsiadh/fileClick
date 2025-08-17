package config

import (
	"os"
	"time"
)

const (
	WalPath       = "data/system/wal/"
	WalMaxSize    = 64 << 20
	RdbMaxFileNum = 3
	RdbPath       = "data/system/rdb/"
	RdbShotEvery  = time.Minute * 5
	FilePath      = "data/files/"
	FileInfoPath  = "data/fileInfo.json"
	FileMaxSize   = 32 << 20
	WalThreads    = 5
	LogPath       = "data/logs/"
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
	err = os.MkdirAll(LogPath, os.ModePerm)
	if err != nil {
		panic("mkdir failed")
	}
}
