package client_test

import (
	"testing"
	"time"

	"github.com/justinjudd/competition-example/web/sse/event"

	"github.com/justinjudd/competition-example/web/sse/client"
	"github.com/stretchr/testify/assert"
)

func TestClient_New(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Timeout   time.Duration
		Tolerance int
		ID        string
	}{
		{Timeout: time.Second, Tolerance: 3},
		{Timeout: time.Second, Tolerance: 3, ID: "test"},
	}

	for _, tc := range tt {
		client := client.New(tc.Timeout, tc.Tolerance, tc.ID)

		assert.NotNil(t, client)
		assert.NotEqual(t, "", client.ID())
		assert.NotEqual(t, true, client.ShouldDisconnect())

		if tc.ID != "" {
			assert.Equal(t, tc.ID, client.ID())
		}
	}
}

func TestClient_ReadWrite(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Timeout       time.Duration
		Tolerance     int
		Data          []byte
		ExpectedError string
		HasListener   bool
		Type          string
	}{
		{Timeout: time.Second, Tolerance: 3, ExpectedError: "timeout exceeded"},
		{Timeout: time.Second, Tolerance: 3, HasListener: true},
	}

	for _, tc := range tt {
		client := client.New(tc.Timeout, tc.Tolerance, "")

		if tc.HasListener {
			go func() { <-client.Listen() }()
		}

		evt := event.New(tc.Type, tc.Data)

		if err := client.Write(evt); err != nil {
			assert.Contains(t, err.Error(), tc.ExpectedError)
		}
	}
}
