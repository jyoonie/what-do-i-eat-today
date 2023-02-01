package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"wdiet/store"
	"wdiet/store/mockstore"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var l, _ = zap.NewProduction()

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

			var resBody User

			if testcase.expectedResponse != nil {
				err := json.Unmarshal(w.Body.Bytes(), &resBody)
				assert.NoError(t, err, "unexpected error unmarshalling the response body")

				assert.Equal(t, testcase.expectedResponse.UserUUID, resBody.UserUUID)
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

	badUserRequest := createUserRequest{
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
			badUserRequest,
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

			var resBody User

			if testcase.expectedResponse != nil {
				err := json.Unmarshal(w.Body.Bytes(), &resBody)
				assert.NoError(t, err, "unexpected error unmarshalling the response body")

				assert.NotEqual(t, resBody.UserUUID, uuid.Nil) //uuid가 랜덤으로 generate 됐기 때문에, 그냥 response uuid가 uuid Nil이 아닌지만 검사.
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

	badUserRequest := User{
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
			badUserRequest,
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

			var resBody User

			if testcase.expectedResponse != nil {
				err = json.Unmarshal(w.Body.Bytes(), &resBody)
				assert.NoError(t, err, "unexpected error unmarshalling the response body")

				assert.Equal(t, testcase.expectedResponse.UserUUID, resBody.UserUUID)
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
				return nil, errors.New("internalservererror")
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

			var resBody Ingredient

			if testcase.expectedResponse != nil {
				err := json.Unmarshal(w.Body.Bytes(), &resBody)
				assert.NoError(t, err, "unexpected error unmarshalling the response body")

				assert.Equal(t, testcase.expectedResponse.IngredientUUID, resBody.IngredientUUID)
				assert.Equal(t, testcase.expectedResponse.IngredientName, resBody.IngredientName)
				assert.Equal(t, testcase.expectedResponse.Category, resBody.Category)
				assert.Equal(t, testcase.expectedResponse.DaysUntilExp, resBody.DaysUntilExp)
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
				return nil, errors.New("internalservererror")
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

			var resBody Ingredient

			if testcase.expectedResponse != nil {
				err := json.Unmarshal(w.Body.Bytes(), &resBody)
				assert.NoError(t, err, "unexpected error unmarshalling the response body")

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

	badIngrRequest := Ingredient{
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
			badIngrRequest,
			nil,
			http.StatusBadRequest,
		},
		{
			"internalServerError",
			func(ctx context.Context, i store.Ingredient) (*store.Ingredient, error) {
				return nil, errors.New("internalservererror")
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

			var resBody Ingredient

			if testcase.expectedResponse != nil {
				err := json.Unmarshal(w.Body.Bytes(), &resBody)
				assert.NoError(t, err, "unexpected error unmarshalling the response body")

				assert.Equal(t, testcase.expectedResponse.IngredientName, resBody.IngredientName)
				assert.Equal(t, testcase.expectedResponse.Category, resBody.Category)
				assert.Equal(t, testcase.expectedResponse.DaysUntilExp, resBody.DaysUntilExp)
				assert.Equal(t, testcase.expectedResponse.IngredientUUID, resBody.IngredientUUID)
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
				return errors.New("internalservererror")
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

func TestSearchIngredients(t *testing.T) {
	testcases := []struct {
		name                          string
		searchIngredientsOverrideFunc func(ctx context.Context, i store.Ingredient) ([]store.Ingredient, error)
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
			func(ctx context.Context, i store.Ingredient) ([]store.Ingredient, error) {
				return []store.Ingredient{}, errors.New("internalservererror")
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

			var resBody []Ingredient

			if testcase.expectedResponse != nil {
				err := json.Unmarshal(w.Body.Bytes(), &resBody)
				assert.NoError(t, err, "unexpected error unmarshaling the response body")

				assert.Equal(t, len(testcase.expectedResponse), len(resBody))

				for i := range testcase.expectedResponse {
					assert.Equal(t, testcase.expectedResponse[i].IngredientUUID, resBody[i].IngredientUUID)
					assert.Equal(t, testcase.expectedResponse[i].IngredientName, resBody[i].IngredientName)
					assert.Equal(t, testcase.expectedResponse[i].Category, resBody[i].Category)
					assert.Equal(t, testcase.expectedResponse[i].DaysUntilExp, resBody[i].DaysUntilExp)
				}
			} else {
				assert.Equal(t, 0, w.Body.Len())
			}
		})
	}
}
