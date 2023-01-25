package service

func (s *Service) registerRoutes() {
	s.r.GET("/ping", s.Ping)

	s.r.GET("/user/:id", s.GetUser)
}
