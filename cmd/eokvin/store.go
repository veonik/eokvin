package main

import (
	"crypto/rand"
	"errors"
	"math/big"
	"sync"
	"time"
)

// randSet is the set of characters used when generating a random key.
const randSet = "abcdefghijklmnopqrstuvwxyz1234567890"
const randSetLen = int64(len(randSet))

// An itemID is a unique key in a store.
type itemID string
func (c itemID) String() string {
	return string(c)
}

// An item is a single item, stored in a store.
type item struct {
	value string
	insertedAt time.Time
}
func (c item) String() string {
	return c.value
}

// A store is an in-memory data store with expiring items.
type store struct {
	mu sync.RWMutex
	entries map[itemID]item
	ttl time.Duration
}

// newItemID creates a new store key, ensuring it is unique.
func (cch *store) newItemID() (itemID, error) {
	b := make([]byte, 8)
	l := big.NewInt(randSetLen)
	for i := 0; i < 8; i++ {
		n, err := rand.Int(rand.Reader, l)
		if err != nil {
			return "", err
		}
		b[i] = randSet[int(n.Int64())]
	}
	k := itemID(b)
	cch.mu.RLock()
	defer cch.mu.RUnlock()
	// Avoid overwriting existing entries
	if _, ok := cch.entries[k]; ok {
		return "", errors.New("cache: collision detected")
	}
	return k, nil
}

// isExpired returns true if the given item is expired.
func (cch *store) isExpired(c item) bool {
	return c.insertedAt.Before(time.Now().Add(-1 * cch.ttl))
}

// expiredItemReaper deletes expired entries from the store at regular
// intervals.
func (cch *store) expiredItemReaper() error {
	for {
		select {
		case <-time.After(30 * time.Second):
			var del []itemID
			cch.mu.RLock()
			for k, v := range cch.entries {
				if cch.isExpired(v) {
					del = append(del, k)
				}
			}
			cch.mu.RUnlock()
			if len(del) == 0 {
				continue
			}
			cch.mu.Lock()
			for _, k := range del {
				delete(cch.entries, k)
			}
			cch.mu.Unlock()
		}
	}
}

// newItem initializes a new item.
func newItem(s string) item {
	return item{value: s, insertedAt: time.Now()}
}

