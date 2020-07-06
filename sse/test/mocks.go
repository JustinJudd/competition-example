package test

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

type (
	// MockRecorder mocks the http.ResponseWriter, http.Flusher & http.CloseNotify interfaces.
	MockRecorder struct {
		mock.Mock

		header *http.Header
		closer chan bool

		Code int
	}
)

// NewMockRecorder creates a new instance of the MockRecorder type.
func NewMockRecorder() *MockRecorder {
	recorder := &MockRecorder{
		header: &http.Header{},
		closer: make(chan bool),
	}

	recorder.On("Header").Return(*recorder.header)
	recorder.On("CloseNotify").Return(recorder.closer)
	recorder.On("Write", mock.Anything).Return(0, nil)
	recorder.On("Flush").Return()
	recorder.On("WriteHeader", mock.Anything).Return()

	return recorder
}

// CloseNotify mocks the CloseNotify method of the http.CloseNotify interface.
func (mr *MockRecorder) CloseNotify() <-chan bool {
	args := mr.Called()

	channel, ok := args.Get(0).(<-chan bool)

	if !ok {
		return nil
	}

	return channel
}

// Header mocks the Header method of the http.ResponseWriter interface.
func (mr *MockRecorder) Header() http.Header {
	args := mr.Called()

	header, ok := args.Get(0).(http.Header)

	if !ok {
		return nil
	}

	return header
}

// Write mocks the Write method of the http.ResponseWriter interface.
func (mr *MockRecorder) Write(data []byte) (int, error) {
	args := mr.Called(data)

	return args.Int(0), args.Error(1)
}

// WriteHeader mocks the WriteHeader method of the http.ResponseWriter interface.
func (mr *MockRecorder) WriteHeader(code int) {
	mr.Called(code)
	mr.Code = code
}

// Flush mocks the Flush method of the http.Flusher interface.
func (mr *MockRecorder) Flush() {
	mr.Called()
}
