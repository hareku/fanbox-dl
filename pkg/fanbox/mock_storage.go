// Code generated by MockGen. DO NOT EDIT.
// Source: storage.go

// Package fanbox is a generated GoMock package.
package fanbox

import (
	io "io"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockStorage is a mock of Storage interface.
type MockStorage struct {
	ctrl     *gomock.Controller
	recorder *MockStorageMockRecorder
}

// MockStorageMockRecorder is the mock recorder for MockStorage.
type MockStorageMockRecorder struct {
	mock *MockStorage
}

// NewMockStorage creates a new mock instance.
func NewMockStorage(ctrl *gomock.Controller) *MockStorage {
	mock := &MockStorage{ctrl: ctrl}
	mock.recorder = &MockStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStorage) EXPECT() *MockStorageMockRecorder {
	return m.recorder
}

// Exist mocks base method.
func (m *MockStorage) Exist(post PostInfoBody, order int, img Image) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Exist", post, order, img)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Exist indicates an expected call of Exist.
func (mr *MockStorageMockRecorder) Exist(post, order, img interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Exist", reflect.TypeOf((*MockStorage)(nil).Exist), post, order, img)
}

// Save mocks base method.
func (m *MockStorage) Save(post PostInfoBody, order int, img Image, file io.Reader) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Save", post, order, img, file)
	ret0, _ := ret[0].(error)
	return ret0
}

// Save indicates an expected call of Save.
func (mr *MockStorageMockRecorder) Save(post, order, img, file interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Save", reflect.TypeOf((*MockStorage)(nil).Save), post, order, img, file)
}
