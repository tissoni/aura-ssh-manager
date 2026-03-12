package keychain

import (
	"encoding/json"
	"github.com/zalando/go-keyring"
)

const service = "aura"

var Bootstrapped = true
var Password string
var Encrypted bool

type Record struct {
	Password   string
	PrivateKey string
	Tags       []string
}

func Open(p string) error {
	// No-op for macOS Keychain
	return nil
}

func Get(host string) (k *Record, err error) {
	val, err := keyring.Get(service, host)
	if err != nil {
		if err == keyring.ErrNotFound {
			return nil, ErrNotFound{Key: host}
		}
		return nil, err
	}

	k = &Record{}
	err = json.Unmarshal([]byte(val), k)

	return k, err
}

func Put(key string, record *Record) error {
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	return keyring.Set(service, key, string(data))
}

func Remove(key string) error {
	return keyring.Delete(service, key)
}

func list() (records map[string]*Record, err error) {
	// Not easily supported by go-keyring without significant overhead.
	// Since we are removing the 'encrypt' command which used this, we return nil.
	return nil, nil
}
