package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/hulb/bolt"
)

// Account is an otpauth account
type Account struct {
	Name     string `json:"name"`
	Label    string `json:"lable"`
	Username string `json:"username"`
	Secret   string `json:"secret"`
	URL      string `json:"url"`
}

// ParseAccount parses a otpauth url to account
// URL should be like:
// otpauth://totp/Google:xxx@gmail.com?secret=6OH6AZFKDFQJB5CTMZLEVSCZCA&issuer=Google
func ParseAccount(URL string) (*Account, error) {
	var err error
	var account *Account

	authURL, err := url.Parse(URL)
	if err != nil {
		return account, err
	}

	switch {
	case authURL.Scheme != "otpauth", authURL.Host != "totp", len(authURL.Path) < 1:
		log.Println(authURL)
		return account, fmt.Errorf("invalid schema")
	}

	pathArr := strings.Split(authURL.Path[1:], ":")
	if len(pathArr) < 2 {
		log.Println(authURL)
		return account, fmt.Errorf("invalid schema")
	}

	label := pathArr[0]
	username := strings.Join(pathArr[1:], ":") // in case `:` exists in username part

	secret := authURL.Query().Get("secret")
	if secret == "" {
		log.Println(authURL)
		return account, fmt.Errorf("invalid schema")
	}

	account = &Account{
		Label:    label,
		Username: username,
		Secret:   secret,
		URL:      authURL.String(),
	}

	return account, err
}

// JSON marshal account to JSON bytes
func (account *Account) JSON() []byte {
	v, err := json.Marshal(account)
	if err != nil {
		panic(err)
	}

	return v
}

func (account *Account) Save() error {
	db, err := bolt.Open("data.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(account.Label))
		if err != nil {
			return err
		}

		return b.Put([]byte(account.Username), account.JSON())
	})
}
