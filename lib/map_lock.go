package lib

import "sync"

type MapLockValues[T any] map[string]T

type MapLock[T any] struct {
	*MapLockValues[T]
	lock *sync.Mutex
}

func NewMapLock[T any]() *MapLock[T] {
	return &MapLock[T]{
		MapLockValues: &MapLockValues[T]{},
		lock:          &sync.Mutex{},
	}
}

func (ml *MapLock[T]) GetOrSet(id string, setter func() (T, error)) (T, error) {
	if value, ok := (*ml.MapLockValues)[id]; ok {
		return value, nil
	}
	ml.lock.Lock()
	defer ml.lock.Unlock()
	if value, ok := (*ml.MapLockValues)[id]; ok {
		return value, nil
	}
	value, err := setter()
	if err != nil {
		(*ml.MapLockValues)[id] = value
	}
	return value, err
}
