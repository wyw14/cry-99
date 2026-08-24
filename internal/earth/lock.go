package earth

import "sync"

type Lock struct {
	mu    sync.RWMutex
	areas map[string]bool
}

func NewLock() *Lock { return &Lock{areas: map[string]bool{}} }
func (l *Lock) Set(area string, locked bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.areas[area] = locked
}
func (l *Lock) IsLocked(area string) bool { l.mu.RLock(); defer l.mu.RUnlock(); return l.areas[area] }
