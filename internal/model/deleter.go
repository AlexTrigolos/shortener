package model

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type DeleterURLs interface {
	DeleteURLs(ctx context.Context, urls []URLsData) error
}

type URLsData struct {
	Urls   []string
	UserID string
}

type Deleter struct {
	storeURLs DeleterURLs
	delChan   chan URLsData
	logger    *zap.Logger
}

func NewDeleter(doneCh chan struct{}, delURLs DeleterURLs, logger *zap.Logger) (*Deleter, chan struct{}) {
	finishChan := make(chan struct{})
	deleter := &Deleter{storeURLs: delURLs, delChan: make(chan URLsData, 1024), logger: logger}
	go deleter.flushURLS(doneCh, finishChan)
	return deleter, finishChan
}

func (d *Deleter) AddURLs(urls []string, userID string) {
	d.delChan <- URLsData{Urls: append([]string(nil), urls...), UserID: userID}
}

func (d *Deleter) flushURLS(doneCh chan struct{}, finishChan chan struct{}) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	var urlsData []URLsData

	for {
		select {
		case <-doneCh:
			d.deleteURLs(urlsData)
			close(finishChan)
			close(d.delChan)
			return
		case buffURLs := <-d.delChan:
			urlsData = append(urlsData, buffURLs)
			if len(urlsData) >= 100 {
				d.deleteURLs(urlsData)
			}
		case <-ticker.C:
			d.deleteURLs(urlsData)
		}
	}
}

func (d *Deleter) deleteURLs(urlsData []URLsData) {
	if len(urlsData) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := d.storeURLs.DeleteURLs(ctx, urlsData)
	if err != nil {
		d.logger.Error("Не смогли сохранить удаленные", zap.Error(err))
		return
	}
	urlsData = nil
}
