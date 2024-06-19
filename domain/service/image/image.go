package image

import (
	"context"
	"errors"
	"scrapper/domain/entity"
	imageRepo "scrapper/domain/repository/image"
	logger "scrapper/infrastructure/log"
	"scrapper/utils/image"
	"time"
)

const maxCreateWorkerCount = 1000
const workerQueueLength = 20000

var ErrServiceUnavailable = errors.New("service unavailable")

type Service struct {
	imageRepo        imageRepo.Image
	logger           logger.Logger
	createQueue      chan *entity.Image
	storageDirectory string
}

func NewService(logger logger.Logger, imageRepo imageRepo.Image, storageDirectory string) *Service {
	s := &Service{
		imageRepo:        imageRepo,
		logger:           logger,
		createQueue:      make(chan *entity.Image, workerQueueLength),
		storageDirectory: storageDirectory,
	}
	s.startWorkers()
	return s
}

func (s Service) startWorkers() {
	for i := 0; i < maxCreateWorkerCount; i++ {
		go s.createWorker()
	}
}

func (s Service) createWorker() {
	imageBatch := make([]*entity.Image, 0)
	t := time.NewTicker(time.Second)
	for {
		select {
		case image, ok := <-s.createQueue:
			if !ok {
				return
			}
			imageBatch = append(imageBatch, image)
		case <-t.C:
			if len(imageBatch) > 0 {
				err := s.imageRepo.CreateBatch(context.Background(), imageBatch)
				if err != nil {
					s.logger.Error(err)
				}
				imageBatch = make([]*entity.Image, 0)
			}
		}
	}

}

func (s Service) Create(downloader image.Downloader, done chan bool) {
	filePathsChan := make(chan string, 100)
	go downloader.Download(filePathsChan)
	go func() {
		for path := range filePathsChan {
			s.createQueue <- &entity.Image{
				File: path,
			}
		}
		done <- true
	}()
}

// reading in chunks
func (s Service) Read(targetCount uint64, ch chan *entity.Image) {
	defer close(ch)
	var count, offset uint64
	for {
		ctx, _ := context.WithTimeout(context.Background(), time.Millisecond*100)
		images, err := s.imageRepo.List(ctx, offset)
		if err != nil {
			s.logger.Error(err)
		}
		if len(images) < 1 {
			offset = 0 //circular reading...
		}
		for _, img := range images {
			ch <- img
			count++
			if count >= targetCount {
				return
			}
		}
		offset += 10
	}
}
