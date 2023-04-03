package events

import (
	"encoding/json"
	"reflect"
	"time"

	"github.com/go-mysql-org/go-mysql/canal"
	"github.com/go-redis/redis"
	"github.com/irononet/go-exchange/conf"
	"github.com/irononet/go-exchange/entities"
	"github.com/irononet/go-exchange/utils"
	"github.com/shopspring/decimal"
	"github.com/siddontang/go-log/log"
)

const (
	TopicOrder = "orders" 
	TopicAccount = "accounts" 
	TopicFill = "fills" 
	TopicBill = "bills"
)

type BinLogStream struct{
	canal.DummyEventHandler
	redisClient *redis.Client
}

func NewBinLogStream() *BinLogStream{
	gexConfig := conf.GetConfig() 

	redisClient := redis.NewClient(&redis.Options{
		Addr: gexConfig.Redis.Addr, 
		Password: gexConfig.Redis.Password, 
		DB: 0, 
	})

	return &BinLogStream{
		redisClient: redisClient,
	}
}

func (s *BinLogStream) OnRow(e *canal.RowsEvent) error{
	switch e.Table.Name{
	case "orders": 
		if e.Action == "delete"{
			return nil 
		}

		var n = 0 
		if e.Action == "update"{
			n = 1 
		}

		var v entities.Order 
		s.parseRow(e, e.Rows[n], &v) 

		buf, _ := json.Marshal(v) 

		ret := s.redisClient.Publish(TopicOrder, buf)
		if ret.Err() != nil{
			log.Error(ret.Err())
		}

	case "fills": 
		if e.Action == "delete" || e.Action == "update"{
			return nil 
		}

		var v entities.Fill 
		s.parseRow(e, e.Rows[0], &v) 

		buf, _ := json.Marshal(v) 
		ret := s.redisClient.LPush(TopicFill, buf) 
		if ret.Err() != nil{
			log.Error(ret.Err())
		}
	case "bills": 
		if e.Action == "delete" || e.Action == "update"{
			return nil
		}

		var v entities.Bill 
		s.parseRow(e, e.Rows[0], &v) 

		buf, _ := json.Marshal(v) 
		ret := s.redisClient.LPush(TopicBill, buf) 
		if ret.Err() != nil{
			log.Error(ret.Err())
		}
	}
	return nil 
}

func (s *BinLogStream) parseRow(e *canal.RowsEvent, row []interface{}, dest interface{}){
	v := reflect.ValueOf(dest).Elem() 
	t := v.Type() 

	for i := 0; i < v.NumField(); i++{
		f := v.Field(i) 

		colIdx := s.getColumnIndexByName(e, utils.SnakeCase(t.Field(i).Name)) 
		rowVal := row[colIdx] 

		switch f.Type().Name(){
		case "int64": 
			f.SetInt(rowVal.(int64))
		case "string": 
			f.SetString(rowVal.(string)) 
		case "bool": 
			if rowVal.(int8) == 0{
				f.SetBool(false) 
			} else{
				f.SetBool(true)
			}
		case "Time": 
			if rowVal != nil{
				f.Set(reflect.ValueOf(rowVal.(time.Time)))
			}
		case "Decimal": 
			d := decimal.NewFromFloat(rowVal.(float64)) 
			f.Set(reflect.ValueOf(d)) 
		default: 
			f.SetString(rowVal.(string))
		}

	}
}

func (s *BinLogStream) getColumnIndexByName(e *canal.RowsEvent, name string) int{
	for id, value := range e.Table.Columns{
		if value.Name == name{
			return id
		}
	}
	return -1
}

func (s *BinLogStream) Start(){
	gexConfig := conf.GetConfig() 
	cfg := canal.NewDefaultConfig() 
	cfg.Addr = gexConfig.DataSource.Addr 
	cfg.User = gexConfig.DataSource.User 
	cfg.Password = gexConfig.DataSource.Password 
	cfg.Dump.ExecutionPath = "" 
	cfg.Dump.TableDB = gexConfig.DataSource.Database 
	cfg.ParseTime = true 
	cfg.IncludeTableRegex = []string{gexConfig.DataSource.Database + "\\..*"} 
	cfg.ExcludeTableRegex = []string{"mysql\\..*"} 
	c, err := canal.NewCanal(cfg) 
	if err != nil{
		panic(err) 
	}
	c.SetEventHandler(s) 

	pos, err := c.GetMasterPos() 
	if err != nil{
		panic(err)
	}
	err = c.RunFrom(pos) 
	if err != nil{
		panic(err)
	}
}