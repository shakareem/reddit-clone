package storage

import (
	"errors"
	"sync"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       string
	Name     string
	Password string
}

type Storage interface {
	GetUser(name, password string) (User, error)
	AddUser(name, password string) (User, error)
}

type InMemoryStorage struct {
	users map[string]User
	mu    *sync.RWMutex
}

var ErrUserAlreadyExists = errors.New("already exists")

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{map[string]User{}, &sync.RWMutex{}}
}

func (s *InMemoryStorage) GetUser(name, password string) (User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.users[name]
	if !ok {
		return User{}, errors.New("wrong username")
	}

	// TODO: compare password with hash

	return user, nil
}

func (s *InMemoryStorage) AddUser(name, password string) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.users[name]
	if ok {
		return User{}, ErrUserAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}

	u := User{
		ID:       uuid.NewString(),
		Name:     name,
		Password: string(hashedPassword),
	}
	s.users[name] = u
	return u, nil
}
