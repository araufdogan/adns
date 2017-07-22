package main

import (
	"time"
	"sync"
	"github.com/miekg/dns"
	"errors"
)


// Cache interface
type Cache interface {
	AddOrUpdate(Q dns.Question, Msg *dns.Msg) error
	Get(q dns.Question) (error, *dns.Msg)
}

// MemoryCache type
type MemoryCache struct {
	Storage		map[string]Message
	MaxCount	int
	Mu		sync.RWMutex
	Timer		*time.Ticker
}

// Mesg represents a cache entry
type Message struct {
	Msg     *dns.Msg
	Expire  time.Time
}

func (c *MemoryCache) AddOrUpdate(q dns.Question, Msg *dns.Msg) error {
	c.Mu.Lock()

	if c.MaxCount != 0 && len(c.Storage) >= c.MaxCount {
		c.Storage = make(map[string]Message, 0)
	}

	c.Storage[q.String()] = Message{
		Msg: Msg,
		Expire: time.Now().Add(time.Duration(Msg.Answer[0].Header().Ttl) * time.Second),
	}

	c.Mu.Unlock()

	return nil
}

func (c *MemoryCache) Get(q dns.Question) (error, *dns.Msg) {
	c.Mu.Lock()

	if Msg, ok := c.Storage[q.String()]; ok {
		c.Mu.Unlock()
		return nil, Msg.Msg
	} else {
		c.Mu.Unlock()
		return errors.New("not found"), nil
	}
}

func (c *MemoryCache) Delete(q string) error {
	c.Mu.Lock()

	delete(c.Storage, q)

	c.Mu.Unlock()

	return nil
}

func (c *MemoryCache) Start() {
	c.Timer = time.NewTicker(10 * time.Second)
	go func() {
		for range c.Timer.C {
			t := time.Now()

			c.Mu.Lock()

			for k, v := range c.Storage {
				if v.Expire.After(t) {
					c.Delete(k)
				}
			}

			c.Mu.Unlock()
		}
	}()
}














