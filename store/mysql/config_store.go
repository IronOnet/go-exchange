package mysql

import "github.com/irononet/go-exchange/entities"

func (s *Store) GetConfigs() ([]*entities.Config, error) {
	var configs []*entities.Config
	err := s.db.Find(&configs).Error
	return configs, err
}
