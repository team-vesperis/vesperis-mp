package database

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	listeners    = make(map[string]*redis.PubSub)
	listenersMux sync.Mutex
	wg           sync.WaitGroup
)

func Publish(ctx context.Context, channel string, message any) error {
	return client.Publish(ctx, channel, message).Err()
}

func Subscribe(ctx context.Context, channel string) *redis.PubSub {
	return client.Subscribe(ctx, channel)
}

func SubscribeWithTimeout(ctx context.Context, channel string, timeout time.Duration) (*redis.Message, error) {
	pubsub := client.Subscribe(ctx, channel)
	defer pubsub.Close()

	ch := pubsub.Channel()
	select {
	case msg := <-ch:
		return msg, nil
	case <-time.After(timeout):
		return nil, context.DeadlineExceeded
	}
}

func SendAndReturn(ctx context.Context, publishChannel, responseChannel string, message any, timeout time.Duration) (*redis.Message, error) {
	pubsub := Subscribe(ctx, responseChannel)
	defer pubsub.Close()

	err := Publish(ctx, publishChannel, message)
	if err != nil {
		return nil, err
	}

	ch := pubsub.Channel()
	select {
	case msg := <-ch:
		return msg, nil
	case <-time.After(timeout):
		return nil, context.DeadlineExceeded
	}
}

func StartListener(ctx context.Context, channel string, handler func(msg *redis.Message)) {
	listenersMux.Lock()
	defer listenersMux.Unlock()

	if _, exists := listeners[channel]; exists {
		logger.Warn("Listener already exists for channel: ", channel)
		return
	}

	pubsub := Subscribe(ctx, channel)
	listeners[channel] = pubsub

	go func() {
		for {
			msg, err := pubsub.ReceiveMessage(ctx)
			if err != nil {
				if err == context.Canceled || ctx.Err() == context.Canceled {
					return
				}
				logger.Error("Error receiving message on channel ", channel, ": ", err)
				continue
			}
			wg.Add(1)
			handler(msg)
			wg.Done()
		}
	}()
}

func closeListeners() {
	logger.Info("Closing all listeners...")

	listenersMux.Lock()
	defer listenersMux.Unlock()

	wg.Wait()
	for channel, pubsub := range listeners {
		pubsub.Close()
		delete(listeners, channel)
	}

	logger.Info("Successfully closed all listeners.")
}
