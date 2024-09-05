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
	User       string
	UID        string
	Hash       string
	Direction  Direction
	SyncedTime time.Time
	Synced     bool
}

type Backend interface {
	Get(ctx context.Context, data Data) (Data, error)
	Put(ctx context.Context, data Data) error
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
