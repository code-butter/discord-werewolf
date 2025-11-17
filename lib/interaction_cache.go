package lib

import "time"

type interactionCacheMap[T any] map[string]interactionCacheMapValue[T]

type interactionCacheMapValue[T any] struct {
	value      T
	expiration time.Time
}

type InteractionCache[T any] struct {
	interactionCacheMap[T]
	expiresDuration time.Duration
	clock           Clock
}

func NewInteractionCache[T any](expiresDuration time.Duration, clock Clock) *InteractionCache[T] {
	return &InteractionCache[T]{
		interactionCacheMap: interactionCacheMap[T]{},
		expiresDuration:     expiresDuration,
		clock:               clock,
	}
}

func (ic InteractionCache[T]) Get(key string) (*T, bool) {
	if icv, ok := ic.interactionCacheMap[key]; ok {
		if icv.expiration.After(time.Now()) {
			return &icv.value, true
		}
	}
	return nil, false
}

func (ic InteractionCache[T]) Set(key string, value T) {
	ic.interactionCacheMap[key] = interactionCacheMapValue[T]{
		value:      value,
		expiration: time.Now().Add(ic.expiresDuration),
	}
}
