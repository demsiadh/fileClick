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
	files       map[int]*File
	idGenerator *util.IDGenerator
	writeCh     chan HitEvent
	wg          sync.WaitGroup
	rbt         *RedBlackTree
}

func GetRankBoard() *RankBoard {
	rankBoardOnce.Do(func() {
		rankBoard = &RankBoard{
			files:       make(map[int]*File),
			idGenerator: util.GetIdGenerator(),
			writeCh:     make(chan HitEvent, 1000),
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

func (rb *RankBoard) RecordHit(event HitEvent) {
	// 校验文件是否已经存在排行榜
	file, exists := rb.files[event.Id]
	if !exists {
		// 不存在则初始化
		id := rankBoard.idGenerator.GenerateID()
		file = &File{
			Id:       id,
			FileName: event.FileName,
			Count:    1,
		}
		rb.files[id] = file
		rb.rbt.insert(file)
	} else {
		rb.rbt.update(file)
	}
}
