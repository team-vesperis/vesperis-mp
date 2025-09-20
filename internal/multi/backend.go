package multi

import (
	"go.minekube.com/gate/pkg/util/uuid"
)

type MultiBackend struct {
	id uuid.UUID

	mp *MultiProxy
}

func (mb *MultiBackend) GetId() uuid.UUID {
	return mb.id
}

// return the multiproxy the multibackend is located under
func (mb *MultiBackend) GetMultiProxy() *MultiProxy {
	return mb.mp
}
