package event_test

import (
	"testing"

	"github.com/justinjudd/competition-example/web/sse/event"
	"github.com/stretchr/testify/assert"
)

func TestEvent_New(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Type string
		Data []byte
	}{
		{Type: "test", Data: make([]byte, 1024)},
	}

	for _, tc := range tt {
		evt := event.New(tc.Type, tc.Data)

		assert.NotNil(t, evt)
	}
}

func TestEvent_String(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Type           string
		Data           []byte
		ExpectedString string
	}{
		{Type: "test", Data: []byte{68, 67, 66}, ExpectedString: "event:test\ndata:DCB\n\n"},
	}

	for _, tc := range tt {
		evt := event.New(tc.Type, tc.Data)
		out := evt.String()

		assert.NotEmpty(t, out)
		assert.Equal(t, tc.ExpectedString, out)
	}
}
