package service

import (
	"context"
	"errors"
	"net/http"
	"wdiet/store"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *Service) Ping(c *gin.Context) {
	l := s.l.Named("Ping")

	if err := s.db.Ping(); err != nil {
		l.Error("could not ping", zap.Error(err)) //what level you use depending on what went wrong?
		c.Status(http.StatusInternalServerError)
		return
	} //디비에서부터 에러 뜨면 걍 여기서 리턴해라

	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

func (s *Service) GetUser(c *gin.Context) {
	l := s.l.Named("GetUser")

	id := c.Param("id")

	uid, err := uuid.Parse(id)
	if err != nil {
		l.Error("could not get user", zap.Error(err)) //error message shouldn't contain single quote(') cause it might break. Spacebar is okay.
		c.Status(http.StatusBadRequest)
		return
	}

	user, err := s.db.GetUser(context.Background(), uid)
	if err != nil {
		l.Error("could not get user", zap.Error(err))
		if errors.Is(err, store.ErrNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, dbUser2ApiUser(*user))
}
