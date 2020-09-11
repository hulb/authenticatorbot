package main

import (
	"encoding/json"
	"errors"

	"github.com/boltdb/bolt"
)

// BoltStorage use boltdb as storage
type BoltStorage struct {
	DBPath string
	DB     *bolt.DB
}

// New returns a boltdb based storage instance
func New(dbPath string) (*BoltStorage, error) {
	var (
		bs  *BoltStorage
		err error
	)
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		return bs, err
	}

	return &BoltStorage{
		DBPath: dbPath,
		DB:     db,
	}, nil
}

// Close closes the underlying bolt db connection
func (bs *BoltStorage) Close() error {
	return bs.DB.Close()
}

// Save saves an account to bolt db
func (bs *BoltStorage) Save(userID string, account *Account, override bool) error {
	return bs.DB.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(userID))
		if err != nil {
			return err
		}

		accountsBytes := b.Get([]byte(account.Name))
		if accountsBytes != nil && !override {
			return errors.New("account name already exists")
		}

		return b.Put([]byte(account.Name), account.JSON())
	})
}

// List returns all accounts
func (bs *BoltStorage) List(userID string) ([]*Account, error) {
	var accounts []*Account
	err := bs.DB.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket([]byte(userID)); b != nil {
			c := b.Cursor()
			for k, v := c.First(); k != nil; k, v = c.Next() {
				var account Account
				if err := json.Unmarshal(v, &account); err != nil {
					return err
				}

				accounts = append(accounts, &account)
			}
		}

		return nil
	})

	return accounts, err
}

// GetByName get account by name
func (bs *BoltStorage) GetByName(userID, name string) (*Account, error) {
	var account *Account
	err := bs.DB.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket([]byte(userID)); b != nil {
			if v := b.Get([]byte(name)); v != nil {
				if err := json.Unmarshal(v, account); err != nil {
					return err
				}
			}
		}
		return nil
	})

	return account, err
}

// Delete delete account by name
func (bs *BoltStorage) Delete(userID, name string) error {
	return bs.DB.Update(func(tx *bolt.Tx) error {
		if b := tx.Bucket([]byte(userID)); b != nil {
			return b.Delete([]byte(name))
		}
		return nil
	})
}
