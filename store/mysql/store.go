package mysql

import (
	"reflect"
	"sync"

	"github.com/irononet/go-exchange/conf"
	"github.com/irononet/go-exchange/entities"
	"github.com/irononet/go-exchange/store"
	"github.com/prometheus/common/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var gexDB *gorm.DB
var gexStore store.Store
var storeOnce sync.Once

type Store struct {
	db *gorm.DB
}

func SharedStore() store.Store {
	storeOnce.Do(func() {
		err := initDb()
		if err != nil {
			panic(err)
		}
		gexStore = NewStore(gexDB)
	})
	return gexStore
}

func NewStore(db *gorm.DB) *Store {
	return &Store{
		db: db,
	}
}

func initDb() error {
	cfg := conf.GetConfig()

	gexDB, err := gorm.Open(mysql.Open(cfg.DataSource.Addr), &gorm.Config{})
	if err != nil {
		return err
	}

	if cfg.DataSource.EnableAutoMigrate {
		var tables = []interface{}{
			&entities.Account{},
			&entities.Order{},
			&entities.Product{},
			&entities.Trade{},
			&entities.Fill{},
			&entities.User{},
			&entities.Bill{},
			&entities.Tick{},
			&entities.Config{},
		}

		for _, table := range tables {
			log.Infof("migrating database, table :%v", reflect.TypeOf(table))
			if err = gexDB.AutoMigrate(table); err != nil {
				return err
			}
		}

	}

	return nil
}

func (s *Store) BeginTx() (store.Store, error) {
	db := s.db.Begin()
	if db.Error != nil {
		return nil, db.Error
	}
	return NewStore(db), nil
}

func (s *Store) Rollback() error {
	return s.db.Rollback().Error
}

func (s *Store) CommitTx() error {
	return s.db.Commit().Error
}
