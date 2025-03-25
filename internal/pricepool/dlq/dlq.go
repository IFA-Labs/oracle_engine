package dlq

import (
	"encoding/json"
	"oracle_engine/internal/logging"
	"oracle_engine/internal/models"
	"sync"

	"go.uber.org/zap"
)

type DLQ struct {
	queue []Entry
	mu    sync.Mutex
}

type Entry struct {
	Price models.Price
	Error error
}

func NewDLQ() *DLQ {
	return &DLQ{}
}

func (d *DLQ) Enqueue(price models.Price, err error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	entry := Entry{Price: price, Error: err}
	d.queue = append(d.queue, entry)

	// TODO: Persist to file/DB if needed
	data, _ := json.Marshal(entry)
	logging.Logger.Warn("DLQ entry", zap.String("entry", string(data)))
}
