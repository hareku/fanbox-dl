// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/fanbox/api_client.go

// Package mock_fanbox is a generated GoMock package.
package mock_fanbox

import (
	context "context"
	http "net/http"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockApiClient is a mock of ApiClient interface.
type MockApiClient struct {
	ctrl     *gomock.Controller
	recorder *MockApiClientMockRecorder
}

// MockApiClientMockRecorder is the mock recorder for MockApiClient.
type MockApiClientMockRecorder struct {
	mock *MockApiClient
}

// NewMockApiClient creates a new mock instance.
func NewMockApiClient(ctrl *gomock.Controller) *MockApiClient {
	mock := &MockApiClient{ctrl: ctrl}
	mock.recorder = &MockApiClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockApiClient) EXPECT() *MockApiClientMockRecorder {
	return m.recorder
}

// Request mocks base method.
func (m *MockApiClient) Request(ctx context.Context, url string) (*http.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Request", ctx, url)
	ret0, _ := ret[0].(*http.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Request indicates an expected call of Request.
func (mr *MockApiClientMockRecorder) Request(ctx, url interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Request", reflect.TypeOf((*MockApiClient)(nil).Request), ctx, url)
}

// RequestAsJSON mocks base method.
func (m *MockApiClient) RequestAsJSON(ctx context.Context, url string, v interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RequestAsJSON", ctx, url, v)
	ret0, _ := ret[0].(error)
	return ret0
}

// RequestAsJSON indicates an expected call of RequestAsJSON.
func (mr *MockApiClientMockRecorder) RequestAsJSON(ctx, url, v interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RequestAsJSON", reflect.TypeOf((*MockApiClient)(nil).RequestAsJSON), ctx, url, v)
}
