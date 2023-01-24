package service

import (
	"wdiet/store"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Service struct {
	r  *gin.Engine
	db store.Store
	l  *zap.Logger
}

func New(s store.Store, l *zap.Logger) *Service {
	newService := &Service{r: gin.Default(), db: s, l: l}

	newService.registerRoutes()

	return newService
}

func (s *Service) Run() {
	l := s.l.Named("Run") //logger specifically created for this function

	if err := s.r.Run(); err != nil {
		l.Fatal("service failed to start", zap.Error(err))
	}
}
