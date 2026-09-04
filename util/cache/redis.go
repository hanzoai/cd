package cache

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/hanzoai/cd/util/env"
	utilio "github.com/hanzoai/cd/util/io"

	"github.com/hanzokv/go/v9"
)

type RedisCompressionType string

var (
	RedisCompressionNone RedisCompressionType = "none"
	RedisCompressionGZip RedisCompressionType = "gzip"
)

const (
	// envRedisKeyPrefix is an env variable name which stores the prefix for redis keys
	envRedisKeyPrefix = "CD_KV_KEY_PREFIX"
)

func CompressionTypeFromString(s string) (RedisCompressionType, error) {
	switch s {
	case string(RedisCompressionNone):
		return RedisCompressionNone, nil
	case string(RedisCompressionGZip):
		return RedisCompressionGZip, nil
	}
	return "", fmt.Errorf("unknown compression type: %s", s)
}

func NewRedisCache(client *kv.Client, expiration time.Duration, compressionType RedisCompressionType) CacheClient {
	return &redisCache{
		client:               client,
		expiration:           expiration,
		redisCompressionType: compressionType,
		prefix:               env.StringFromEnv(envRedisKeyPrefix, ""),
	}
}

// compile-time validation of adherence of the CacheClient contract
var _ CacheClient = &redisCache{}

type redisCache struct {
	expiration           time.Duration
	client               *kv.Client
	redisCompressionType RedisCompressionType
	// prefix is added to all keys stored in redis
	prefix string
}

func (r *redisCache) getKey(key string) string {
	prefixedKey := r.prefix + key
	switch r.redisCompressionType {
	case RedisCompressionGZip:
		return prefixedKey + ".gz"
	default:
		return prefixedKey
	}
}

func (r *redisCache) marshal(obj any) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	var w io.Writer = buf
	if r.redisCompressionType == RedisCompressionGZip {
		w = gzip.NewWriter(buf)
	}
	encoder := json.NewEncoder(w)

	if err := encoder.Encode(obj); err != nil {
		return nil, err
	}
	if flusher, ok := w.(interface{ Flush() error }); ok {
		if err := flusher.Flush(); err != nil {
			return nil, err
		}
	}
	if closer, ok := w.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func (r *redisCache) unmarshal(data []byte, obj any) error {
	buf := bytes.NewReader(data)
	var reader io.Reader = buf
	if r.redisCompressionType == RedisCompressionGZip {
		gzipReader, err := gzip.NewReader(buf)
		if err != nil {
			return err
		}
		reader = gzipReader
	}
	if err := json.NewDecoder(reader).Decode(obj); err != nil {
		return fmt.Errorf("failed to decode cached data: %w", err)
	}
	return nil
}

func (r *redisCache) Rename(oldKey string, newKey string, _ time.Duration) error {
	err := r.client.Rename(context.TODO(), r.getKey(oldKey), r.getKey(newKey)).Err()
	if err != nil && err.Error() == "ERR no such key" {
		err = ErrCacheMiss
	}

	return err
}

func (r *redisCache) Set(item *Item) error {
	expiration := item.CacheActionOpts.Expiration
	if expiration == 0 {
		expiration = r.expiration
	}

	val, err := r.marshal(item.Object)
	if err != nil {
		return err
	}

	key := r.getKey(item.Key)
	if item.CacheActionOpts.DisableOverwrite {
		return r.client.SetNX(context.TODO(), key, val, expiration).Err()
	}
	return r.client.Set(context.TODO(), key, val, expiration).Err()
}

func (r *redisCache) Get(key string, obj any) error {
	data, err := r.client.Get(context.TODO(), r.getKey(key)).Bytes()
	if errors.Is(err, kv.Nil) {
		return ErrCacheMiss
	}
	if err != nil {
		return err
	}
	return r.unmarshal(data, obj)
}

func (r *redisCache) Delete(key string) error {
	return r.client.Del(context.TODO(), r.getKey(key)).Err()
}

func (r *redisCache) OnUpdated(ctx context.Context, key string, callback func() error) error {
	pubsub := r.client.Subscribe(ctx, key)
	defer utilio.Close(pubsub)

	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ch:
			if err := callback(); err != nil {
				return err
			}
		}
	}
}

func (r *redisCache) NotifyUpdated(key string) error {
	return r.client.Publish(context.TODO(), key, "").Err()
}

type MetricsRegistry interface {
	IncRedisRequest(failed bool)
	ObserveRedisRequestDuration(duration time.Duration)
}

type redisHook struct {
	registry MetricsRegistry
}

// ignoredCommandNames are commands the client may issue during connection setup and bookkeeping
// and which we don't want to count as application-level requests in metrics.
var ignoredCommandNames = map[string]struct{}{
	"hello":  {},
	"client": {},
	// Optional: we can enable if we want also want to exclude other setup/noise commands.
	// "auth":   {},
	// "select": {},
	// "ping":   {},
}

func shouldIgnoreRedisCmd(cmd kv.Cmder) bool {
	name := strings.ToLower(strings.TrimSpace(cmd.Name()))
	_, ok := ignoredCommandNames[name]
	return ok
}

func (rh *redisHook) DialHook(next kv.DialHook) kv.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		conn, err := next(ctx, network, addr)
		return conn, err
	}
}

func (rh *redisHook) ProcessHook(next kv.ProcessHook) kv.ProcessHook {
	return func(ctx context.Context, cmd kv.Cmder) error {
		startTime := time.Now()

		err := next(ctx, cmd)

		if shouldIgnoreRedisCmd(cmd) {
			return err
		}

		rh.registry.IncRedisRequest(err != nil && !errors.Is(err, kv.Nil))
		rh.registry.ObserveRedisRequestDuration(time.Since(startTime))

		return err
	}
}

func (redisHook) ProcessPipelineHook(_ kv.ProcessPipelineHook) kv.ProcessPipelineHook {
	return nil
}

// CollectMetrics add transport wrapper that pushes metrics into the specified metrics registry
// Lock should be shared between functions that can add/process a Redis hook.
func CollectMetrics(client *kv.Client, registry MetricsRegistry, lock *sync.RWMutex) {
	if lock != nil {
		lock.Lock()
		defer lock.Unlock()
	}
	client.AddHook(&redisHook{registry: registry})
}
