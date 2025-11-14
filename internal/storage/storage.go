package storage

type Storage interface {
	UserStorage
	PostStorage
}

type InMemoryStorage struct {
	*UserInMemStorage
	*PostInMemStorage
}

func NewInMemStorage() InMemoryStorage {
	return InMemoryStorage{
		NewUserInMemStorage(),
		NewPostInMemStorage(),
	}
}
