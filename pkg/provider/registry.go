package provider

import (
	"fmt"
	"sync"
)

type Registry struct {
	mu        sync.RWMutex
	providers map[string]Provider
}

var (
	globalRegistry *Registry
	once           sync.Once
)

func GetRegistry() *Registry {
	once.Do(func() {
		globalRegistry = &Registry{
			providers: make(map[string]Provider),
		}
	})
	return globalRegistry
}

func (r *Registry) Register(p Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[p.ID()] = p
}

func (r *Registry) Get(id string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.providers[id]
	if !ok {
		return nil, fmt.Errorf("provider %s not found", id)
	}
	return p, nil
}

func (r *Registry) List() []Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]Provider, 0, len(r.providers))
	for _, p := range r.providers {
		list = append(list, p)
	}
	return list
}
