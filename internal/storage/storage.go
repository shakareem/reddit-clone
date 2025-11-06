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
	Password []byte
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
var ErrUserNotFound = errors.New("user not found")
var ErrInvalidPassword = errors.New("invalid password")

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{map[string]User{}, &sync.RWMutex{}}
}

func (s *InMemoryStorage) GetUser(name, password string) (User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.users[name]
	if !ok {
		return User{}, ErrUserNotFound
	}

	err := bcrypt.CompareHashAndPassword(user.Password, []byte(password))
	if err != nil {
		return User{}, ErrInvalidPassword
	}

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
		Password: hashedPassword,
	}
	s.users[name] = u
	return u, nil
}
