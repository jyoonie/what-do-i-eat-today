package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
	"wdiet/store"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
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

func (s *Service) Login(c *gin.Context) {
	l := s.l.Named("Login")

	var loginRequest Login

	if err := json.NewDecoder(c.Request.Body).Decode(&loginRequest); err != nil { //decode 한다음에 그 내용이 valid한지 비교해야지 바보야.. 저 위에 var loginRequest create 한거는 새로 생긴거자나.. 으이구
		l.Info("error logging in", zap.Error(err))
		c.Status(http.StatusBadRequest)
		return
	}

	if !isValidLoginRequest(loginRequest) {
		l.Info("error logging in")
		c.Status(http.StatusBadRequest)
		return
	}

	user, err := s.db.GetUserByEmail(context.Background(), loginRequest.EmailAddress)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			l.Info("error logging in", zap.Error(err))
			c.Status(http.StatusNotFound)
			return
		}
		l.Error("error logging in", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(loginRequest.Password)); err != nil {
		l.Info("error logging in", zap.Error(err))
		c.Status(http.StatusBadRequest)
		return
	}

	// Create the claims
	claims := jwt.RegisteredClaims{
		// A usual scenario is to set the expiration time relative to the current time
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Issuer:    "whatDoIEatToday",
		ID:        user.UserUUID.String(),
		Audience:  []string{"whatDoIEatToday"},
	}

	token := jwt.NewWithClaims(s.mySigningMethod, claims)
	signedToken, err := token.SignedString(s.mySigningKey)
	if err != nil {
		l.Error("error signing the token")
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, Token{Token: signedToken})
}

func (s *Service) ValidateToken(c *gin.Context) {
	token := c.Request.Header.Get("Authorization")
	if token == "" {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if len(strings.Split(token, " ")) < 2 {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	} //return 했으므로 else 할 필요없음. return에 안걸리면 어차피 else에 해당하는 부분은 continue 되기 때문에.

	realToken := strings.Split(token, " ")[1]

	t, err := jwt.Parse(realToken, //getting rid of "bearer " from the original token
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return s.mySigningKey, nil //Parse method is going to use this function to reencrypt the body, so the first time you create the jwt, you know the first part is alg, second claim, third encrypted.
		},
		jwt.WithValidMethods([]string{s.mySigningMethod.Alg()}),
	)
	if err != nil || !t.Valid {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	c.Status(http.StatusOK)
}

func (s *Service) GetUser(c *gin.Context) {
	l := s.l.Named("GetUser")

	id := c.Param("id")

	uid, err := uuid.Parse(id)
	if err != nil {
		l.Info("error getting user", zap.Error(err)) //error message shouldn't contain single quote(') cause it might break. Spacebar is okay.
		c.Status(http.StatusBadRequest)
		return
	}

	user, err := s.db.GetUser(context.Background(), uid)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			l.Info("error getting user", zap.Error(err))
			c.Status(http.StatusNotFound)
			return
		}
		l.Error("error getting user", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, dbUser2ApiUser(user))
}

func (s *Service) CreateUser(c *gin.Context) {
	l := s.l.Named("CreateUser")

	var createUserRequest struct { //embedding User struct
		User
		Password string `json:"password,omitempty"` //you only need to use this field once at this one spot(CreateUser)
	}

	if err := json.NewDecoder(c.Request.Body).Decode(&createUserRequest); err != nil {
		l.Info("error creating user", zap.Error(err))
		c.Status(http.StatusBadRequest)
		return
	}

	if !isValidCreateUserRequest(createUserRequest.User, createUserRequest.Password) {
		l.Info("error creating user")
		c.Status(http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(createUserRequest.Password), bcrypt.MinCost)
	if err != nil {
		l.Error("error generating hashed password", zap.Error(err)) //"unexpected error ..."는 테스트 할 때 써라.
		c.Status(http.StatusInternalServerError)
		return
	}

	u := apiUser2DBUser(createUserRequest.User)
	u.HashedPassword = string(hashedPassword)

	user, err := s.db.CreateUser(context.Background(), u) //이렇게 에러 처리해놓으면 굳이 db에서 *store.User를 리턴할 필요는 없지만, 이 에러처리를 까먹는 개발자도 있음..
	if err != nil {
		l.Error("error creating user", zap.Error(err)) //error creating user, user.UUID(create 실패한 user인데 이 user의 UUID를 에러 메시지로 반환하려고;;) 이렇게 zero value 필드에 접근하는 실수를 하는 개발자도 있다고..;; 그래서 위를 포함한 이러한 상황에 대비해 *store.User로 리턴하는 것임..
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, dbUser2ApiUser(user))
}

func (s *Service) UpdateUser(c *gin.Context) {
	l := s.l.Named("UpdateUser")

	id := c.Param("id")

	uid, err := uuid.Parse(id)
	if err != nil {
		l.Info("error updating user", zap.Error(err))
		c.Status(http.StatusBadRequest)
		return
	}

	var updateUserRequest User

	if err := json.NewDecoder(c.Request.Body).Decode(&updateUserRequest); err != nil {
		l.Info("error updating user", zap.Error(err))
		c.Status(http.StatusBadRequest)
		return
	}

	if !isValidUpdateUserRequest(updateUserRequest, uid) {
		l.Info("error updating user")
		c.Status(http.StatusBadRequest)
		return
	}

	user, err := s.db.UpdateUser(context.Background(), apiUser2DBUser(updateUserRequest)) //if I have two variables, I can still do combined if statement, like if user, err ... ; err != nil {}, but then user can only survive within the next 3 lines of if statement. So I can't return user variable at the bottom in c.JSON().
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			l.Info("error updating user", zap.Error(err))
			c.Status(http.StatusNotFound)
			return
		}
		l.Error("error updating user", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, dbUser2ApiUser(user))
}

func (s *Service) GetIngredient(c *gin.Context) {
	l := s.l.Named("GetIngredient")

	id := c.Param("id")

	iid, err := uuid.Parse(id)
	if err != nil {
		l.Info("error getting ingredient", zap.Error(err))
		c.Status(http.StatusBadRequest)
		return
	}

	ingredient, err := s.db.GetIngredient(context.Background(), iid)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			l.Info("error getting ingredient", zap.Error(err))
			c.Status(http.StatusNotFound)
			return
		}
		l.Error("error getting ingredient", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, dbIngr2ApiIngr(ingredient))
}

func (s *Service) SearchIngredients(c *gin.Context) {
	l := s.l.Named("SearchIngredients")

	var searchIngrRequest SearchIngredient

	if err := json.NewDecoder(c.Request.Body).Decode(&searchIngrRequest); err != nil {
		l.Info("error searching ingredients", zap.Error(err))
		c.Status(http.StatusBadRequest)
		return
	}

	if !isValidSearchIngrRequest(searchIngrRequest) {
		l.Info("error searching ingredients")
		c.Status(http.StatusBadRequest)
		return
	}

	ingredients, err := s.db.SearchIngredients(context.Background(), apiSearchIngr2DBSearchIngr(searchIngrRequest))
	if err != nil {
		// if errors.Is(err, store.ErrNotFound) {
		// 	l.Info("error searching ingredients", zap.Error(err))
		// 	c.Status(http.StatusNotFound)
		// 	return
		// }
		l.Error("error searching ingredients", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	if len(ingredients) == 0 {
		c.Status(http.StatusOK)
		return
	}

	var searchIngrResponse []Ingredient

	for _, ingredient := range ingredients {
		i := dbIngr2ApiIngr(&ingredient)
		searchIngrResponse = append(searchIngrResponse, i)
	}

	c.JSON(http.StatusOK, searchIngrResponse)
}

func (s *Service) CreateIngredient(c *gin.Context) {
	l := s.l.Named("CreateIngredient")

	var createIngrRequest Ingredient

	if err := json.NewDecoder(c.Request.Body).Decode(&createIngrRequest); err != nil {
		l.Info("error creating ingredient", zap.Error(err))
		c.Status(http.StatusBadRequest)
		return
	}

	if !isValidCreateIngrRequest(createIngrRequest) {
		l.Info("error creating ingredient")
		c.Status(http.StatusBadRequest)
		return
	}

	ingredient, err := s.db.CreateIngredient(context.Background(), apiIngr2DBIngr(createIngrRequest))
	if err != nil {
		l.Error("error creating ingredient", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, dbIngr2ApiIngr(ingredient))
}

func (s *Service) UpdateIngredient(c *gin.Context) {
	l := s.l.Named("UpdateIngredient")

	id := c.Param("id")

	iid, err := uuid.Parse(id)
	if err != nil {
		l.Info("error updating ingredient", zap.Error(err)) //error message shouldn't contain single quote(') cause it might break. Spacebar is okay.
		c.Status(http.StatusBadRequest)
		return
	}

	var updateIngrRequest Ingredient

	if err := json.NewDecoder(c.Request.Body).Decode(&updateIngrRequest); err != nil {
		l.Info("error updating ingredient", zap.Error(err))
		c.Status(http.StatusBadRequest)
		return
	}

	if !isValidUpdateIngrRequest(updateIngrRequest, iid) {
		l.Info("error updating ingredient")
		c.Status(http.StatusBadRequest)
		return
	}

	ingredient, err := s.db.UpdateIngredient(context.Background(), apiIngr2DBIngr(updateIngrRequest)) //if I have two variables, I can still do combined if statement, like if user, err ... ; err != nil {}, but then user can only survive within the next 3 lines of if statement. So I can't return user variable at the bottom in c.JSON().
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			l.Info("error updating ingredient", zap.Error(err))
			c.Status(http.StatusNotFound)
			return
		}
		l.Error("error updating ingredient", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, dbIngr2ApiIngr(ingredient))
}

func (s *Service) DeleteIngredient(c *gin.Context) {
	l := s.l.Named("DeleteIngredient")

	id := c.Param("id")

	iid, err := uuid.Parse(id)
	if err != nil {
		l.Info("error deleting ingredient", zap.Error(err))
		c.Status(http.StatusBadRequest)
		return
	}

	if err = s.db.DeleteIngredient(context.Background(), iid); err != nil {
		l.Error("error deleting ingredient", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

func (s *Service) ListFridgeIngredients(c *gin.Context) {
	l := s.l.Named("ListFridgeIngredients")

	id := c.Param("id")

	uid, err := uuid.Parse(id)
	if err != nil {
		l.Info("error listing fridge ingredients", zap.Error(err))
		c.Status(http.StatusBadRequest)
		return
	}

	fridgeIngredients, err := s.db.ListFridgeIngredients(context.Background(), uid)
	if err != nil {
		// if errors.Is(err, store.ErrNotFound) {
		// 	l.Info("error listing fridge ingredients", zap.Error(err))
		// 	c.Status(http.StatusNotFound)
		// 	return
		// }
		l.Error("error listing fridge ingredients", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	if len(fridgeIngredients) == 0 {
		c.Status(http.StatusOK)
		return
	}

	var listFIngrResponse []FridgeIngredient

	for _, f := range fridgeIngredients {
		fridgeIngredient := dbFIngr2ApiFIngr(&f)
		listFIngrResponse = append(listFIngrResponse, fridgeIngredient)
	}

	c.JSON(http.StatusOK, listFIngrResponse)
}

func (s *Service) CreateFridgeIngredient(c *gin.Context) {
	l := s.l.Named("CreateFridge")

	var createFIngrRequest FridgeIngredient

	if err := json.NewDecoder(c.Request.Body).Decode(&createFIngrRequest); err != nil {
		l.Info("error creating fridge ingredient", zap.Error(err))
		c.Status(http.StatusBadRequest)
		return
	}

	if !isValidCreateFIngrRequest(createFIngrRequest) {
		l.Info("error creating fridge ingredient")
		c.Status(http.StatusBadRequest)
		return
	}

	ingredient, err := s.db.GetIngredient(context.Background(), createFIngrRequest.IngredientUUID)
	if err != nil {
		l.Error("error creating fridge ingredient", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	createFIngrRequest.ExpirationDate = createFIngrRequest.PurchasedDate.Add(24 * time.Hour * time.Duration(ingredient.DaysUntilExp))

	fridgeIngredient, err := s.db.CreateFridgeIngredient(context.Background(), apiFIngr2DBFIngr(createFIngrRequest))
	if err != nil {
		l.Error("error creating fridge ingredient", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, dbFIngr2ApiFIngr(fridgeIngredient))
}

func (s *Service) UpdateFridgeIngredient(c *gin.Context) {
	l := s.l.Named("UpdateFridgeIngredient")

	id := c.Param("id")

	uid, err := uuid.Parse(id)
	if err != nil {
		l.Info("error updating ingredient", zap.Error(err))
		c.Status(http.StatusBadRequest)
		return
	}

	var updateFIngrRequest FridgeIngredient

	if err := json.NewDecoder(c.Request.Body).Decode(&updateFIngrRequest); err != nil {
		l.Info("error updating fridge ingredient", zap.Error(err))
		c.Status(http.StatusBadRequest)
		return
	}

	if !isValidUpdateFIngrRequest(updateFIngrRequest, uid) {
		l.Info("error updating fridge ingredient")
		c.Status(http.StatusBadRequest)
		return
	}

	ingredient, err := s.db.GetIngredient(context.Background(), updateFIngrRequest.IngredientUUID)
	if err != nil {
		l.Error("error upating fridge ingredient", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	updateFIngrRequest.ExpirationDate = updateFIngrRequest.PurchasedDate.Add(24 * time.Hour * time.Duration(ingredient.DaysUntilExp))

	fridgeIngredient, err := s.db.UpdateFridgeIngredient(context.Background(), apiFIngr2DBFIngr(updateFIngrRequest))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			l.Info("error updating fridge ingredient", zap.Error(err))
			c.Status(http.StatusNotFound)
			return
		}
		l.Error("error updating fridge ingredient", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, dbFIngr2ApiFIngr(fridgeIngredient))
}

func (s *Service) DeleteFridgeIngredient(c *gin.Context) {
	l := s.l.Named("DeleteFridgeIngredient")

	id := c.Param("uid")
	id2 := c.Param("fid")

	uid, err := uuid.Parse(id)
	if err != nil {
		l.Info("error deleting fridge ingredient", zap.Error(err))
		c.Status(http.StatusBadRequest)
		return
	}
	fid, err := uuid.Parse(id2)
	if err != nil {
		l.Info("error deleting fridge ingredient", zap.Error(err))
		c.Status(http.StatusBadRequest)
		return
	}

	// var deleteFIngrRequest DeleteFIngr

	// if err := json.NewDecoder(c.Request.Body).Decode(&deleteFIngrRequest); err != nil {
	// 	l.Info("error deleting fridge ingredient", zap.Error(err))
	// 	c.Status(http.StatusBadRequest)
	// 	return
	// }

	// if !isValidDeleteFIngrRequest(deleteFIngrRequest, uid) {
	// 	l.Info("error deleting fridge ingredient")
	// 	c.Status(http.StatusBadRequest)
	// 	return
	// }

	if err := s.db.DeleteFridgeIngredient(context.Background(), uid, fid); err != nil {
		l.Error("error deleting fridge ingredient", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

func (s *Service) GetRecipe(c *gin.Context) {
	l := s.l.Named("GetRecipe")

	id := c.Param("id")

	rid, err := uuid.Parse(id)
	if err != nil {
		l.Info("error getting recipe", zap.Error(err))
		c.Status(http.StatusBadRequest)
		return
	}

	recipe, err := s.db.GetRecipe(context.Background(), rid)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			l.Info("error getting recipe", zap.Error(err))
			c.Status(http.StatusNotFound)
			return
		}
		l.Error("error getting recipe", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, dbRecipe2ApiRecipe(recipe))
}

func (s *Service) ListRecipes(c *gin.Context) {
	l := s.l.Named("ListRecipes")

	id := c.Param("id")

	uid, err := uuid.Parse(id)
	if err != nil {
		l.Info("error listing recipes", zap.Error(err))
		c.Status(http.StatusBadRequest)
		return
	}

	recipes, err := s.db.ListRecipes(context.Background(), uid)
	if err != nil {
		l.Error("error listing recipes", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	if len(recipes) == 0 {
		c.Status(http.StatusOK)
		return
	}

	var listRecipesResponse []Recipe

	for _, recipe := range recipes {
		r := dbRecipe2ApiRecipe(&recipe)
		listRecipesResponse = append(listRecipesResponse, r)
	}

	c.JSON(http.StatusOK, listRecipesResponse)
}

func (s *Service) SearchRecipes(c *gin.Context) {
	l := s.l.Named("SearchRecipes")

	var searchRecipesRequest SearchRecipes

	if err := json.NewDecoder(c.Request.Body).Decode(&searchRecipesRequest); err != nil {
		l.Info("error searching recipes", zap.Error(err))
		c.Status(http.StatusBadRequest)
		return
	}

	if !isValidSearchRecipesRequest(searchRecipesRequest) {
		l.Info("error searching recipes")
		c.Status(http.StatusBadRequest)
		return
	}

	recipes, err := s.db.SearchRecipes(context.Background(), apiSearchR2DBSearchR(searchRecipesRequest))
	if err != nil {
		l.Error("error searching recipes", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	if len(recipes) == 0 {
		c.Status(http.StatusOK)
		return
	}

	var searchRecipesResponse []Recipe

	for _, recipe := range recipes {
		r := dbRecipe2ApiRecipe(&recipe)
		searchRecipesResponse = append(searchRecipesResponse, r)
	}

	c.JSON(http.StatusOK, searchRecipesResponse)
}

func (s *Service) CreateRecipe(c *gin.Context) {
	l := s.l.Named("CreateRecipe")

	var createRecipeRequest Recipe

	if err := json.NewDecoder(c.Request.Body).Decode(&createRecipeRequest); err != nil {
		l.Info("error creating recipe", zap.Error(err))
		c.Status(http.StatusBadRequest)
		return
	}

	if !isValidCreateRecipeRequest(createRecipeRequest) {
		l.Info("error creating recipe")
		c.Status(http.StatusBadRequest)
		return
	}

	recipe, err := s.db.CreateRecipe(context.Background(), apiRecipe2DBRecipe(createRecipeRequest))
	if err != nil {
		l.Error("error creating recipe", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, dbRecipe2ApiRecipe(recipe))
}

func (s *Service) UpdateRecipe(c *gin.Context) {
	l := s.l.Named("UpdateRecipe")

	id := c.Param("id")

	rid, err := uuid.Parse(id)
	if err != nil {
		l.Info("error updating recipe", zap.Error(err))
		c.Status(http.StatusBadRequest)
		return
	}

	var updateRecipeRequest Recipe

	if err := json.NewDecoder(c.Request.Body).Decode(&updateRecipeRequest); err != nil {
		l.Info("error updating recipe", zap.Error(err))
		c.Status(http.StatusBadRequest)
		return
	}

	if !isValidUpdateRecipeRequest(updateRecipeRequest, rid) {
		l.Info("error updating recipe", zap.Error(err))
		c.Status(http.StatusBadRequest)
		return
	}

	recipe, err := s.db.UpdateRecipe(context.Background(), apiRecipe2DBRecipe(updateRecipeRequest))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			l.Info("error updating recipe", zap.Error(err))
			c.Status(http.StatusNotFound)
			return
		}
		l.Error("error updating recipe", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, dbRecipe2ApiRecipe(recipe))
}

func (s *Service) DeleteRecipe(c *gin.Context) {
	l := s.l.Named("DeleteRecipe")

	id := c.Param("id")

	rid, err := uuid.Parse(id)
	if err != nil {
		l.Info("error deleting recipe", zap.Error(err))
		c.Status(http.StatusBadRequest)
		return
	}

	if err = s.db.DeleteRecipe(context.Background(), rid); err != nil {
		l.Error("error deleting recipe", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}
