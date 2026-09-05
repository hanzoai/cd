package cache

import (
	"context"
	"errors"
	"net"

	"github.com/hanzokv/go/v9"
	log "github.com/sirupsen/logrus"
)

type redisReconnectHook struct {
	reconnectCallback func()
}

func NewRedisReconnectHook(reconnectCallback func()) *redisReconnectHook {
	return &redisReconnectHook{reconnectCallback: reconnectCallback}
}

func (hook *redisReconnectHook) DialHook(next kv.DialHook) kv.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		conn, err := next(ctx, network, addr)
		return conn, err
	}
}

func (hook *redisReconnectHook) ProcessHook(next kv.ProcessHook) kv.ProcessHook {
	return func(ctx context.Context, cmd kv.Cmder) error {
		var dnsError *net.DNSError
		err := next(ctx, cmd)
		if err != nil && errors.As(err, &dnsError) {
			log.Warnf("Reconnect to redis because error: \"%v\"", err)
			hook.reconnectCallback()
		}
		return err
	}
}

func (hook *redisReconnectHook) ProcessPipelineHook(_ kv.ProcessPipelineHook) kv.ProcessPipelineHook {
	return nil
}
