package host

import (
	"sync"

	"github.com/google/uuid"
)

type SimplePathResolver struct {
	pathMap map[string]uuid.UUID
	mtx     *sync.Mutex
}

func NewSimplePathResolver() *SimplePathResolver {
	return &SimplePathResolver{
		pathMap: make(map[string]uuid.UUID),
		mtx:     new(sync.Mutex),
	}
}

func (r *SimplePathResolver) TrySetMapping(path string, dest uuid.UUID) bool {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	if _, ok := r.pathMap[path]; ok {
		return false
	}
	r.pathMap[path] = dest
	return true
}

func (r *SimplePathResolver) DeleteMapping(path string) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	delete(r.pathMap, path)
}

func (r *SimplePathResolver) PathToSessionID(path string, _ string) (uuid.UUID, bool) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	res, ok := r.pathMap[path]
	return res, ok
}
