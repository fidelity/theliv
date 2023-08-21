package common

import "sync"

type LockedSlice struct {
	mx    *sync.Mutex
	store []string
}

func InitLockedSlice() *LockedSlice {
	return &LockedSlice{
		mx:    &sync.Mutex{},
		store: make([]string, 0),
	}
}

func (l *LockedSlice) GetStore() []string {
	l.mx.Lock()
	defer l.mx.Unlock()
	store := l.store

	return store
}

func (l *LockedSlice) SetStore(store []string) {
	l.mx.Lock()
	defer l.mx.Unlock()

	l.store = store
}

func (l *LockedSlice) Append(item ...string) {
	l.mx.Lock()
	defer l.mx.Unlock()

	l.store = append(l.store, item...)
}
