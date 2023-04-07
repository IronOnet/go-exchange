package matching

import (
	"github.com/irononet/go-exchange/entities"
)

type OrderReader interface {
	SetOffset(offset int64) error

	FetchOrder() (offset int64, order *entities.Order, err error)
}

type LogStore interface {
	Store(logs []interface{}) error
}

type LogReader interface {
	GetProductId() string
	RegisterObserver(observer LogObserver)

	Run(seq, offset int64)
}

type LogObserver interface {
	OnOpenLOg(log *OpenLog, offset int64)

	OnMatchLog(log *MatchLog, offset int64)

	OnDoneLog(log *DoneLog, offset int64)
}

type SnapshotStore interface {
	Store(snapshot *Snapshot) error

	GetLatest() (*Snapshot, error)
}
