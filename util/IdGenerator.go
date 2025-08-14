package util

import (
	"sync"
	"time"
)

var IdGenerator *IDGenerator
var once sync.Once

// IDGenerator ID生成器结构体
type IDGenerator struct {
	lastTimestamp int64 // 上次生成ID的时间戳
	sequence      int   // 当前序列号
	sequenceMask  int   // 序列号最大值（最多支持每毫秒生成1000个ID）
}

const (
	epoch        = 1617113720000 // 自定义起始时间戳（2021年3月上旬）
	sequenceBits = 10            // 序列号的位数（最大支持每毫秒生成1024个ID）
)

func GetIdGenerator() *IDGenerator {
	once.Do(func() {
		IdGenerator = &IDGenerator{
			lastTimestamp: 0,
			sequence:      0,
			sequenceMask:  (1 << sequenceBits) - 1,
		}
	})
	return IdGenerator
}

// GenerateID generateID 生成唯一ID
func (g *IDGenerator) GenerateID() int {

	// 获取当前时间戳（毫秒）
	timestamp := time.Now().UnixMilli()

	if timestamp == g.lastTimestamp {
		// 当前毫秒内生成多个ID，递增序列号
		g.sequence = (g.sequence + 1) & g.sequenceMask
		if g.sequence == 0 {
			// 序列号达到最大值，等待下一毫秒
			for timestamp <= g.lastTimestamp {
				timestamp = time.Now().UnixMilli()
			}
		}
	} else {
		// 时间戳变化，重置序列号
		g.sequence = 0
	}

	// 更新最后的时间戳
	g.lastTimestamp = timestamp

	// 将时间戳左移，并与序列号拼接生成最终的ID（确保不会超出int范围）
	id := int((timestamp-epoch)<<sequenceBits) | int(g.sequence)
	return id
}
