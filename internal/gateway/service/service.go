package service

import (
	pkgconfig "github.com/cossim/coss-server/pkg/config"
)

// Service struct
type Service struct {
}

func New(c *pkgconfig.AppConfig) (s *Service) {
	s = &Service{}
	return s
}
