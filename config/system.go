package config

import (
	"time"
)

const (
	WalPath      = "data/system/wal"
	WalMaxSize   = 64 << 20
	RdbPath      = "data/system/rdb"
	RdbShotEvery = time.Minute * 5
)
