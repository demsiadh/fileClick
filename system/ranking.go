package system

import (
	"fileClick/util"
	"sync"
)

const (
	red   = "red"
	black = "black"
)

var rankBoardOnce sync.Once
var rankBoard *RankBoard

type RankBoard struct {
	files       map[uint64]*File
	idGenerator *util.IDGenerator
	writeCh     chan *HitEvent
	wg          sync.WaitGroup
	rbt         *RedBlackTree
}

func GetRankBoard() *RankBoard {
	rankBoardOnce.Do(func() {
		rankBoard = &RankBoard{
			files:       make(map[uint64]*File),
			idGenerator: util.GetIdGenerator(),
			writeCh:     make(chan *HitEvent, 1000),
			rbt:         &RedBlackTree{},
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
		rb.RecordHit(event)
	}
}

func (rb *RankBoard) RecordHit(event *HitEvent) {
	// 校验文件是否已经存在排行榜
	file, exists := rb.files[event.Id]
	if !exists {
		file = &File{
			Id:       event.Id,
			FileName: event.FileName,
			Count:    1,
		}
		rb.files[event.Id] = file
		rb.rbt.insert(file)
	} else {
		file.Count++
		rb.rbt.update(file)
	}
}
