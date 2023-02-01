package service

func (s *Service) registerRoutes() {
	s.r.GET("/ping", s.Ping)

	s.r.GET("/users/:id", s.GetUser)
	s.r.POST("/users", s.CreateUser)
	s.r.POST("/users/:id", s.UpdateUser)

	s.r.GET("/ingredients/:id", s.GetIngredient)
	s.r.POST("/ingredients", s.CreateIngredient)
	s.r.POST("/ingredients/:id", s.UpdateIngredient)
	s.r.POST("/ingredients/search", s.SearchIngredients)
	s.r.DELETE("/ingredients/:id", s.DeleteIngredient)

}
