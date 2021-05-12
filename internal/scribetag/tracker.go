package scribetag

import (
	"fmt"
	"sync"
	"time"
)

// TODO:
//   The correct implementation is to use heap based on ScribeTag.LastPulse
//   However, this works good enough for now, as long as `now` parameter isn't
//     being inserted with anything too far off from real time.Now()
type ScribeTracker struct {
	tags   map[string]*ScribeTag
	newest string
	oldest string
	lock   *sync.RWMutex
}

func NewTracker() *ScribeTracker {
	return &ScribeTracker{
		tags: map[string]*ScribeTag{},
		lock: &sync.RWMutex{},
	}
}

func (st *ScribeTracker) Oldest() *ScribeTag {
	st.lock.RLock()
	tag := st.tags[st.oldest]
	st.lock.RUnlock()

	return tag
}

func (st *ScribeTracker) Add(key string, now time.Time) error {
	st.lock.Lock()
	defer st.lock.Unlock()

	if existing := st.tags[key]; existing != nil {
		return fmt.Errorf("abort: attempt to overwrite existing tag with key: %s", key)
	}

	st.tags[key] = &ScribeTag{
		Name:      key,
		LastPulse: now,
	}

	st.setNewest(key)
	if st.oldest == "" {
		st.oldest = key
	}
	return nil
}

func (st *ScribeTracker) Take(key string) *ScribeTag {
	st.lock.Lock()
	st.detach(key)
	tag := st.tags[key]
	delete(st.tags, key)
	st.lock.Unlock()

	return tag
}

func (st *ScribeTracker) Pulse(key string, now time.Time) error {
	st.lock.Lock()
	defer st.lock.Unlock()

	tag := st.tags[key]
	if tag == nil {
		return fmt.Errorf("untracked scribe with key: %s", key)
	}
	tag.LastPulse = now
	st.detach(key)
	st.setNewest(key)
	return nil
}

func (st *ScribeTracker) setNewest(key string) {
	oldNewest := st.tags[st.newest]
	if oldNewest != nil {
		oldNewest.before = key
	}
	st.tags[key].after = st.newest
	st.newest = key
}

func (st *ScribeTracker) detach(key string) {
	tag := st.tags[key]
	if tag == nil {
		return
	}
	before := st.tags[tag.before]
	if before != nil {
		before.after = tag.after
	}
	after := st.tags[tag.after]
	if after != nil {
		after.before = tag.before
	}
	if st.newest == key {
		st.newest = tag.after
	}
	if st.oldest == key {
		st.oldest = tag.before
	}
}
