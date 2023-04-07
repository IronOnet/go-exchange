package matching 

import (
	"context"
	"encoding/json" 
	"github.com/irononet/go-exchange/entities" 
	"github.com/segmentio/kafka-go"
)

const (
	TopicOrderPrefix = "matching_order_"
)

type KafkaOrderReader struct{
	OrderReader *kafka.Reader 
}

func NewKafkaOrderReader(productId string, brokers []string) *KafkaOrderReader{
	s := &KafkaOrderReader{} 

	s.OrderReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers, 
		Topic: TopicOrderPrefix + productId, 
		Partition: 0, 
		MinBytes: 1, 
		MaxBytes: 10e6, 
	})

	return s
}

func (s *KafkaOrderReader) SetOffset(offset int64) error{
	return s.OrderReader.SetOffset(offset)
}

func (s *KafkaOrderReader) FetchOrder() (offset int64, order *entities.Order, err error){
	message, err := s.OrderReader.FetchMessage(context.Background()) 
	if err != nil{
		return 0, nil, err 
	}

	err = json.Unmarshal(message.Value, &order) 
	if err != nil{
		return 0, nil, err
	}

	return message.Offset, order, nil 
}