package main

type Storage interface {
	Save(userID string) error
	Delete(userID, name string) error
	List(userID string) error
	GetByName(userID, name string) (*Account, error)
	Init() error
	Close() error
}
