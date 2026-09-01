package repository

import (
	"context"
	"sync"
)

// InMemoryServiceRegistrationStore stores service registrations in memory.
//
// This implementation is intended for local development and testing.
// Registrations are lost when the process restarts.
type InMemoryServiceRegistrationStore struct {
	mu            sync.RWMutex
	registrations map[string]ServiceRegistration
}

func NewInMemoryServiceRegistrationStore() *InMemoryServiceRegistrationStore {
	return &InMemoryServiceRegistrationStore{
		registrations: make(map[string]ServiceRegistration),
	}
}

func (s *InMemoryServiceRegistrationStore) Get(
	_ context.Context,
	project string,
	serviceName string,
) (ServiceRegistration, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	registration, exists := s.registrations[registrationKey(
		project,
		serviceName,
	)]

	return registration, exists, nil
}

func (s *InMemoryServiceRegistrationStore) Save(
	_ context.Context,
	registration ServiceRegistration,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.registrations[registrationKey(
		registration.Project,
		registration.ServiceName,
	)] = registration

	return nil
}

func registrationKey(
	project string,
	serviceName string,
) string {
	return project + "/" + serviceName
}
