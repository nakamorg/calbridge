package backend

import (
	"context"
	"time"
)

type Direction string

const (
	DirectionOut Direction = "out"
	DirectionIn  Direction = "in"
)

type Data struct {
	User       string    `json:"user"`
	UID        string    `json:"uid"`
	Hash       string    `json:"hash"`
	Direction  Direction `json:"direction"`
	SyncedTime time.Time `json:"synced_time"`
	Synced     bool      `json:"synced"`
}

type Backend interface {
	Get(ctx context.Context, data Data) (Data, error)
	Put(ctx context.Context, data Data) error
	Close() error
}

type DummyBackend struct{}

// NewDummyBackend returns a Backend. This backend does not store any data and always returns
// nil for error and echoes back the data
func NewDummyBackend() Backend {
	return &DummyBackend{}
}

func (b *DummyBackend) Get(ctx context.Context, data Data) (Data, error) {
	return data, nil
}

func (b *DummyBackend) Put(ctx context.Context, data Data) error {
	return nil
}

func (b *DummyBackend) Close() error {
	return nil
}
