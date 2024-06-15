// Code generated by MockGen. DO NOT EDIT.
// Source: ./utils/image/downloader.go

// Package mock_image is a generated GoMock package.
package mock_image

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockDownloader is a mock of Downloader interface.
type MockDownloader struct {
	ctrl     *gomock.Controller
	recorder *MockDownloaderMockRecorder
}

// MockDownloaderMockRecorder is the mock recorder for MockDownloader.
type MockDownloaderMockRecorder struct {
	mock *MockDownloader
}

// NewMockDownloader creates a new mock instance.
func NewMockDownloader(ctrl *gomock.Controller) *MockDownloader {
	mock := &MockDownloader{ctrl: ctrl}
	mock.recorder = &MockDownloaderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDownloader) EXPECT() *MockDownloaderMockRecorder {
	return m.recorder
}

// Download mocks base method.
func (m *MockDownloader) Download(pathsChan chan string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Download", pathsChan)
}

// Download indicates an expected call of Download.
func (mr *MockDownloaderMockRecorder) Download(pathsChan interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Download", reflect.TypeOf((*MockDownloader)(nil).Download), pathsChan)
}
