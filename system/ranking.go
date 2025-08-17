package system

import (
	"fileClick/config"
	"sync"
)

var rankBoardOnce sync.Once
var rankBoard *RankBoard

type RankBoard struct {
	writeCh chan *FileEvent
	wg      sync.WaitGroup
	lru     *LRUList
}

func GetRankBoard() *RankBoard {
	rankBoardOnce.Do(func() {
		rankBoard = &RankBoard{
			writeCh: make(chan *FileEvent, config.FileEventMax),
			lru:     NewLRURanking(),
		}
		go rankBoard.worker()
	})
	return rankBoard
}

// 启动工作线程
func (rb *RankBoard) worker() {
	rb.wg.Add(1)
	defer rb.wg.Done()

	for event := range rb.writeCh {
		rb.lru.hit(event.Id)
	}
}
