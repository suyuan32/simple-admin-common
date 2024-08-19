package nats

import (
	"fmt"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

func TestNewJetstream(t *testing.T) {
	t.Skip()

	conf := Conf{
		Hosts:         []string{"nats://localhost:4222"},
		ReconnectWait: 5,
		MaxReconnect:  5,
		UserCred:      "",
	}

	s, err := conf.NewJetStream()
	if err != nil {
		return
	}

	ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(10*time.Minute))

	stream, err := s.CreateStream(ctx, jetstream.StreamConfig{
		Name:        "Test",
		Description: "Test",
		Subjects:    []string{"Test.*"},
	})

	assert.Nil(t, err)

	endlessPublish := func(ctx context.Context, js jetstream.JetStream) {
		var i int
		for {
			time.Sleep(500 * time.Millisecond)

			if _, err := js.Publish(ctx, "Test.new", []byte(fmt.Sprintf("test msg %d", i))); err != nil {
				fmt.Println("pub error: ", err)
			}
			i++
		}
	}

	go endlessPublish(ctx, s)

	cons, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Name:      "TestCons",
		AckPolicy: jetstream.AckExplicitPolicy,
	})

	assert.Nil(t, err)

	cc, err := cons.Consume(func(msg jetstream.Msg) {
		fmt.Println(string(msg.Data()))
		msg.Ack()
	}, jetstream.ConsumeErrHandler(func(consumeCtx jetstream.ConsumeContext, err error) {
		fmt.Println(err)
	}))
	if err != nil {
		log.Fatal(err)
	}
	defer cc.Stop()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
}
