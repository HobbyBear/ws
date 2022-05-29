package broker

import (
	"context"
	"github.com/go-redis/redis"
)

type redisPubSub interface {
	redis.Cmdable
	Subscribe(ctx context.Context, channels ...string) *redis.PubSub
	PSubscribe(ctx context.Context, channels ...string) *redis.PubSub
	Close() error
}

type RedisBroker struct {
	redisPubSub
	ch chan []byte
}

func (r *RedisBroker) Close() error {
	return r.redisPubSub.Close()
}

func (r *RedisBroker) Pub(msg []byte) error {
	_, err := r.Publish(context.TODO(), "pushmsg", msg).Result()
	return err
}

func (r *RedisBroker) Sub() chan []byte {
	sub := r.Subscribe(context.TODO(), "pushmsg")
	go func() {
		for msg := range sub.Channel() {
			r.ch <- []byte(msg.Payload)
		}
	}()
	return r.ch
}

func NewRedisBroker(r redisPubSub) *RedisBroker {
	return &RedisBroker{redisPubSub: r, ch: make(chan []byte, 10000)}
}
