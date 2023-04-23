package mysql

import (
	"os"
	"reflect"
	"sync"

	"github.com/irononet/go-exchange/conf"
	"github.com/irononet/go-exchange/entities"
	"github.com/irononet/go-exchange/store"
	"github.com/shopspring/decimal"

	//log "github.com/prometheus/common"
	"github.com/siddontang/go-log/log"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
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

	if os.Getenv("GEX_ENV") == "dev" {
		gexDB, err := gorm.Open(sqlite.Open(cfg.TestDataSource.Database), &gorm.Config{})
		if err != nil {
			return err
		}

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
			//log.Infof("migating database, table: %v", reflect.TypeOf(table))
			if err = gexDB.AutoMigrate(table); err != nil {
				return err
			}
		}

		// Create mock data here
		productBTCUSD := &entities.Product{
			BaseCurrency: "USD", 
			QuoteCurrency: "BTC", 
			BaseMinSize: decimal.NewFromFloat(0.00001), 
			BaseMaxSize: decimal.NewFromFloat(10000.0), 
			QuoteMaxSize: decimal.NewFromFloat(4), 
			QuoteMinSize: decimal.NewFromFloat(2),
			BaseScale: 4, 
			QuoteScale: 2,
		}

		productBTCUSDT := &entities.Product{
			BaseCurrency: "USDT", 
			QuoteCurrency: "BTC", 
			BaseMinSize: decimal.NewFromFloat(0.00001), 
			BaseMaxSize: decimal.NewFromFloat(10000.0), 
			QuoteMaxSize: decimal.NewFromFloat(4), 
			QuoteMinSize: decimal.NewFromFloat(2),
			BaseScale: 4, 
			QuoteScale: 2,
		}

		gexDB.Create(&productBTCUSD) 
		gexDB.Create(&productBTCUSDT)

	} else if os.Getenv("GEX_ENV") == "prod" {
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
				//log.Infof("migating database, table: %v", reflect.TypeOf(table))
				if err = gexDB.AutoMigrate(table); err != nil {
					return err
				}
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
