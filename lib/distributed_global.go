package lib

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"sync"
	"time"

	"log/slog"

	"github.com/Clever/leakybucket"
	"github.com/Clever/leakybucket/memory"
)

type ClusterGlobalRateLimiter struct {
	sync.RWMutex
	globalBucketsMap map[uint64]*leakybucket.Bucket
	memStorage       *memory.Storage
}

func NewClusterGlobalRateLimiter() *ClusterGlobalRateLimiter {
	memStorage := memory.New()
	return &ClusterGlobalRateLimiter{
		memStorage:       memStorage,
		globalBucketsMap: make(map[uint64]*leakybucket.Bucket),
	}
}

func (c *ClusterGlobalRateLimiter) Take(botHash uint64, botLimit uint) {
	bucket := *c.getOrCreate(botHash, botLimit)
takeGlobal:
	_, err := bucket.Add(1)
	if err != nil {
		reset := bucket.Reset()

		slog.Debug("Failed to grab global token, sleeping for a bit",
			"waitTime", time.Until(reset),
		)

		time.Sleep(time.Until(reset))

		goto takeGlobal
	}
}

func (c *ClusterGlobalRateLimiter) getOrCreate(botHash uint64, botLimit uint) *leakybucket.Bucket {
	c.RLock()
	b, ok := c.globalBucketsMap[botHash]
	c.RUnlock()

	if !ok {
		c.Lock()
		// Check if it wasn't created while we didn't hold the exclusive lock
		b, ok = c.globalBucketsMap[botHash]
		if ok {
			c.Unlock()

			return b
		}

		globalBucket, _ := c.memStorage.Create(strconv.FormatUint(botHash, 10), botLimit, 1*time.Second)
		c.globalBucketsMap[botHash] = &globalBucket

		c.Unlock()

		return &globalBucket
	} else {
		return b
	}
}

func (c *ClusterGlobalRateLimiter) FireGlobalRequest(ctx context.Context, addr string, botHash uint64, botLimit uint) error {
	globalReq, err := http.NewRequestWithContext(ctx, "GET", "http://"+addr+"/nirn/global", nil)
	if err != nil {
		return err
	}

	globalReq.Header.Set("Bot-Hash", strconv.FormatUint(botHash, 10))
	globalReq.Header.Set("Bot-Limit", strconv.FormatUint(uint64(botLimit), 10))

	// The node handling the request will only return if we grabbed a token or an error was thrown
	resp, err := client.Do(globalReq)

	slog.Debug("Got go-ahead for global")

	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return errors.New("global request failed with status " + resp.Status)
	}

	return nil
}
