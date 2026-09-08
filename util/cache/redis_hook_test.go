package cache

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"

	"github.com/hanzokv/go/v9"
)

func Test_ReconnectCallbackHookCalled(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer mr.Close()

	called := false
	hook := NewRedisReconnectHook(func() {
		called = true
	})

	faultyDNSRedisClient := kv.NewClient(&kv.Options{Addr: "invalidredishost.invalid:12345"})
	faultyDNSRedisClient.AddHook(hook)

	faultyDNSClient := NewRedisCache(faultyDNSRedisClient, 60*time.Second, RedisCompressionNone)
	err = faultyDNSClient.Set(&Item{Key: "baz", Object: "foo"})
	assert.True(t, called)
	assert.Error(t, err)
}

func Test_ReconnectCallbackHookNotCalled(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer mr.Close()

	called := false
	hook := NewRedisReconnectHook(func() {
		called = true
	})

	redisClient := kv.NewClient(&kv.Options{Addr: mr.Addr()})
	redisClient.AddHook(hook)
	client := NewRedisCache(redisClient, 60*time.Second, RedisCompressionNone)

	err = client.Set(&Item{Key: "foo", Object: "bar"})
	assert.False(t, called)
	assert.NoError(t, err)
}
