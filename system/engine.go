package system

import (
	"context"
	"fileClick/config"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

var RankEngine *Engine

type Engine struct {
	mu        sync.Mutex
	rankBoard *RankBoard

	wal *Wal
	rdb *Rdb

	snapInterval time.Duration
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
}

func NewEngine() (*Engine, error) {
	wal, err := NewWAL()
	if err != nil {
		return nil, err
	}
	rdb := NewRDB()

	e := &Engine{
		rankBoard:    GetRankBoard(),
		wal:          wal,
		rdb:          rdb,
		snapInterval: config.RdbShotEvery,
	}
	e.ctx, e.cancel = context.WithCancel(context.Background())
	return e, nil
}

func (e *Engine) Recover() error {
	ld, err := e.rdb.LoadLatest()
	if err != nil {
		return err
	}
	e.mu.Lock()
	// 恢复数据
	fileMap := make(map[uint64]*File, len(ld.Files))
	for _, file := range ld.Files {
		fileMap[file.Id] = file
		e.rankBoard.rbt.insert(file)
	}
	e.rankBoard.files = fileMap
	e.mu.Unlock()

	apply := func(fileId uint64, ts int64) error {
		e.rankBoard.writeCh <- &HitEvent{Id: fileId}
		return nil
	}

	if err = e.wal.ReplayAll(ld.SnapshotTs, apply); err != nil {
		return err
	}
	return nil
}

func (e *Engine) Click(fileId uint64) error {
	ts := time.Now().Unix()
	if err := e.wal.Append(fileId, ts); err != nil {
		return err
	}
	e.rankBoard.writeCh <- &HitEvent{
		Id: fileId,
	}
	return nil
}

func (e *Engine) TopN(n int) []*File {
	return e.rankBoard.rbt.TopN(n)
}

func (e *Engine) TopAll() []*File {
	return e.rankBoard.rbt.GetAllNodesDesc()
}

// StartScheduler 周期快照 & AOF 清理
func (e *Engine) StartScheduler() {
	if e.snapInterval <= 0 {
		return
	}
	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		tk := time.NewTicker(e.snapInterval)
		defer tk.Stop()
		for {
			select {
			case <-e.ctx.Done():
				return
			case <-tk.C:
				e.doSnapshotAndPrune()
			}
		}
	}()
}

func (e *Engine) Stop() {
	e.cancel()
	e.wg.Wait()
	_ = e.wal.Close()
	// 退出前再做一次快照
	e.doSnapshotAndPrune()
}

func (e *Engine) doSnapshotAndPrune() {
	// 1) 保存 RDB
	data := e.TopAll() // 拷贝快照态
	snapTs, _, err := e.rdb.Save(data)
	if err != nil {
		fmt.Println("RDB save failed:", err)
		return
	}

	// 2) 删除早于 snapTs 的 WAL 段
	_ = pruneOldWAL(e.wal.dir, snapTs)
}

func pruneOldWAL(dir string, snapTs int64) error {
	// 新的WAL文件格式: wal-{threadId}-{seq}.log
	files, _ := filepath.Glob(filepath.Join(dir, "wal-*-*.log"))
	sort.Strings(files)
	for _, p := range files {
		// 简单策略：如果文件的 mtime < snapTs，就删掉
		fi, err := os.Stat(p)
		if err != nil {
			continue
		}
		if fi.ModTime().Unix() < snapTs {
			_ = os.Remove(p)
		}
	}
	return nil
}
