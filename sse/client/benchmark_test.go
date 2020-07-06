package client_test

import (
	"testing"
	"time"

	"github.com/justinjudd/competition-example/web/sse/client"
	"github.com/justinjudd/competition-example/web/sse/event"
)

func BenchmarkClient_Write(b *testing.B) {
	b.StopTimer()
	client := client.New(time.Second, 3, "test")

	go func() {
		for {
			<-client.Listen()
		}
	}()

	data := make([]byte, 1024)
	evt := event.New("test", data)

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		client.Write(evt)
	}
}

func BenchmarkClient_Listen(b *testing.B) {
	b.StopTimer()
	client := client.New(time.Second, 3, "test")

	go func() {
		for i := 0; i < b.N; i++ {
			data := make([]byte, 1024)
			evt := event.New("test", data)

			client.Write(evt)
		}
	}()

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		<-client.Listen()
	}
}
