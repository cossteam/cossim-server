package service

import (
	"github.com/cossim/coss-server/pkg/config"
)

// Service struct
type Service struct {
}

func New(c *config.AppConfig) (s *Service) {
	s = &Service{}
	return s
}
