package service

import (
	"im/pkg/config"
)

// Service struct
type Service struct {
}

func New(c *config.AppConfig) (s *Service) {
	s = &Service{}
	return s
}
