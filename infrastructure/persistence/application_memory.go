// infrastructure/persistence/application_memory.go
package persistence

import (
	"errors"
	"sync"

	"github.com/shinchi-pmtech/ringi/domain/application"
)

type ApplicationMemoryRepository struct {
	mu    sync.RWMutex
	store map[application.ApplicationID]*application.Application
}

func NewApplicationMemoryRepository() *ApplicationMemoryRepository {
	return &ApplicationMemoryRepository{
		store: map[application.ApplicationID]*application.Application{},
	}
}

func (r *ApplicationMemoryRepository) Save(app *application.Application) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.store[app.ID()] = app
	return nil
}

func (r *ApplicationMemoryRepository) FindByID(id application.ApplicationID) (*application.Application, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	app, ok := r.store[id]
	if !ok {
		return nil, errors.New("application not found")
	}
	return app, nil
}
