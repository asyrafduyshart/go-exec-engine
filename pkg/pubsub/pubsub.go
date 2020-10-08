package pubsub

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"cloud.google.com/go/pubsub"
	log "github.com/asyrafduyshart/go-exec-engine/pkg/log"
)

// PullMsgsConcurrenyControl ..
func PullMsgsConcurrenyControl(projectID, subID string) error {
	// projectID := "my-project-id"
	// subID := "my-sub"
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("pubsub.NewClient: %v", err)
	}
	defer client.Close()

	sub := client.Subscription(subID)
	// Must set ReceiveSettings.Synchronous to false (or leave as default) to enable
	// concurrency settings. Otherwise, NumGoroutines will be set to 1.
	sub.ReceiveSettings.Synchronous = false
	// NumGoroutines is the number of goroutines sub.Receive will spawn to pull messages concurrently.
	sub.ReceiveSettings.NumGoroutines = runtime.NumCPU()

	// Receive messages for 10 seconds.
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Create a channel to handle messages to as they come in.
	cm := make(chan *pubsub.Message)
	defer close(cm)
	// Handle individual messages in a goroutine.
	go func() {
		for msg := range cm {
			log.Info("Got message :%q\n", string(msg.Data))
			msg.Ack()
		}
	}()

	// Receive blocks until the context is cancelled or an error occurs.
	err = sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		cm <- msg
	})
	if err != nil {
		return fmt.Errorf("Receive: %v", err)
	}

	return nil
}

// PullMsgs ..
func PullMsgs(ctx context.Context, projectID, subID string, f func(string)) error {
	// projectID := "my-project-id"
	// subID := "my-sub"
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("pubsub.NewClient: %v", err)
	}

	// Consume 10 messages.
	var mu sync.Mutex
	received := 0
	sub := client.Subscription(subID)
	err = sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		mu.Lock()
		defer mu.Unlock()
		log.Info("Got message :%q\n", string(msg.Data))
		msg.Ack()
		f(string(msg.Data))
		received++
	})
	if err != nil {
		return fmt.Errorf("Receive: %v", err)
	}
	return nil
}
