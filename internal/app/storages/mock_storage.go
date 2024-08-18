// Code generated by MockGen. DO NOT EDIT.
// Source: internal/app/storages/storage.go

// Package storages is a generated GoMock package.
package storages

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	zap "go.uber.org/zap"
)

// MockURLStorage is a mock of URLStorage interface.
type MockURLStorage struct {
	ctrl     *gomock.Controller
	recorder *MockURLStorageMockRecorder
}

// MockURLStorageMockRecorder is the mock recorder for MockURLStorage.
type MockURLStorageMockRecorder struct {
	mock *MockURLStorage
}

// NewMockURLStorage creates a new mock instance.
func NewMockURLStorage(ctrl *gomock.Controller) *MockURLStorage {
	mock := &MockURLStorage{ctrl: ctrl}
	mock.recorder = &MockURLStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockURLStorage) EXPECT() *MockURLStorageMockRecorder {
	return m.recorder
}

// DeleteHard mocks base method.
func (m *MockURLStorage) DeleteHard(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteHard", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteHard indicates an expected call of DeleteHard.
func (mr *MockURLStorageMockRecorder) DeleteHard(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteHard", reflect.TypeOf((*MockURLStorage)(nil).DeleteHard), ctx)
}

// DeleteUserURLs mocks base method.
func (m *MockURLStorage) DeleteUserURLs(ctx context.Context, listDeleted []string, logger *zap.SugaredLogger) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteUserURLs", ctx, listDeleted, logger)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteUserURLs indicates an expected call of DeleteUserURLs.
func (mr *MockURLStorageMockRecorder) DeleteUserURLs(ctx, listDeleted, logger interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteUserURLs", reflect.TypeOf((*MockURLStorage)(nil).DeleteUserURLs), ctx, listDeleted, logger)
}

// GetByID mocks base method.
func (m *MockURLStorage) GetByID(ctx context.Context, id string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByID", ctx, id)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByID indicates an expected call of GetByID.
func (mr *MockURLStorageMockRecorder) GetByID(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByID", reflect.TypeOf((*MockURLStorage)(nil).GetByID), ctx, id)
}

// GetUserURLs mocks base method.
func (m *MockURLStorage) GetUserURLs(ctx context.Context, baseURL string) ([]UserURLs, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserURLs", ctx, baseURL)
	ret0, _ := ret[0].([]UserURLs)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserURLs indicates an expected call of GetUserURLs.
func (mr *MockURLStorageMockRecorder) GetUserURLs(ctx, baseURL interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserURLs", reflect.TypeOf((*MockURLStorage)(nil).GetUserURLs), ctx, baseURL)
}

// IsExists mocks base method.
func (m *MockURLStorage) IsExists(ctx context.Context, key string) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsExists", ctx, key)
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsExists indicates an expected call of IsExists.
func (mr *MockURLStorageMockRecorder) IsExists(ctx, key interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsExists", reflect.TypeOf((*MockURLStorage)(nil).IsExists), ctx, key)
}

// LoadURLs mocks base method.
func (m *MockURLStorage) LoadURLs(arg0 context.Context, arg1 []Incoming, arg2 string) ([]Output, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LoadURLs", arg0, arg1, arg2)
	ret0, _ := ret[0].([]Output)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LoadURLs indicates an expected call of LoadURLs.
func (mr *MockURLStorageMockRecorder) LoadURLs(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LoadURLs", reflect.TypeOf((*MockURLStorage)(nil).LoadURLs), arg0, arg1, arg2)
}

// SaveURL mocks base method.
func (m *MockURLStorage) SaveURL(ctx context.Context, originalURL string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveURL", ctx, originalURL)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SaveURL indicates an expected call of SaveURL.
func (mr *MockURLStorageMockRecorder) SaveURL(ctx, originalURL interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveURL", reflect.TypeOf((*MockURLStorage)(nil).SaveURL), ctx, originalURL)
}