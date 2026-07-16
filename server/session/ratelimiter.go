package session

import (
	utilio "github.com/hanzoai/deploy/v3/util/io"
	"github.com/hanzoai/deploy/v3/util/session"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"
)

func NewLoginRateLimiter(maxNumber int) func() (utilio.Closer, error) {
	semaphore := semaphore.NewWeighted(int64(maxNumber))
	return func() (utilio.Closer, error) {
		if !semaphore.TryAcquire(1) {
			log.Warnf("Exceeded number of concurrent login requests")
			return nil, session.InvalidLoginErr
		}
		return utilio.NewCloser(func() error {
			defer semaphore.Release(1)
			return nil
		}), nil
	}
}
