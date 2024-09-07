package backend

import (
	"context"
	"encoding/json"

	bolt "go.etcd.io/bbolt"
)

type BoltBackend struct {
	user string
	db   *bolt.DB
}

func NewBoltBackend(dbPath, user string) (Backend, error) {
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		return nil, err
	}

	// Ensure the bucket exists
	// TODO: Don't think it;s a good idea to create a client per user. As the DB file is
	// shared - these clents will block each/other endlessly. Maybe better to share the client
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(user))
		return err
	})
	if err != nil {
		return nil, err
	}

	return &BoltBackend{db: db, user: user}, nil
}

func (bb *BoltBackend) Get(ctx context.Context, data Data) (Data, error) {
	key := bb.key(data)
	err := bb.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bb.user))
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
		b := tx.Bucket([]byte(bb.user))

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
