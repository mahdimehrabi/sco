// Code generated by MockGen. DO NOT EDIT.
// Source: ./domain/repository/image/image.go

// Package mock_image is a generated GoMock package.
package mock_image

import (
	context "context"
	"m1-article-service/domain/entity"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockImage is a mock of Image interface.
type MockImage struct {
	ctrl     *gomock.Controller
	recorder *MockImageMockRecorder
}

// MockImageMockRecorder is the mock recorder for MockImage.
type MockImageMockRecorder struct {
	mock *MockImage
}

// NewMockImage creates a new mock instance.
func NewMockImage(ctrl *gomock.Controller) *MockImage {
	mock := &MockImage{ctrl: ctrl}
	mock.recorder = &MockImageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockImage) EXPECT() *MockImageMockRecorder {
	return m.recorder
}

// CreateBatch mocks base method.
func (m *MockImage) CreateBatch(arg0 context.Context, arg1 []*entity.Image) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateBatch", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateBatch indicates an expected call of CreateBatch.
func (mr *MockImageMockRecorder) CreateBatch(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateBatch", reflect.TypeOf((*MockImage)(nil).CreateBatch), arg0, arg1)
}

// List mocks base method.
func (m *MockImage) List(arg0 context.Context, arg1 uint64) ([]*entity.Image, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", arg0, arg1)
	ret0, _ := ret[0].([]*entity.Image)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List.
func (mr *MockImageMockRecorder) List(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockImage)(nil).List), arg0, arg1)
}
