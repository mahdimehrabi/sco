package image

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"m1-article-service/domain/entity"
	mock_log "m1-article-service/mock/infrastructure"
	mock_image "m1-article-service/mock/repository"
	mock_utils "m1-article-service/mock/utils"
	"testing"
	"time"
)

func TestService_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})
	err := errors.New("error")
	downloadCount := 10

	var tests = []struct {
		name           string
		Image          *entity.Image
		loggerMock     func() *mock_log.MockLog
		ImageRepoMock  func() *mock_image.MockImage
		downloaderMock func() *mock_utils.MockDownloader
		ctx            context.Context
	}{
		{
			name: "success",
			loggerMock: func() *mock_log.MockLog {
				loggerInfra := mock_log.NewMockLog(ctrl)
				return loggerInfra
			},
			ImageRepoMock: func() *mock_image.MockImage {
				repoLogMock := mock_image.NewMockImage(ctrl)
				repoLogMock.EXPECT().CreateBatch(gomock.Any(), gomock.Any()).Times(downloadCount).Return(nil)
				return repoLogMock
			},
			downloaderMock: func() *mock_utils.MockDownloader {
				downloaderMock := mock_utils.NewMockDownloader(ctrl)
				downloaderMock.EXPECT().Download(gomock.Any()).Do(func(ch chan string) {
					for i := 0; i < downloadCount; i++ {
						ch <- fmt.Sprintf("%d", i)
					}
					close(ch)
				}).MaxTimes(1)
				return downloaderMock
			},
			ctx: context.Background(),
		},
		{
			name: "RepoError",
			loggerMock: func() *mock_log.MockLog {
				loggerInfra := mock_log.NewMockLog(ctrl)
				loggerInfra.EXPECT().Error(err).MinTimes(10).Return()
				return loggerInfra
			},
			ImageRepoMock: func() *mock_image.MockImage {
				repoLogMock := mock_image.NewMockImage(ctrl)
				repoLogMock.EXPECT().CreateBatch(gomock.Any(), gomock.Any()).AnyTimes().Return(err)
				return repoLogMock
			},
			downloaderMock: func() *mock_utils.MockDownloader {
				downloaderMock := mock_utils.NewMockDownloader(ctrl)
				downloaderMock.EXPECT().Download(gomock.Any()).Do(func(ch chan string) {
					for i := 0; i < downloadCount; i++ {
						ch <- fmt.Sprintf("%d", i)
					}
					close(ch)
				}).MaxTimes(1)
				return downloaderMock
			},
			ctx: context.Background(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logRepoMock := test.ImageRepoMock()
			loggerMock := test.loggerMock()
			downloaderMock := test.downloaderMock()

			service := NewService(loggerMock, logRepoMock, "")
			ch := make(chan bool)
			service.Create(downloaderMock, ch)
			<-ch
			time.Sleep(2 * time.Second)
			loggerMock.EXPECT()
			logRepoMock.EXPECT()
		})
	}
}

func BenchmarkService_Create(b *testing.B) {
	ctrl := gomock.NewController(b)
	downloadCount := 100
	downloaderMock := mock_utils.NewMockDownloader(ctrl)
	downloaderMock.EXPECT().Download(gomock.Any()).Do(func(ch chan string) {
		for i := 0; i < downloadCount; i++ {
			ch <- fmt.Sprintf("%d", i)
		}
		close(ch)
	}).MaxTimes(1)

	repoImageMock := mock_image.NewMockImage(ctrl)
	repoImageMock.EXPECT().CreateBatch(gomock.Any(), gomock.Any()).Times(downloadCount).Return(nil)

	loggerMock := mock_log.NewMockLog(ctrl)
	b.ResetTimer()
	service := NewService(loggerMock, repoImageMock, "/")
	done := make(chan bool)
	service.Create(downloaderMock, done)
	<-done
	fmt.Println("create method:", b.Elapsed())
	if b.Elapsed() > 10*time.Millisecond {
		b.Error("Image service-create takes too long to run")
	}
	time.Sleep(2 * time.Second)
	loggerMock.EXPECT()
	repoImageMock.EXPECT()
}

func TestService_Read(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})
	err := errors.New("error")

	var tests = []struct {
		name           string
		Image          *entity.Image
		loggerMock     func() *mock_log.MockLog
		ImageRepoMock  func() *mock_image.MockImage
		downloaderMock func() *mock_utils.MockDownloader
		error          error
		ctx            context.Context
		count          uint64
		mustDoneCount  uint64
	}{
		{
			name: "success",
			loggerMock: func() *mock_log.MockLog {
				loggerInfra := mock_log.NewMockLog(ctrl)
				return loggerInfra
			},
			ImageRepoMock: func() *mock_image.MockImage {
				repoLogMock := mock_image.NewMockImage(ctrl)
				repoLogMock.EXPECT().List(gomock.Any(), gomock.Any()).AnyTimes().Return([]*entity.Image{
					{"1"},
					{"2"},
					{"3"},
					{"4"},
					{"5"},
					{"6"},
					{"7"},
					{"8"},
					{"9"},
					{"10"},
				}, nil)
				return repoLogMock
			},
			error:         nil,
			ctx:           context.Background(),
			count:         10,
			mustDoneCount: 10,
		},
		{
			name: "RepoError",
			loggerMock: func() *mock_log.MockLog {
				loggerInfra := mock_log.NewMockLog(ctrl)
				loggerInfra.EXPECT().Error(err).MinTimes(100).Return()
				return loggerInfra
			},
			ImageRepoMock: func() *mock_image.MockImage {
				repoLogMock := mock_image.NewMockImage(ctrl)
				repoLogMock.EXPECT().List(gomock.Any(), gomock.Any()).AnyTimes().Return(nil, err)
				return repoLogMock
			},
			error:         err,
			ctx:           context.Background(),
			count:         uint64(10),
			mustDoneCount: uint64(0),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			imageRepoMock := test.ImageRepoMock()
			loggerMock := test.loggerMock()

			service := NewService(loggerMock, imageRepoMock, "")
			images := make(chan *entity.Image, 10)
			go service.Read(test.count, images)
			var count uint64
			//considering a deadline in case of errors in repository
			tick := time.NewTicker(time.Second * 2)

			done := false
			for !done {
				select {
				case _, ok := <-images:
					if !ok {
						done = true
						break
					}
					count++
				case <-tick.C:
					done = true
				}
			}
			if count != test.mustDoneCount {
				t.Errorf("count:%d is not equal to:%d", count, test.mustDoneCount)
			}
			loggerMock.EXPECT()
			imageRepoMock.EXPECT()
		})
	}
}

func BenchmarkService_Read(b *testing.B) {
	ctrl := gomock.NewController(b)

	repoImageMock := mock_image.NewMockImage(ctrl)
	repoImageMock.EXPECT().List(gomock.Any(), gomock.Any()).
		AnyTimes().Return([]*entity.Image{
		{"1"},
		{"2"},
		{"3"},
		{"4"},
		{"5"},
		{"6"},
		{"7"},
		{"8"},
		{"9"},
		{"10"},
	}, nil)

	loggerMock := mock_log.NewMockLog(ctrl)
	b.ResetTimer()
	service := NewService(loggerMock, repoImageMock, "")
	images := make(chan *entity.Image, 10)
	mustDoneCount := uint64(1000)
	go service.Read(mustDoneCount, images)

	var count uint64
	for image := range images {
		image = image //aviod unused error
		count++
	}

	if count != mustDoneCount {
		b.Errorf("count:%d is not equal to:%d", count, mustDoneCount)
	}
	fmt.Println("read method:", b.Elapsed())
	if b.Elapsed() > 10*time.Millisecond {
		b.Error("Image service-create takes too long to run")
	}
	time.Sleep(2 * time.Second)
	loggerMock.EXPECT()
	repoImageMock.EXPECT()
}
