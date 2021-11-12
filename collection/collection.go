package collection

import (
	crand "crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"sync"
	"time"
)

var (
	ErrorCollision = errors.New("collection shortname collision")
)

type Entry struct {
	URL     *url.URL
	Expires time.Time
	Uses    int
}

type Collection struct {
	entries map[string]*Entry
	m       sync.RWMutex
}

func generate(url string, n int) (string, error) {
	buf := make([]byte, n)
	// Use crypto/rand just in case.
	_, err := crand.Read(buf)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func (c *Collection) Fetch(shortname string) (*Entry, bool) {
	c.m.RLock()
	defer c.m.RUnlock()

	got, ok := c.entries[shortname]
	if !ok {
		return nil, false
	}
	if got.Uses == 1 {
		delete(c.entries, shortname)
		return got, true
	} else if got.Uses > 1 {
		got.Uses--
	}
	return got, false
}

func (c *Collection) Submit(
	u *url.URL,
	shortbytes int,
	expires time.Time,
	uses int) (string, error) {

	shortname, err := generate(u.String(), shortbytes)
	if err != nil {
		return "", err
	}

	c.m.Lock()
	defer c.m.Unlock()

	if _, ok := c.entries[shortname]; ok {
		return "", ErrorCollision
	}

	c.entries[shortname] = &Entry{
		URL:     u,
		Expires: expires,
		Uses:    uses,
	}
	return shortname, nil
}

func (c *Collection) Reap(cb func(string, *Entry)) {
	now := time.Now()

	c.m.Lock()
	defer c.m.Unlock()

	for short, entry := range c.entries {
		if now.After(entry.Expires) {
			if cb != nil {
				cb(short, entry)
			}
			delete(c.entries, short)
		}
	}
}

func (c *Collection) Dump(w io.Writer) error {
	e := json.NewEncoder(w)
	e.SetIndent("", "  ")
	c.m.RLock()
	defer c.m.RUnlock()
	return e.Encode(c.entries)
}

func (c *Collection) Import(r io.Reader) error {
	e := json.NewDecoder(r)
	return e.Decode(&c.entries)
}

func New() *Collection {
	return &Collection{entries: map[string]*Entry{}}
}
