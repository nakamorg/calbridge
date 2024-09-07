package backend

import (
	"context"
	"encoding/json"

	bolt "go.etcd.io/bbolt"
)

type BoltBackend struct {
	db *bolt.DB
}

func NewBoltBackend(dbPath string) (Backend, error) {
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		return nil, err
	}

	return &BoltBackend{db: db}, nil
}

func (bb *BoltBackend) Get(ctx context.Context, data Data) (Data, error) {
	key := bb.key(data)
	err := bb.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(data.User))
		if b == nil {
			return nil
		}

		v := b.Get(key)
		if v == nil {
			return nil
		}

		if err := json.Unmarshal(v, &data); err != nil {
			return err
		}
		return nil
	})
	return data, err
}

func (bb *BoltBackend) Put(ctx context.Context, data Data) error {
	key := bb.key(data)
	return bb.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(data.User))
		if err != nil {
			return err
		}

		dataBytes, err := json.Marshal(data)
		if err != nil {
			return err
		}

		return b.Put(key, dataBytes)
	})
}

func (bb *BoltBackend) key(data Data) []byte {
	// Create a composite key combining data.UID and data.Hash with a delimiter
	return []byte(data.UID + ":" + data.Hash)
}

func (bb *BoltBackend) Close() error {
	return bb.db.Close()
}
