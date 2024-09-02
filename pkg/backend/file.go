package backend

import (
	"context"
	"encoding/csv"
	"os"
	"sync"
	"time"
)

type FileBackend struct {
	mu       sync.Mutex
	filePath string
}

func NewFileBackend(filePath string) Backend {
	return &FileBackend{filePath: filePath}
}

func (fb *FileBackend) Get(ctx context.Context, data Data) (Data, error) {
	fb.mu.Lock()
	defer fb.mu.Unlock()

	file, err := os.Open(fb.filePath)
	if err != nil {
		return data, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return data, err
	}

	for _, record := range records {
		if record[0] == data.User && record[1] == data.UID && record[2] == data.Hash {
			data.Direction = Direction(record[3])
			data.SyncedTime, _ = time.Parse(time.RFC3339, record[4])
			data.Synced = record[5] == "true"
			return data, nil
		}
	}
	// return original data if not found in the backend
	return data, nil
}

func (fb *FileBackend) Put(ctx context.Context, data Data) error {
	fb.mu.Lock()
	defer fb.mu.Unlock()

	file, err := os.OpenFile(fb.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	record := []string{
		data.User,
		data.UID,
		data.Hash,
		string(data.Direction),
		data.SyncedTime.Format(time.RFC3339),
		"",
	}
	if data.Synced {
		record[5] = "true"
	} else {
		record[5] = "false"
	}

	writer.Write(record)
	writer.Flush()
	return writer.Error()
}

func (fb *FileBackend) Delete(ctx context.Context, data Data) error {
	fb.mu.Lock()
	defer fb.mu.Unlock()

	file, err := os.OpenFile(fb.filePath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	records, err := csv.NewReader(file).ReadAll()
	if err != nil {
		return err
	}

	for i, record := range records {
		if record[0] == data.User && record[1] == data.UID && record[2] == data.Hash {
			records = append(records[:i], records[i+1:]...)
			break
		}
	}

	file.Seek(0, 0)
	file.Truncate(0)
	writer := csv.NewWriter(file)
	writer.WriteAll(records)
	writer.Flush()
	return writer.Error()
}
