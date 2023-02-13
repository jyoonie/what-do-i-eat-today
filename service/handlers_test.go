package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"wdiet/store"
	"wdiet/store/mockstore"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var l = zap.NewNop()

var testServer = New(&mockstore.Mockstore{}, l)

func TestPing(t *testing.T) {
	testcases := []struct {
		name                   string
		pingOverrideFunc       func() error
		expectedResponseBody   string
		expectedResponseStatus int
	}{
		{
			"happyPath",
			nil,
			`{"message":"pong"}`,
			http.StatusOK,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/ping", nil)
			w := httptest.NewRecorder()

			testServer.db = &mockstore.Mockstore{PingOverride: testcase.pingOverrideFunc}
			testServer.r.ServeHTTP(w, req)

			assert.Equal(t, testcase.expectedResponseStatus, w.Code)
			assert.Equal(t, testcase.expectedResponseBody, w.Body.String()) //여기선 json message를 담을 struct 구조가 없으니까 걍 json 비교해..
		})
	}
}

func TestLogin(t *testing.T) {
	testcases := []struct {
		name               string
		getUserByEmailFunc func(ctx context.Context, email string) (*store.User, error)
		requestBody        Login
		expectedStatus     int
	}{
		{
			"happyPath",
			nil,
			Login{EmailAddress: "jywoo92324@gmail.com", Password: "hello"},
			http.StatusOK,
		},
		{
			"badRequest",
			nil,
			Login{EmailAddress: "", Password: "hello"},
			http.StatusBadRequest,
		},
		{
			"internalServerError",
			func(ctx context.Context, email string) (*store.User, error) {
				return nil, errors.New("internalServerError")
			},
			Login{EmailAddress: "jywoo92324@gmail.com", Password: "hello"},
			http.StatusInternalServerError,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			reqBody, err := json.Marshal(testcase.requestBody)
			assert.NoError(t, err, "unexpected error marshalling the request body")

			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(reqBody))
			w := httptest.NewRecorder()

			testServer.db = &mockstore.Mockstore{GetUserByEmailOverride: testcase.getUserByEmailFunc}
			testServer.r.ServeHTTP(w, req)

			assert.Equal(t, testcase.expectedStatus, w.Code)

			if testcase.expectedStatus == http.StatusOK {
				var resBody Token

				err = json.Unmarshal(w.Body.Bytes(), &resBody)
				assert.NoError(t, err, "unexpected error unmarshalling the response body")

				tkn, err := jwt.Parse(resBody.Token,
					func(t *jwt.Token) (interface{}, error) {
						if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
							return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
						}
						return testServer.mySigningKey, nil
					},
					jwt.WithValidMethods([]string{testServer.mySigningMethod.Alg()}),
				)
				assert.NoError(t, err, "unexpected error parsing the token")

				assert.Equal(t, true, tkn.Valid)
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	claims := jwt.RegisteredClaims{
		// A usual scenario is to set the expiration time relative to the current time
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Issuer:    "whatDoIEatToday",
		ID:        "080b5f09-527b-4581-bb56-19adbfe50ebf",
		Audience:  []string{"whatDoIEatToday"},
	}

	token := jwt.NewWithClaims(testServer.mySigningMethod, claims)
	signedToken, err := token.SignedString(testServer.mySigningKey)
	assert.NoError(t, err, "unexpected error signing the token")

	testcases := []struct {
		name           string
		authHeader     string
		expecterStatus int
	}{
		{
			"happyPath",
			"Bearer " + signedToken,
			http.StatusOK,
		},
		{
			"invalidToken",
			"Bearer " + "potatoes",
			http.StatusUnauthorized,
		},
		{
			"missingToken",
			"Bearer " + "",
			http.StatusUnauthorized,
		},
		{
			"emptyHeader", //panic: tried to access part of an array that didn't exist, 왜냐하면 bearer 부분이 없기 때문 ㅋㅋ... 근데 "abc"가 아니라 아예 ""로 보내면 핸들러에서 toekn == ""에 걸리기 때문에, 그 뒤에 array access할 것도 없이 return되버리므로 panic이 안 났던것임..
			"abc",
			http.StatusUnauthorized,
		},
	}

	testServer.r.GET("/testValidateToken", testServer.ValidateToken)

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/testValidateToken", nil) //request body가 없어도 밑에 줄에서 header는 set할 수 있음. 밑에 줄이 중요한거. 이 new request는 어떤 request든 걍 명목적인거.
			req.Header.Set("Authorization", testcase.authHeader)
			w := httptest.NewRecorder()

			testServer.r.ServeHTTP(w, req)

			assert.Equal(t, testcase.expecterStatus, w.Code)
		})
	}
}

func TestGetUser(t *testing.T) {
	testcases := []struct {
		name                   string
		getUserOverrideFunc    func(ctx context.Context, id uuid.UUID) (*store.User, error)
		requestBody            string
		expectedResponse       *User
		expectedResponseStatus int
	}{
		{
			"happyPath",
			nil,
			"080b5f09-527b-4581-bb56-19adbfe50ebf",
			&User{UserUUID: uuid.MustParse("080b5f09-527b-4581-bb56-19adbfe50ebf"),
				Active:       true,
				FirstName:    "jy",
				LastName:     "woo",
				EmailAddress: "jywoo92324@gmail.com"},
			http.StatusOK,
		},
		{
			"badRequest",
			nil,
			"badrequesthehehe",
			nil,
			http.StatusBadRequest,
		},
		{
			"internalServerError",
			func(ctx context.Context, id uuid.UUID) (*store.User, error) {
				return nil, errors.New("internalServerError")
			},
			"080b5f09-527b-4581-bb56-19adbfe50ebf",
			nil,
			http.StatusInternalServerError,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/users/"+testcase.requestBody, nil)
			w := httptest.NewRecorder()

			testServer.db = &mockstore.Mockstore{GetUserOverride: testcase.getUserOverrideFunc}
			testServer.r.ServeHTTP(w, req)

			assert.Equal(t, testcase.expectedResponseStatus, w.Code)

			if testcase.expectedResponse != nil { //if you have testcase.expectedResponse, check if it's not nil. Otherwise, you can just check if the testcase.expectedStatus is http.ok like in Login handler test.
				var resBody User

				err := json.Unmarshal(w.Body.Bytes(), &resBody)
				assert.NoError(t, err, "unexpected error unmarshalling the response body")

				assert.Equal(t, *testcase.expectedResponse, resBody)
				// assert.Equal(t, testcase.expectedResponse.UserUUID, resBody.UserUUID)
				// assert.Equal(t, testcase.expectedResponse.Active, resBody.Active)
				// assert.Equal(t, testcase.expectedResponse.FirstName, resBody.FirstName)
				// assert.Equal(t, testcase.expectedResponse.LastName, resBody.LastName)
				// assert.Equal(t, testcase.expectedResponse.EmailAddress, resBody.EmailAddress)
			} else {
				assert.Equal(t, 0, w.Body.Len())
			}
		})
	}
}

func TestCreateUser(t *testing.T) {
	type createUserRequest struct {
		User
		Password string
	}

	goodUser := createUserRequest{
		User: User{
			Active:       true,
			FirstName:    "jy",
			LastName:     "woo",
			EmailAddress: "jywoo92324@gmail.com"},
		Password: "abcdefgh"}

	badUser := createUserRequest{
		User: User{
			Active:       true,
			FirstName:    "",
			LastName:     "",
			EmailAddress: "jywoo92324@gmail.com"},
		Password: "abcdefgh"}

	testcases := []struct {
		name                   string
		createUserOverrideFunc func(ctx context.Context, u store.User) (*store.User, error)
		requestBody            createUserRequest
		expectedResponse       *User
		expectedResponseStatus int
	}{
		{
			"happyPath",
			nil,
			goodUser,
			&goodUser.User,
			http.StatusOK,
		},
		{
			"badRequest",
			nil,
			badUser,
			nil,
			http.StatusBadRequest,
		},
		{
			"internalServerError",
			func(ctx context.Context, u store.User) (*store.User, error) {
				return nil, errors.New("internalServerError")
			},
			goodUser,
			nil,
			http.StatusInternalServerError,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			reqBody, err := json.Marshal(testcase.requestBody)
			assert.NoError(t, err, "unexpected error marshalling the request body")

			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(reqBody))
			w := httptest.NewRecorder()

			testServer.db = &mockstore.Mockstore{CreateUserOverride: testcase.createUserOverrideFunc}
			testServer.r.ServeHTTP(w, req)

			assert.Equal(t, testcase.expectedResponseStatus, w.Code)

			if testcase.expectedResponse != nil {
				var resBody User

				err := json.Unmarshal(w.Body.Bytes(), &resBody)
				assert.NoError(t, err, "unexpected error unmarshalling the response body")

				assert.NotEqual(t, resBody.UserUUID, uuid.Nil) //uuid가 랜덤으로 generate 됐기 때문에, 그냥 response uuid가 uuid Nil이 아닌지만 검사.
				//assert.Equal(t, *testcase.expectedResponse, resBody)
				assert.Equal(t, testcase.expectedResponse.Active, resBody.Active)
				assert.Equal(t, testcase.expectedResponse.FirstName, resBody.FirstName)
				assert.Equal(t, testcase.expectedResponse.LastName, resBody.LastName)
				assert.Equal(t, testcase.expectedResponse.EmailAddress, resBody.EmailAddress)
			} else {
				assert.Equal(t, 0, w.Body.Len())
			}
		})
	}
}

func TestUpdateUser(t *testing.T) {
	goodUser := User{ //존 왈, requestBody랑 expectedResponse가 완전히 똑같으므로, 그냥 여기에 하나를 정의해서 복붙해라 ㅋㅋ
		UserUUID:     uuid.MustParse("080b5f09-527b-4581-bb56-19adbfe50ebf"),
		Active:       true,
		FirstName:    "jy",
		LastName:     "woo",
		EmailAddress: "jywoo92324@gmail.com",
	}

	badUser := User{
		UserUUID:     uuid.MustParse("080b5f09-527b-4581-bb56-19adbfe50ebf"),
		Active:       true,
		FirstName:    "",
		LastName:     "",
		EmailAddress: "jywoo92324@gmail.com",
	}

	testcases := []struct {
		name                   string
		updateUserOverrideFunc func(ctx context.Context, u store.User) (*store.User, error)
		requestBody            User
		expectedResponse       *User
		expectedResponseStatus int
	}{
		{
			"happyPath",
			nil,
			goodUser,
			&goodUser,
			http.StatusOK,
		},
		{
			"badRequest",
			nil,
			badUser,
			nil,
			http.StatusBadRequest,
		},
		{
			"internalServerError",
			func(ctx context.Context, u store.User) (*store.User, error) {
				return nil, errors.New("internalServerError")
			},
			goodUser,
			nil,
			http.StatusInternalServerError,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			reqBody, err := json.Marshal(testcase.requestBody)
			assert.NoError(t, err, "unexpected error marshalling request body")

			req := httptest.NewRequest(http.MethodPost, "/users/"+testcase.requestBody.UserUUID.String(), bytes.NewBuffer(reqBody)) //이 경우는 특이. url에 param이랑, request body가 같이 오므로 "/users/" 뒤에 +도 해줘야돼고, bytes.NewBuffer도 해줘야 됨!
			w := httptest.NewRecorder()

			testServer.db = &mockstore.Mockstore{UpdateUserOverride: testcase.updateUserOverrideFunc}
			testServer.r.ServeHTTP(w, req)

			assert.Equal(t, testcase.expectedResponseStatus, w.Code)

			if testcase.expectedResponse != nil {
				var resBody User

				err = json.Unmarshal(w.Body.Bytes(), &resBody)
				assert.NoError(t, err, "unexpected error unmarshalling the response body")

				assert.Equal(t, *testcase.expectedResponse, resBody)
				// assert.Equal(t, testcase.expectedResponse.UserUUID, resBody.UserUUID)
				// assert.Equal(t, testcase.expectedResponse.Active, resBody.Active)
				// assert.Equal(t, testcase.expectedResponse.FirstName, resBody.FirstName)
				// assert.Equal(t, testcase.expectedResponse.LastName, resBody.LastName)
				// assert.Equal(t, testcase.expectedResponse.EmailAddress, resBody.EmailAddress)
			} else {
				assert.Equal(t, 0, w.Body.Len())
			}
		})
	}
}

func TestGetIngredient(t *testing.T) {
	testcases := []struct {
		name                      string
		getIngredientOverrideFunc func(ctx context.Context, id uuid.UUID) (*store.Ingredient, error)
		requestBody               string
		expectedResponse          *Ingredient
		expectedStatus            int
	}{
		{
			"happyPath",
			nil,
			"080b5f09-527b-4581-bb56-19adbfe50ebf",
			&Ingredient{
				IngredientUUID: uuid.MustParse("080b5f09-527b-4581-bb56-19adbfe50ebf"),
				IngredientName: "onion",
				Category:       "vegetables",
				DaysUntilExp:   7},
			http.StatusOK,
		},
		{
			"badRequest",
			nil,
			"maerong",
			nil,
			http.StatusBadRequest,
		},
		{
			"internalServerError",
			func(ctx context.Context, id uuid.UUID) (*store.Ingredient, error) {
				return nil, errors.New("internalServerError")
			},
			"080b5f09-527b-4581-bb56-19adbfe50ebf",
			nil,
			http.StatusInternalServerError,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/ingredients/"+testcase.requestBody, nil)
			w := httptest.NewRecorder()

			testServer.db = &mockstore.Mockstore{GetIngredientOverride: testcase.getIngredientOverrideFunc}
			testServer.r.ServeHTTP(w, req)

			assert.Equal(t, testcase.expectedStatus, w.Code)

			if testcase.expectedResponse != nil {
				var resBody Ingredient

				err := json.Unmarshal(w.Body.Bytes(), &resBody)
				assert.NoError(t, err, "unexpected error unmarshalling the response body")

				assert.Equal(t, *testcase.expectedResponse, resBody)
				// assert.Equal(t, testcase.expectedResponse.IngredientUUID, resBody.IngredientUUID)
				// assert.Equal(t, testcase.expectedResponse.IngredientName, resBody.IngredientName)
				// assert.Equal(t, testcase.expectedResponse.Category, resBody.Category)
				// assert.Equal(t, testcase.expectedResponse.DaysUntilExp, resBody.DaysUntilExp)
			} else {
				assert.Equal(t, 0, w.Body.Len())
			}
		})
	}
}

func TestSearchIngredients(t *testing.T) {
	testcases := []struct {
		name                          string
		searchIngredientsOverrideFunc func(ctx context.Context, i store.SearchIngredient) ([]store.Ingredient, error)
		requestBody                   Ingredient
		expectedResponse              []Ingredient
		expectedStatus                int
	}{
		{
			"happyPath",
			nil,
			Ingredient{IngredientName: "tuna"},
			[]Ingredient{
				{
					IngredientUUID: uuid.MustParse("080b5f09-527b-4581-bb56-19adbfe50ebf"),
					IngredientName: "tuna",
					Category:       "tuna kimbap",
					DaysUntilExp:   3,
				},
				{
					IngredientUUID: uuid.MustParse("080b5f09-527b-4581-bb56-19adbfe50ebf"),
					IngredientName: "tuna",
					Category:       "tuna sushi",
					DaysUntilExp:   3,
				},
			},
			http.StatusOK,
		},
		{
			"badRequest",
			nil,
			Ingredient{},
			nil,
			http.StatusBadRequest,
		},
		{
			"internalServerError",
			func(ctx context.Context, i store.SearchIngredient) ([]store.Ingredient, error) {
				return []store.Ingredient{}, errors.New("internalServerError")
			},
			Ingredient{IngredientName: "tuna"},
			nil,
			http.StatusInternalServerError,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			reqBody, err := json.Marshal(testcase.requestBody)
			assert.NoError(t, err, "unexpected error marshalling the request body")

			req := httptest.NewRequest(http.MethodPost, "/ingredients/search", bytes.NewBuffer(reqBody))
			w := httptest.NewRecorder()

			testServer.db = &mockstore.Mockstore{SearchIngredientsOverride: testcase.searchIngredientsOverrideFunc}
			testServer.r.ServeHTTP(w, req)

			assert.Equal(t, testcase.expectedStatus, w.Code)

			if testcase.expectedResponse != nil {
				var resBody []Ingredient

				err := json.Unmarshal(w.Body.Bytes(), &resBody)
				assert.NoError(t, err, "unexpected error unmarshaling the response body")

				assert.Equal(t, len(testcase.expectedResponse), len(resBody))

				for i := range testcase.expectedResponse {
					assert.Equal(t, testcase.expectedResponse[i], resBody[i])
					// assert.Equal(t, testcase.expectedResponse[i].IngredientUUID, resBody[i].IngredientUUID)
					// assert.Equal(t, testcase.expectedResponse[i].IngredientName, resBody[i].IngredientName)
					// assert.Equal(t, testcase.expectedResponse[i].Category, resBody[i].Category)
					// assert.Equal(t, testcase.expectedResponse[i].DaysUntilExp, resBody[i].DaysUntilExp)
				}
			} else {
				assert.Equal(t, 0, w.Body.Len())
			}
		})
	}
}

func TestCreateIngredient(t *testing.T) {
	goodIngredient := Ingredient{
		IngredientName: "onion",
		Category:       "vegetables",
		DaysUntilExp:   7,
	}

	badIngredient := Ingredient{
		IngredientName: "",
		Category:       "",
		DaysUntilExp:   7,
	}

	testcases := []struct {
		name                         string
		createIngredientOverrideFunc func(ctx context.Context, i store.Ingredient) (*store.Ingredient, error)
		requestBody                  Ingredient
		expectedResponse             *Ingredient
		expectedStatus               int
	}{
		{
			"happyPath",
			nil,
			goodIngredient,
			&goodIngredient,
			http.StatusOK,
		},
		{
			"badRequest",
			nil,
			badIngredient,
			nil,
			http.StatusBadRequest,
		},
		{
			"internalServerError",
			func(ctx context.Context, i store.Ingredient) (*store.Ingredient, error) {
				return nil, errors.New("internalServerError")
			},
			goodIngredient,
			nil,
			http.StatusInternalServerError,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			reqBody, err := json.Marshal(testcase.requestBody)
			assert.NoError(t, err, "unexpected error marshalling the request body")

			req := httptest.NewRequest(http.MethodPost, "/ingredients", bytes.NewBuffer(reqBody))
			w := httptest.NewRecorder()

			testServer.db = &mockstore.Mockstore{CreateIngredientOverride: testcase.createIngredientOverrideFunc}
			testServer.r.ServeHTTP(w, req)

			if testcase.expectedResponse != nil {
				var resBody Ingredient

				err := json.Unmarshal(w.Body.Bytes(), &resBody)
				assert.NoError(t, err, "unexpected error unmarshalling the response body")

				//assert.Equal(t, *testcase.expectedResponse, resBody)
				assert.NotEqual(t, resBody.IngredientUUID, uuid.Nil)
				assert.Equal(t, testcase.expectedResponse.IngredientName, resBody.IngredientName)
				assert.Equal(t, testcase.expectedResponse.Category, resBody.Category)
				assert.Equal(t, testcase.expectedResponse.DaysUntilExp, resBody.DaysUntilExp)
			} else {
				assert.Equal(t, 0, w.Body.Len())
			}
		})
	}
}

func TestUpdateIngredient(t *testing.T) {
	goodIngredient := Ingredient{
		IngredientUUID: uuid.MustParse("080b5f09-527b-4581-bb56-19adbfe50ebf"),
		IngredientName: "onion",
		Category:       "vegetables",
		DaysUntilExp:   7,
	}

	badIngredient := Ingredient{
		IngredientUUID: uuid.MustParse("080b5f09-527b-4581-bb56-19adbfe50ebf"),
		IngredientName: "",
		Category:       "",
		DaysUntilExp:   7,
	}

	testcases := []struct {
		name                         string
		updateIngredientOverrideFunc func(ctx context.Context, i store.Ingredient) (*store.Ingredient, error)
		requestBody                  Ingredient
		expectedResponse             *Ingredient
		expectedStatus               int
	}{
		{
			"happyPath",
			nil,
			goodIngredient,
			&goodIngredient,
			http.StatusOK,
		},
		{
			"badRequest",
			//it's misleading to have database override here, because it's not gonna hit the database anyway. To more clearly see it's a bad request, just leave it nil.
			nil,
			badIngredient,
			nil,
			http.StatusBadRequest,
		},
		{
			"internalServerError",
			func(ctx context.Context, i store.Ingredient) (*store.Ingredient, error) {
				return nil, errors.New("internalServerError")
			},
			goodIngredient,
			nil,
			http.StatusInternalServerError,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			reqBody, err := json.Marshal(testcase.requestBody)
			assert.NoError(t, err, "unexpected error marshalling the request body")

			req := httptest.NewRequest(http.MethodPost, "/ingredients/"+testcase.requestBody.IngredientUUID.String(), bytes.NewBuffer(reqBody))
			w := httptest.NewRecorder()

			testServer.db = &mockstore.Mockstore{UpdateIngredientOverride: testcase.updateIngredientOverrideFunc}
			testServer.r.ServeHTTP(w, req)

			assert.Equal(t, testcase.expectedStatus, w.Code)

			if testcase.expectedResponse != nil {
				var resBody Ingredient

				err := json.Unmarshal(w.Body.Bytes(), &resBody)
				assert.NoError(t, err, "unexpected error unmarshalling the response body")

				assert.Equal(t, *testcase.expectedResponse, resBody)
				// assert.Equal(t, testcase.expectedResponse.IngredientName, resBody.IngredientName)
				// assert.Equal(t, testcase.expectedResponse.Category, resBody.Category)
				// assert.Equal(t, testcase.expectedResponse.DaysUntilExp, resBody.DaysUntilExp)
				// assert.Equal(t, testcase.expectedResponse.IngredientUUID, resBody.IngredientUUID)
			} else {
				assert.Equal(t, 0, w.Body.Len()) //이거 w.Body.Len이 아니라 끝에 꼭 () 붙여줘 ㅎㅎ;;
			}
		})
	}
}

func TestDeleteIngredient(t *testing.T) {
	testcases := []struct {
		name                         string
		deleteIngredientOverrideFunc func(ctx context.Context, id uuid.UUID) error
		requestBody                  string
		expectedStatus               int
	}{
		{
			"happyPath",
			nil,
			"080b5f09-527b-4581-bb56-19adbfe50ebf",
			http.StatusOK,
		},
		{
			"badRequest",
			nil,
			"maerongmaerong",
			http.StatusBadRequest,
		},
		{
			"happyPath",
			func(ctx context.Context, id uuid.UUID) error {
				return errors.New("internalServerError")
			},
			"080b5f09-527b-4581-bb56-19adbfe50ebf",
			http.StatusInternalServerError,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/ingredients/"+testcase.requestBody, nil)
			w := httptest.NewRecorder()

			testServer.db = &mockstore.Mockstore{DeleteIngredientOverride: testcase.deleteIngredientOverrideFunc}
			testServer.r.ServeHTTP(w, req)

			assert.Equal(t, testcase.expectedStatus, w.Code)
		})
	}
}

func TestListFridgeIngredients(t *testing.T) {
	testcases := []struct {
		name                              string
		listFridgeIngredientsOverrideFunc func(ctx context.Context, id uuid.UUID) ([]store.FridgeIngredient, error)
		requestBody                       string
		expectedResponse                  []FridgeIngredient
		expectedStatus                    int
	}{
		{
			"happyPath",
			nil,
			"080b5f09-527b-4581-bb56-19adbfe50ebf",
			[]FridgeIngredient{
				{
					UserUUID:       uuid.MustParse("080b5f09-527b-4581-bb56-19adbfe50ebf"),
					IngredientUUID: uuid.MustParse("ffff7c73-52b0-4e3d-bf3f-0c26785ef972"),
					Amount:         3,
					Unit:           "kg",
					PurchasedDate:  time.Date(2023, time.March, 24, 15, 0, 0, 0, time.Now().Location()),
					ExpirationDate: time.Date(2023, time.March, 24, 15, 0, 0, 0, time.Now().Location()),
				},
				{
					UserUUID:       uuid.MustParse("080b5f09-527b-4581-bb56-19adbfe50ebf"),
					IngredientUUID: uuid.MustParse("2c98fff4-7ccc-4536-8259-67a88380e99c"),
					Amount:         2,
					Unit:           "L",
					PurchasedDate:  time.Date(2023, time.March, 24, 15, 0, 0, 0, time.Now().Location()),
					ExpirationDate: time.Date(2023, time.March, 31, 15, 0, 0, 0, time.Now().Location()),
				},
			},
			http.StatusOK,
		},
		{
			"badRequest",
			nil,
			"maerong",
			nil,
			http.StatusBadRequest,
		},
		{
			"internalServerError",
			func(ctx context.Context, id uuid.UUID) ([]store.FridgeIngredient, error) {
				return nil, errors.New("internalServerError")
			},
			"080b5f09-527b-4581-bb56-19adbfe50ebf",
			nil,
			http.StatusInternalServerError,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/users/"+testcase.requestBody+"/fridge_ingredients", nil)
			w := httptest.NewRecorder()

			testServer.db = &mockstore.Mockstore{ListFridgeIngredientsOverride: testcase.listFridgeIngredientsOverrideFunc}
			testServer.r.ServeHTTP(w, req)

			assert.Equal(t, testcase.expectedStatus, w.Code)

			if testcase.expectedResponse != nil {
				var resBody []FridgeIngredient

				err := json.Unmarshal(w.Body.Bytes(), &resBody)
				assert.NoError(t, err, "unexpected error unmarshalling the response body")

				assert.Equal(t, len(testcase.expectedResponse), len(resBody))

				for i := range testcase.expectedResponse {
					assert.Equal(t, testcase.expectedResponse[i], resBody[i])
				}
			} else {
				assert.Equal(t, 0, w.Body.Len())
			}
		})
	}
}

func TestCreateFridgeIngredient(t *testing.T) {
	goodFridgeIngredient := FridgeIngredient{
		UserUUID:       uuid.MustParse("080b5f09-527b-4581-bb56-19adbfe50ebf"),
		IngredientUUID: uuid.MustParse("ffff7c73-52b0-4e3d-bf3f-0c26785ef972"),
		Amount:         3,
		Unit:           "kg",
		PurchasedDate:  time.Now(),
	}

	badFridgeIngredient := FridgeIngredient{
		UserUUID:       uuid.Nil,
		IngredientUUID: uuid.MustParse("2c98fff4-7ccc-4536-8259-67a88380e99c"),
		Amount:         3,
		Unit:           "kg",
		PurchasedDate:  time.Now(),
	}

	testcases := []struct {
		name                               string
		getIngredientOverrideFunc          func(ctx context.Context, id uuid.UUID) (*store.Ingredient, error)
		createFridgeIngredientOverrideFunc func(ctx context.Context, f store.FridgeIngredient) (*store.FridgeIngredient, error)
		requestBody                        FridgeIngredient
		expectedResponse                   *FridgeIngredient
		expectedStatus                     int
	}{
		{
			"happyPath",
			nil,
			nil,
			goodFridgeIngredient,
			&FridgeIngredient{
				UserUUID:       uuid.MustParse("080b5f09-527b-4581-bb56-19adbfe50ebf"),
				IngredientUUID: uuid.MustParse("ffff7c73-52b0-4e3d-bf3f-0c26785ef972"),
				Amount:         3,
				Unit:           "kg",
				PurchasedDate:  time.Now(),
				ExpirationDate: time.Now().Add(7 * 24 * time.Hour), //여기서의 time.Now()는 직전의 time.Now()랑 아주 약간 달라질 것이므로 expectedResponse 검사할 때 정확한 time.Time을 검사하는게 아니라 year, month, day만 검사하래..
			},
			http.StatusOK,
		},
		{
			"badRequest",
			nil,
			nil,
			badFridgeIngredient,
			nil,
			http.StatusBadRequest,
		},
		{
			"internalServerError:getIngredient",
			func(ctx context.Context, id uuid.UUID) (*store.Ingredient, error) {
				return nil, errors.New("internalServerError:getIngredient")
			},
			nil,
			goodFridgeIngredient,
			nil,
			http.StatusInternalServerError,
		},
		{
			"internalServerError:createFridgeIngredient",
			nil,
			func(ctx context.Context, f store.FridgeIngredient) (*store.FridgeIngredient, error) {
				return nil, errors.New("internalServerError:createFridgeIngredient")
			},
			goodFridgeIngredient,
			nil,
			http.StatusInternalServerError,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			reqBody, err := json.Marshal(testcase.requestBody)
			assert.NoError(t, err, "unexpected error marshalling the request body")

			req := httptest.NewRequest(http.MethodPost, "/fridge_ingredients", bytes.NewBuffer(reqBody))
			w := httptest.NewRecorder()

			testServer.db = &mockstore.Mockstore{
				GetIngredientOverride:          testcase.getIngredientOverrideFunc,
				CreateFridgeIngredientOverride: testcase.createFridgeIngredientOverrideFunc}
			testServer.r.ServeHTTP(w, req)

			assert.Equal(t, testcase.expectedStatus, w.Code)

			if testcase.expectedResponse != nil {
				var resBody FridgeIngredient

				err := json.Unmarshal(w.Body.Bytes(), &resBody)
				assert.NoError(t, err, "unexpected error unmarshalling the response body")

				assert.Equal(t, testcase.expectedResponse.PurchasedDate.Year(), resBody.PurchasedDate.Year())
				assert.Equal(t, testcase.expectedResponse.PurchasedDate.Month(), resBody.PurchasedDate.Month())
				assert.Equal(t, testcase.expectedResponse.PurchasedDate.Day(), resBody.PurchasedDate.Day()) //이거 한줄로 줄이는 방법 없을까..

				assert.Equal(t, testcase.expectedResponse.ExpirationDate.Year(), resBody.ExpirationDate.Year())
				assert.Equal(t, testcase.expectedResponse.ExpirationDate.Month(), resBody.ExpirationDate.Month())
				assert.Equal(t, testcase.expectedResponse.ExpirationDate.Day(), resBody.ExpirationDate.Day())

				assert.Equal(t, testcase.expectedResponse.UserUUID, resBody.UserUUID)
				assert.Equal(t, testcase.expectedResponse.IngredientUUID, resBody.IngredientUUID)
				assert.Equal(t, testcase.expectedResponse.Amount, resBody.Amount)
				assert.Equal(t, testcase.expectedResponse.Amount, resBody.Amount)
				assert.Equal(t, testcase.expectedResponse.Unit, resBody.Unit)
			} else {
				assert.Equal(t, 0, w.Body.Len())
			}
		})
	}
}

func TestUpdateFridgeIngredient(t *testing.T) {
	goodFridgeIngredient := FridgeIngredient{
		UserUUID:       uuid.MustParse("080b5f09-527b-4581-bb56-19adbfe50ebf"),
		IngredientUUID: uuid.MustParse("ffff7c73-52b0-4e3d-bf3f-0c26785ef972"),
		Amount:         5,
		Unit:           "lb",
		PurchasedDate:  time.Now(),
	}

	badFridgeIngredient := FridgeIngredient{
		UserUUID:       uuid.Nil,
		IngredientUUID: uuid.MustParse("2c98fff4-7ccc-4536-8259-67a88380e99c"),
		Amount:         5,
		Unit:           "kg",
		PurchasedDate:  time.Now(),
	}

	testcases := []struct {
		name                               string
		updateFridgeIngredientOverrideFunc func(ctx context.Context, f store.FridgeIngredient) (*store.FridgeIngredient, error)
		requestBody                        FridgeIngredient
		expectedResponse                   *FridgeIngredient
		expectedStatus                     int
	}{
		{
			"happyPath",
			nil,
			goodFridgeIngredient,
			&goodFridgeIngredient,
			http.StatusOK,
		},
		{
			"badRequest",
			nil,
			badFridgeIngredient,
			nil,
			http.StatusBadRequest,
		},
		{
			"internalServerError",
			func(ctx context.Context, f store.FridgeIngredient) (*store.FridgeIngredient, error) {
				return nil, errors.New("internalServerError")
			},
			goodFridgeIngredient,
			nil,
			http.StatusInternalServerError,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			reqBody, err := json.Marshal(testcase.requestBody)
			assert.NoError(t, err, "unexpected error marshalling the request body")

			req := httptest.NewRequest(http.MethodPost, "/fridge_ingredients/"+testcase.requestBody.UserUUID.String(), bytes.NewBuffer(reqBody))
			w := httptest.NewRecorder()

			testServer.db = &mockstore.Mockstore{UpdateFridgeIngredientOverride: testcase.updateFridgeIngredientOverrideFunc}
			testServer.r.ServeHTTP(w, req)

			assert.Equal(t, testcase.expectedStatus, w.Code)

			if testcase.expectedResponse != nil {
				var resBody FridgeIngredient

				err := json.Unmarshal(w.Body.Bytes(), &resBody)
				assert.NoError(t, err, "unexpected error unmarshalling the response body")

				assert.Equal(t, testcase.expectedResponse.PurchasedDate.Year(), resBody.PurchasedDate.Year())
				assert.Equal(t, testcase.expectedResponse.PurchasedDate.Month(), resBody.PurchasedDate.Month())
				assert.Equal(t, testcase.expectedResponse.PurchasedDate.Day(), resBody.PurchasedDate.Day()) //이거 한줄로 줄이는 방법 없을까..

				assert.Equal(t, testcase.expectedResponse.ExpirationDate.Year(), resBody.ExpirationDate.Year())
				assert.Equal(t, testcase.expectedResponse.ExpirationDate.Month(), resBody.ExpirationDate.Month())
				assert.Equal(t, testcase.expectedResponse.ExpirationDate.Day(), resBody.ExpirationDate.Day())

				assert.Equal(t, testcase.expectedResponse.UserUUID, resBody.UserUUID)
				assert.Equal(t, testcase.expectedResponse.IngredientUUID, resBody.IngredientUUID)
				assert.Equal(t, testcase.expectedResponse.Amount, resBody.Amount)
				assert.Equal(t, testcase.expectedResponse.Amount, resBody.Amount)
				assert.Equal(t, testcase.expectedResponse.Unit, resBody.Unit)
			} else {
				assert.Equal(t, 0, w.Body.Len())
			}
		})
	}
}

func TestDeleteFridgeIngredient(t *testing.T) {
	testcases := []struct {
		name                               string
		deleteFridgeIngredientOverrideFunc func(ctx context.Context, uid uuid.UUID, fid uuid.UUID) error
		uid                                string
		fid                                string
		expectedStatus                     int
	}{
		{
			"happyPath",
			nil,
			"080b5f09-527b-4581-bb56-19adbfe50ebf",
			"ffff7c73-52b0-4e3d-bf3f-0c26785ef972",
			http.StatusOK,
		},
		{
			"badRequest",
			nil,
			"",
			"",
			http.StatusBadRequest,
		},
		{
			"internalServerError",
			func(ctx context.Context, uid uuid.UUID, fid uuid.UUID) error {
				return errors.New("internalServerError")
			},
			"080b5f09-527b-4581-bb56-19adbfe50ebf",
			"ffff7c73-52b0-4e3d-bf3f-0c26785ef972",
			http.StatusInternalServerError,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/users/"+testcase.uid+"/fridge_ingredients/"+testcase.fid, nil)
			w := httptest.NewRecorder()

			testServer.db = &mockstore.Mockstore{DeleteFridgeIngredientOverride: testcase.deleteFridgeIngredientOverrideFunc}
			testServer.r.ServeHTTP(w, req)

			assert.Equal(t, testcase.expectedStatus, w.Code)
		})
	}
}

func TestGetRecipe(t *testing.T) {
	testcases := []struct {
		name                  string
		getRecipeOverrideFunc func(ctx context.Context, id uuid.UUID) (*store.Recipe, error)
		requestBody           string
		expectedResponse      *Recipe
		expectedStatus        int
	}{
		{
			"happyPath",
			nil,
			"080b5f09-527b-4581-bb56-19adbfe50ebf",
			&Recipe{
				RecipeUUID: uuid.MustParse("080b5f09-527b-4581-bb56-19adbfe50ebf"),
				UserUUID:   uuid.MustParse("2c98fff4-7ccc-4536-8259-67a88380e99c"),
				RecipeName: "kimchi fried rice",
				Category:   "Korean",
			},
			http.StatusOK,
		},
		{
			"badRequest",
			nil,
			"maerong",
			nil,
			http.StatusBadRequest,
		},
		{
			"internalServerError",
			func(ctx context.Context, id uuid.UUID) (*store.Recipe, error) {
				return nil, errors.New("internalServerError")
			},
			"080b5f09-527b-4581-bb56-19adbfe50ebf",
			nil,
			http.StatusInternalServerError,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/recipes/"+testcase.requestBody, nil)
			w := httptest.NewRecorder()

			testServer.db = &mockstore.Mockstore{GetRecipeOverride: testcase.getRecipeOverrideFunc}
			testServer.r.ServeHTTP(w, req)

			assert.Equal(t, testcase.expectedStatus, w.Code)

			if testcase.expectedResponse != nil {
				var resBody Recipe

				err := json.Unmarshal(w.Body.Bytes(), &resBody)
				assert.NoError(t, err, "unexpected error unmarshalling the response body")

				assert.Equal(t, *testcase.expectedResponse, resBody)
			} else {
				assert.Equal(t, 0, w.Body.Len())
			}
		})
	}
}

func TestListRecipes(t *testing.T) {
	testcases := []struct {
		name                    string
		listRecipesOverrideFunc func(ctx context.Context, id uuid.UUID) ([]store.Recipe, error)
		requestBody             string
		expectedResponse        []Recipe
		expectedStatus          int
	}{
		{
			"happyPath",
			nil,
			"080b5f09-527b-4581-bb56-19adbfe50ebf",
			[]Recipe{
				{
					RecipeUUID: uuid.MustParse("ffff7c73-52b0-4e3d-bf3f-0c26785ef972"),
					UserUUID:   uuid.MustParse("080b5f09-527b-4581-bb56-19adbfe50ebf"),
					RecipeName: "kimchi jeon",
					Category:   "Korean",
				},
				{
					RecipeUUID: uuid.MustParse("2c98fff4-7ccc-4536-8259-67a88380e99c"),
					UserUUID:   uuid.MustParse("080b5f09-527b-4581-bb56-19adbfe50ebf"),
					RecipeName: "kimchi mandu",
					Category:   "Korean",
				},
			},
			http.StatusOK,
		},
		{
			"badRequest",
			nil,
			"babo",
			nil,
			http.StatusBadRequest,
		},
		{
			"internalServerError",
			func(ctx context.Context, id uuid.UUID) ([]store.Recipe, error) {
				return nil, errors.New("internalServerError")
			},
			"080b5f09-527b-4581-bb56-19adbfe50ebf",
			nil,
			http.StatusInternalServerError,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/users/"+testcase.requestBody+"/recipes", nil)
			w := httptest.NewRecorder()

			testServer.db = &mockstore.Mockstore{ListRecipesOverride: testcase.listRecipesOverrideFunc}
			testServer.r.ServeHTTP(w, req)

			assert.Equal(t, testcase.expectedStatus, w.Code)

			if testcase.expectedResponse != nil {
				var resBody []Recipe

				err := json.Unmarshal(w.Body.Bytes(), &resBody)
				assert.NoError(t, err, "unexpected error unmarshalling the response body")

				assert.Equal(t, len(testcase.expectedResponse), len(resBody))

				for i := range testcase.expectedResponse {
					assert.Equal(t, testcase.expectedResponse[i], resBody[i])
				}
			} else {
				assert.Equal(t, 0, w.Body.Len())
			}
		})
	}
}

func TestSearchRecipes(t *testing.T) {
	goodRecipe := SearchRecipes{
		RecipeName: "salmon nigiri",
	}

	badRecipe := SearchRecipes{}

	testcases := []struct {
		name                      string
		searchRecipesOverrideFunc func(ctx context.Context, r store.SearchRecipes) ([]store.Recipe, error)
		requestBody               SearchRecipes
		expectedResponse          []Recipe
		expectedStatus            int
	}{
		{
			"happyPath",
			nil,
			goodRecipe,
			[]Recipe{
				{
					RecipeUUID: uuid.MustParse("ffff7c73-52b0-4e3d-bf3f-0c26785ef972"),
					UserUUID:   uuid.MustParse("080b5f09-527b-4581-bb56-19adbfe50ebf"),
					RecipeName: "salmon nigiri",
					Category:   "Japanese",
				},
				{
					RecipeUUID: uuid.MustParse("2c98fff4-7ccc-4536-8259-67a88380e99c"),
					UserUUID:   uuid.MustParse("ebe96725-44ef-47ee-979f-8baf823d7283"),
					RecipeName: "salmon nigiri",
					Category:   "Japanese",
				},
			},
			http.StatusOK,
		},
		{
			"badRequest",
			nil,
			badRecipe,
			nil,
			http.StatusBadRequest,
		},
		{
			"internalServerError",
			func(ctx context.Context, r store.SearchRecipes) ([]store.Recipe, error) {
				return nil, errors.New("internalServerError")
			},
			goodRecipe,
			nil,
			http.StatusInternalServerError,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			reqBody, err := json.Marshal(testcase.requestBody)
			assert.NoError(t, err, "unexpected error marshalling the request body")

			req := httptest.NewRequest(http.MethodPost, "/recipes/search", bytes.NewBuffer(reqBody))
			w := httptest.NewRecorder()

			testServer.db = &mockstore.Mockstore{SearchRecipesOverride: testcase.searchRecipesOverrideFunc}
			testServer.r.ServeHTTP(w, req)

			assert.Equal(t, testcase.expectedStatus, w.Code)

			if testcase.expectedResponse != nil {
				var resBody []Recipe

				err := json.Unmarshal(w.Body.Bytes(), &resBody)
				assert.NoError(t, err, "unexpected error unmarshalling the response body")

				assert.Equal(t, len(testcase.expectedResponse), len(resBody))

				for i := range testcase.expectedResponse {
					assert.Equal(t, testcase.expectedResponse[i], resBody[i])
				}
			} else {
				assert.Equal(t, 0, w.Body.Len())
			}
		})
	}
}

func TestCreateRecipe(t *testing.T) {
	goodRecipe := Recipe{
		UserUUID:   uuid.MustParse("2c98fff4-7ccc-4536-8259-67a88380e99c"),
		RecipeName: "kimchi fried rice",
		Category:   "Korean",
		Ingredients: []RecipeIngredient{
			{
				IngredientUUID: uuid.MustParse("2c98fff4-7ccc-4536-8259-67a88380e99b"),
				Amount:         1,
				Unit:           "kg",
			},
			{
				IngredientUUID: uuid.MustParse("2c98fff4-7ccc-4536-8259-67a88380e99b"),
				Amount:         1,
				Unit:           "unit",
			},
		},
		Instructions: []RecipeInstruction{
			{
				StepNum:     1,
				Instruction: "Chop kimchi, onion and pork belly",
			},
			{
				StepNum:     2,
				Instruction: "grill the pan and put some oil on it",
			},
		},
	}

	badRecipe := Recipe{
		UserUUID:     uuid.MustParse("2c98fff4-7ccc-4536-8259-67a88380e99c"),
		RecipeName:   "kimchi fried rice",
		Category:     "",
		Ingredients:  []RecipeIngredient{},
		Instructions: []RecipeInstruction{},
	}

	testcases := []struct {
		name                     string
		createRecipeOverrideFunc func(ctx context.Context, r store.Recipe) (*store.Recipe, error)
		requestBody              Recipe
		expectedReponse          *Recipe
		expectedStatus           int
	}{
		{
			"happyPath",
			nil,
			goodRecipe,
			&goodRecipe, //RecipeUUID 필드가 핸들러에서 돌아오겠지만(w.Body()), 내가 여기서 expected response로 체크하고 싶은건 내맘대로 쓸 수 있다. RecipeUUID는 밑에서 그냥 uuid.Nil이 아닌지만 체크하고, 나머지 필드들만 값이 내가 예상한거랑 같은지 보는것!
			http.StatusOK,
		},
		{
			"badRequest",
			nil,
			badRecipe,
			nil,
			http.StatusBadRequest,
		},
		{
			"internalServerError",
			func(ctx context.Context, r store.Recipe) (*store.Recipe, error) {
				return nil, errors.New("internalServerError")
			},
			goodRecipe,
			nil,
			http.StatusInternalServerError,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			reqBody, err := json.Marshal(testcase.requestBody)
			assert.NoError(t, err, "unexpected error marshalling the request body")

			req := httptest.NewRequest(http.MethodPost, "/recipes", bytes.NewBuffer(reqBody))
			w := httptest.NewRecorder()

			testServer.db = &mockstore.Mockstore{CreateRecipeOverride: testcase.createRecipeOverrideFunc}
			testServer.r.ServeHTTP(w, req)

			assert.Equal(t, testcase.expectedStatus, w.Code)

			if testcase.expectedReponse != nil {
				var resBody Recipe

				err := json.Unmarshal(w.Body.Bytes(), &resBody)
				assert.NoError(t, err, "unexpected error unmarshalling the response body")

				assert.NotEqual(t, resBody.RecipeUUID, uuid.Nil)
				assert.Equal(t, testcase.expectedReponse.UserUUID, resBody.UserUUID)
				assert.Equal(t, testcase.expectedReponse.RecipeName, resBody.RecipeName)
				assert.Equal(t, testcase.expectedReponse.Category, resBody.Category)
				assert.Equal(t, testcase.expectedReponse.Ingredients, resBody.Ingredients)
				assert.Equal(t, testcase.expectedReponse.Instructions, resBody.Instructions)
			} else {
				assert.Equal(t, 0, w.Body.Len())
			}
		})
	}
}

func TestUpdateRecipe(t *testing.T) {
	goodRecipe := Recipe{
		RecipeUUID: uuid.MustParse("2c98fff4-7ccc-4536-8259-67a88380e99d"),
		UserUUID:   uuid.MustParse("2c98fff4-7ccc-4536-8259-67a88380e99c"),
		RecipeName: "kimchi fried rice",
		Category:   "Korean",
		Ingredients: []RecipeIngredient{
			{
				IngredientUUID: uuid.MustParse("2c98fff4-7ccc-4536-8259-67a88380e99b"),
				Amount:         1,
				Unit:           "kg",
			},
			{
				IngredientUUID: uuid.MustParse("2c98fff4-7ccc-4536-8259-67a88380e99b"),
				Amount:         1,
				Unit:           "unit",
			},
		},
		Instructions: []RecipeInstruction{
			{
				StepNum:     1,
				Instruction: "Chop kimchi, onion and pork belly",
			},
			{
				StepNum:     2,
				Instruction: "grill the pan and put some oil on it",
			},
		},
	}

	badRecipe := Recipe{
		UserUUID:     uuid.MustParse("2c98fff4-7ccc-4536-8259-67a88380e99c"),
		RecipeName:   "kimchi fried rice",
		Category:     "",
		Ingredients:  []RecipeIngredient{},
		Instructions: []RecipeInstruction{},
	}

	testcases := []struct {
		name                     string
		updateRecipeOverrideFunc func(ctx context.Context, r store.Recipe) (*store.Recipe, error)
		requestBody              Recipe
		expectedReponse          *Recipe
		expectedStatus           int
	}{
		{
			"happyPath",
			nil,
			goodRecipe,
			&goodRecipe,
			http.StatusOK,
		},
		{
			"badRequest",
			nil,
			badRecipe,
			nil,
			http.StatusBadRequest,
		},
		{
			"internalServerError",
			func(ctx context.Context, r store.Recipe) (*store.Recipe, error) {
				return nil, errors.New("internalServerError")
			},
			goodRecipe,
			nil,
			http.StatusInternalServerError,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			reqBody, err := json.Marshal(testcase.requestBody)
			assert.NoError(t, err, "unexpected error marshalling the request body")

			req := httptest.NewRequest(http.MethodPost, "/recipes/"+testcase.requestBody.RecipeUUID.String(), bytes.NewBuffer(reqBody))
			w := httptest.NewRecorder()

			testServer.db = &mockstore.Mockstore{UpdateRecipeOverride: testcase.updateRecipeOverrideFunc}
			testServer.r.ServeHTTP(w, req)

			assert.Equal(t, testcase.expectedStatus, w.Code)

			if testcase.expectedReponse != nil {
				var resBody Recipe

				err := json.Unmarshal(w.Body.Bytes(), &resBody)
				assert.NoError(t, err, "unexpected error unmarshalling the response body")

				assert.Equal(t, *testcase.expectedReponse, resBody)
			} else {
				assert.Equal(t, 0, w.Body.Len())
			}
		})
	}
}

func TestDeleteRecipe(t *testing.T) {
	testcases := []struct {
		name                     string
		deleteRecipeOverrideFunc func(ctx context.Context, id uuid.UUID) error
		requestBody              string
		expectedStatus           int
	}{
		{
			"happyPath",
			nil,
			"080b5f09-527b-4581-bb56-19adbfe50ebf",
			http.StatusOK,
		},
		{
			"badRequest",
			nil,
			"maerong",
			http.StatusBadRequest,
		},
		{
			"internalServerError",
			func(ctx context.Context, id uuid.UUID) error {
				return errors.New("internalServerError")
			},
			"080b5f09-527b-4581-bb56-19adbfe50ebf",
			http.StatusInternalServerError,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/recipes/"+testcase.requestBody, nil)
			w := httptest.NewRecorder()

			testServer.db = &mockstore.Mockstore{DeleteRecipeOverride: testcase.deleteRecipeOverrideFunc}
			testServer.r.ServeHTTP(w, req)

			assert.Equal(t, testcase.expectedStatus, w.Code)
		})
	}
}
