package broker_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/justinjudd/competition-example/web/sse/broker"
	"github.com/justinjudd/competition-example/web/sse/event"
	"github.com/justinjudd/competition-example/web/sse/test"
	"github.com/stretchr/testify/assert"
)

func TestBroker_New(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Timeout   time.Duration
		Tolerance int
	}{
		{Timeout: time.Second, Tolerance: 3},
	}

	for _, tc := range tt {
		broker := broker.New(tc.Timeout, tc.Tolerance, nil)

		assert.NotNil(t, broker)
	}
}

func TestBroker_UniqueIDs(t *testing.T) {
	tt := []struct {
		Timeout       time.Duration
		Tolerance     int
		ExpectedError string
		ExpectedCode  int
		ClientID      string
	}{
		{
			Timeout:      time.Second,
			Tolerance:    3,
			ClientID:     "1234",
			ExpectedCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tt {
		broker := broker.New(tc.Timeout, tc.Tolerance, nil)

		w := test.NewMockRecorder()
		r := httptest.NewRequest("GET", "/connect?id="+tc.ClientID, nil)

		go broker.ClientHandler(w, r)
		<-time.Tick(time.Millisecond)

		r2 := httptest.NewRequest("GET", "/connect?id="+tc.ClientID, nil)
		w2 := test.NewMockRecorder()

		broker.ClientHandler(w2, r2)

		assert.Equal(t, tc.ExpectedCode, w2.Code)
	}
}

func TestBroker_Handlers(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Timeout            time.Duration
		Tolerance          int
		ContentType        string
		ExpectedError      string
		Recorder           http.ResponseWriter
		Data               []byte
		ExpectedCode       int
		AssertErrorHandler bool
		ClientID           string
	}{
		{
			Timeout:      time.Second,
			Tolerance:    3,
			ContentType:  "text/event-stream",
			Recorder:     test.NewMockRecorder(),
			ExpectedCode: http.StatusOK,
			Data:         []byte{1, 2, 3},
		},
		{
			Timeout:      time.Second,
			Tolerance:    3,
			ContentType:  "text/event-stream",
			Recorder:     test.NewMockRecorder(),
			ExpectedCode: http.StatusOK,
			Data:         []byte{1, 2, 3},
		},
		{
			Timeout:            time.Second,
			Tolerance:          3,
			ContentType:        "text/event-stream",
			Recorder:           httptest.NewRecorder(),
			ExpectedError:      "client does not support streaming",
			AssertErrorHandler: true,
			Data:               []byte{1, 2, 3},
		},
		{
			Timeout:       time.Second,
			Tolerance:     3,
			ContentType:   "text/event-stream",
			Recorder:      httptest.NewRecorder(),
			ExpectedError: "client does not support streaming",
			Data:          []byte{1, 2, 3},
		},
		{
			Timeout:       time.Second,
			Tolerance:     3,
			ContentType:   "text/event-stream",
			Recorder:      httptest.NewRecorder(),
			ExpectedError: "no event data provided",
		},
		{
			Timeout:       time.Second,
			Tolerance:     3,
			ContentType:   "text/event-stream",
			Recorder:      httptest.NewRecorder(),
			ExpectedError: "no event data provided",
			ClientID:      "123",
			Data:          []byte{1, 2, 3},
		},
	}

	for _, tc := range tt {
		var handler broker.ErrorHandler

		if tc.AssertErrorHandler {
			handler = func(w http.ResponseWriter, r *http.Request, err error) {
				assert.Contains(t, err.Error(), tc.ExpectedError)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}

		// Create a new broker
		broker := broker.New(tc.Timeout, tc.Tolerance, handler)

		// The test recorder allows us to cast to http.Flusher & http.CloseNotifier
		w := tc.Recorder

		// Create the request
		r, _ := http.NewRequest("GET", "/connect?id="+tc.ClientID, nil)
		r.Header.Add("Content-Type", tc.ContentType)

		// Connect to the broker, give it 10ms to create the
		// client
		go broker.ClientHandler(w, r)
		<-time.Tick(time.Millisecond * 10)

		rec := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/broadcast?id="+tc.ClientID, bytes.NewBuffer(tc.Data))

		// Post an event.
		broker.EventHandler(rec, req)

		if tc.ExpectedCode > 0 {
			assert.Equal(t, tc.ExpectedCode, rec.Code)
		}
	}
}

func TestBroker_Listen(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Timeout       time.Duration
		Tolerance     int
		ContentType   string
		ExpectedError string
		Data          []byte
		Type          string
	}{
		{Timeout: time.Second, Tolerance: 3, Data: []byte{1, 2, 3}},
	}

	for _, tc := range tt {
		// Create a new broker
		broker := broker.New(tc.Timeout, tc.Tolerance, nil)

		go func() {
			data := <-broker.Listen()

			assert.Equal(t, tc.Data, data)
		}()

		evt := event.New(tc.Type, tc.Data)

		if err := broker.Broadcast(evt); err != nil {
			assert.Contains(t, err.Error(), tc.ExpectedError)
		}
	}
}

func TestBroker_Broadcast(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Timeout       time.Duration
		Tolerance     int
		ContentType   string
		ExpectedError string
		Data          []byte
		Type          string
	}{
		{
			Timeout:     time.Second,
			Tolerance:   3,
			ContentType: "text/event-stream",
		},
	}

	for _, tc := range tt {
		// Create a new broker
		broker := broker.New(tc.Timeout, tc.Tolerance, nil)

		// The test recorder allows us to cast to http.Flusher & http.CloseNotifier
		w := test.NewMockRecorder()

		// Create the request
		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Add("Content-Type", tc.ContentType)

		// Connect to the broker, give it 1 second to create the
		// client
		go broker.ClientHandler(w, r)
		<-time.Tick(time.Second)

		evt := event.New(tc.Type, tc.Data)

		if err := broker.Broadcast(evt); err != nil {
			assert.Contains(t, err.Error(), tc.ExpectedError)
		}
	}
}

func TestBroker_BroadcastTo(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Timeout       time.Duration
		Tolerance     int
		ContentType   string
		ExpectedError string
		ClientID      string
		IDParam       string
		Data          []byte
		Type          string
	}{
		{
			Timeout:     time.Second,
			Tolerance:   3,
			ContentType: "text/event-stream",
			ClientID:    "1234",
			IDParam:     "1234",
		},
		{
			Timeout:       time.Second,
			Tolerance:     3,
			ContentType:   "text/event-stream",
			ClientID:      "",
			IDParam:       "1234",
			ExpectedError: "no client with id",
		},
	}

	for _, tc := range tt {
		// Create a new broker
		broker := broker.New(tc.Timeout, tc.Tolerance, nil)

		// The test recorder allows us to cast to http.Flusher & http.CloseNotifier
		w := test.NewMockRecorder()

		// Create the request
		r := httptest.NewRequest("GET", "/connect?id="+tc.IDParam, nil)
		r.Header.Add("Content-Type", tc.ContentType)

		// Connect to the broker, give it 1 second to create the
		// client
		go broker.ClientHandler(w, r)
		<-time.Tick(time.Second)

		evt := event.New(tc.Type, tc.Data)

		if err := broker.BroadcastTo(tc.ClientID, evt); err != nil {
			assert.Contains(t, err.Error(), tc.ExpectedError)
		}
	}
}
