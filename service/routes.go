package service

func (s *Service) registerRoutes() {
	s.r.GET("/ping", s.Ping)

	s.r.GET("/users/:id", s.GetUser)
	s.r.POST("/users", s.CreateUser)
	s.r.POST("/users/:id", s.UpdateUser)
}
