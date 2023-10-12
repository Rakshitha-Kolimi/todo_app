package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"todo-project/constants"
	"todo-project/entity"
)

func createJwtToken(username, id string) (*jwt.Token, string) {
	claims := jwt.MapClaims{
		"name":    username,
		"user_id": id,
		"exp":     jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
	}

	jwtSecretKey := os.Getenv("JWT_AUTH_SECRET")

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	rawToken, err := token.SignedString([]byte(jwtSecretKey))
	if err != nil {
		return nil, ""
	}

	return token, rawToken
}

func TestApiService(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}

	authRepo := &ApiRepository{
		DB: db,
	}

	as := ApiService{
		R: *authRepo,
	}

	dueDate, err := time.Parse("2006-01-02 15:04:05", "2023-10-15 09:38:24")
	if err != nil {
		t.Fatalf("Failed to parse due_date: %v", err)
	}

	createdTime, err := time.Parse("2006-01-02 15:04:05", "2023-10-15 09:38:24")
	if err != nil {
		t.Fatalf("Failed to parse created_at: %v", err)
	}

	updatedAt, err := time.Parse("2006-01-02 15:04:05", "2023-09-29 00:01:19.000000")
	if err != nil {
		t.Fatalf("Failed to parse created_at: %v", err)
	}

	defer db.Close()

	e := echo.New()

	username := "example@gmail.com"
	user_id := "ed6caeda-1fa9-442e-a41d-dd2b135cea67"

	t.Run("Is item of the user", func(t *testing.T) {
		itemID := "item123"
		userID := "user123"
		expectedQuery := `SELECT EXISTS \(SELECT 1 FROM todo_items WHERE id='item123' AND user_id='user123'\) AS is_current_user`

		t.Run("Verifies that the item belongs to the user", func(t *testing.T) {
			rows := mock.NewRows([]string{"is_current_user"}).AddRow(true)
			mock.ExpectQuery(expectedQuery).WillReturnRows(rows)

			got, err := authRepo.IsItemOftheUser(itemID, userID)

			if err != nil {
				t.Fatalf("Error executing IsItemOftheUser: %v", err)
			}

			assert.True(t, got, "Expected the item to belong to the user")
			assert.NoError(t, mock.ExpectationsWereMet(), "Expectations were not met")
		})

		t.Run("Verifies that the item doesnot belongs to the user", func(t *testing.T) {
			rows := mock.NewRows([]string{"is_current_user"}).AddRow(false)
			mock.ExpectQuery(expectedQuery).WillReturnRows(rows)

			got, err := authRepo.IsItemOftheUser(itemID, userID)

			if err != nil {
				t.Fatalf("Error executing IsItemOftheUser: %v", err)
			}

			assert.False(t, got, "Expected the item to belong to the user")
			assert.NoError(t, mock.ExpectationsWereMet(), "Expectations were not met")
		})

		t.Run("Verifies that the error is thrown when cannot find if the item belongs to user", func(t *testing.T) {
			err := errors.New("internal server error")
			mock.ExpectQuery(expectedQuery).WillReturnError(err)

			_, err = authRepo.IsItemOftheUser(itemID, userID)

			if err == nil {
				t.Fatalf("Error must be thrown")
			}

			assert.NoError(t, mock.ExpectationsWereMet(), "Expectations were not met")
		})
	})

	t.Run("Test Create todo item", func(t *testing.T) {
		t.Run("Cannot get jwt token", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/item/create", strings.NewReader(`{
				"details": {
					"description": "This is description for the todo list item.",
					"due_date": "2023-05-22T09:38:24.405027Z",
					"priority": "HIGH"
				},
				"name": "Todo list 1"
			}`))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set(echo.HeaderAuthorization, "Bearer INVALID_JWT_TOKEN")
			rec := httptest.NewRecorder()

			ctx := e.NewContext(req, rec)

			_ = as.CreateTodoItem(ctx)

			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		})

		t.Run("Unauthorized", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/item/create", strings.NewReader(`{
				"details": {
					"description": "This is description for the todo list item.",
					"due_date": "2023-05-22T09:38:24.405027Z",
					"priority": "HIGH"
				},
				"name": "Todo list 1"
			}`))

			token := &jwt.Token{
				Claims: jwt.MapClaims{},
			}

			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set(echo.HeaderAuthorization, "Bearer"+" "+token.Raw)

			rec := httptest.NewRecorder()

			ctx := e.NewContext(req, rec)
			ctx.Set("user", token)

			_ = as.CreateTodoItem(ctx)

			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		})

		t.Run("Internal server error - Bind", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/item/create", strings.NewReader(`{
				"details": {
					"description": "This is description for the todo list item.",
					"due_date": "2023-05-22T09:38:24.405027Z",
					"priority": "HIGH"
				},
				"name": "Todo list 1"
			}`))

			token, rawToken := createJwtToken(username, user_id)

			if err != nil {
				t.Errorf("Error creating JWT token: %v", err)
			}

			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set(echo.HeaderAuthorization, "Bearer"+" "+rawToken)

			rec := httptest.NewRecorder()

			ctx := e.NewContext(req, rec)
			ctx.Set("user", token)
			ctx.Bind(nil)

			_ = as.CreateTodoItem(ctx)

			assert.Equal(t, http.StatusInternalServerError, rec.Code)
		})
		t.Run("Validation error", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/item/create", strings.NewReader(`{
				"details": {
					"description": "This is description for the todo list item.",
					"due_date": "2023-05-22T09:38:24.405027Z",
					"priority": "High"
				},
				"name": ""
			}`))

			token, rawToken := createJwtToken(username, user_id)

			if err != nil {
				t.Errorf("Error creating JWT token: %v", err)
			}

			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set(echo.HeaderAuthorization, "Bearer"+" "+rawToken)

			rec := httptest.NewRecorder()

			ctx := e.NewContext(req, rec)
			ctx.Set("user", token)

			_ = as.CreateTodoItem(ctx)

			assert.Equal(t, http.StatusBadRequest, rec.Code)
		})

		t.Run("Create item success", func(t *testing.T) {
			rows := mock.NewRows([]string{"id", "name", "description", "due_date", "priority", "created_at", "updated_at", "is_completed", "is_deleted", "user_id"}).
				AddRow("3a35452e-957c-4588-8d40-c88f370067d2", "Todo list item 1", "This is item 1", dueDate, "HIGH", createdTime, updatedAt, true, false, "ed6caeda-1fa9-442e-a41d-dd2b135cea67")

			var item = &entity.TodoItemDto{
				Name: "Todo list item 1",
				Details: entity.TodoItemDetailsDto{
					Description: "This is item 1",
					Priority:    "HIGH",
					DueDate:     dueDate,
				},
			}

			mock.ExpectQuery("INSERT INTO todo_items").
				WithArgs(sqlmock.AnyArg(), item.Name, item.Details.Description, sqlmock.AnyArg(), item.Details.Priority, sqlmock.AnyArg(), sqlmock.AnyArg(), user_id).
				WillReturnRows(rows)

			req := httptest.NewRequest(http.MethodPost, "/item/create",
				strings.NewReader(`{
					"details": {
						"description": "This is item 1",
						"due_date": "2023-05-22T09:38:24.405027Z",
						"priority": "HIGH"
					},
					"name": "Todo list item 1"
				}`))

			token, rawToken := createJwtToken(username, user_id)

			if err != nil {
				t.Errorf("Error creating JWT token: %v", err)
			}

			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set(echo.HeaderAuthorization, "Bearer"+" "+rawToken)

			rec := httptest.NewRecorder()

			ctx := e.NewContext(req, rec)
			ctx.Set("user", token)

			err := as.CreateTodoItem(ctx)
			if err != nil {
				t.Errorf("Error while creatung a todo item: %v", err)
			}

			var result entity.TodoItem
			err = json.Unmarshal(rec.Body.Bytes(), &result)
			if err != nil {
				t.Errorf("Error unmarshalling response: %v", err)
			}
			assert.NoError(t, err)

			want := entity.TodoItem{
				Id: "3a35452e-957c-4588-8d40-c88f370067d2",
				Item: entity.TodoItemDto{
					Name: "Todo list item 1",
					Details: entity.TodoItemDetailsDto{
						Description: "This is item 1",
						Priority:    "HIGH",
						DueDate:     dueDate,
					},
				},
				IsCompleted: true,
				IsDeleted:   false,
				CreatedAt:   createdTime,
				UpdatedAt:   updatedAt,
			}

			assert.Equal(t, want, result)
			assert.Equal(t, http.StatusCreated, rec.Code)
		})

		t.Run("Create item error", func(t *testing.T) {
			var item = &entity.TodoItemDto{
				Name: "Todo list item 1",
				Details: entity.TodoItemDetailsDto{
					Description: "This is item 1",
					Priority:    "HIGH",
					DueDate:     dueDate,
				},
			}

			mock.ExpectQuery("INSERT INTO todo_items").
				WithArgs(sqlmock.AnyArg(), item.Name, item.Details.Description, sqlmock.AnyArg(), item.Details.Priority, sqlmock.AnyArg(), sqlmock.AnyArg(), user_id).
				WillReturnError(errors.New("bad request"))

			mock.ExpectQuery("SELECT").WillReturnError(errors.New("bad request"))

			req := httptest.NewRequest(http.MethodPost, "/item/create",
				strings.NewReader(`{
					"details": {
						"description": "This is item 1",
						"due_date": "2023-05-22T09:38:24.405027Z",
						"priority": "HIGH"
					},
					"name": "Todo list item 1"
				}`))

			token, rawToken := createJwtToken(username, user_id)

			if err != nil {
				t.Errorf("Error creating JWT token: %v", err)
			}

			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set(echo.HeaderAuthorization, "Bearer"+" "+rawToken)

			rec := httptest.NewRecorder()

			ctx := e.NewContext(req, rec)
			ctx.Set("user", token)

			as.CreateTodoItem(ctx)

			assert.Equal(t, http.StatusBadRequest, rec.Code)
		})
	})

	t.Run("Test Find by id", func(t *testing.T) {
		t.Run("Cannot get jwt token", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/item/3a35452e-957c-4588-8d40-c88f370067d20", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set(echo.HeaderAuthorization, "Bearer INVALID_JWT_TOKEN") 
			rec := httptest.NewRecorder()

			ctx := e.NewContext(req, rec)

			_ = as.FindById(ctx)

			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		})
		t.Run("Unauthorized", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/item/3a35452e-957c-4588-8d40-c88f370067d20", nil)
			token := &jwt.Token{
				Claims: jwt.MapClaims{},
			}

			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set(echo.HeaderAuthorization, "Bearer"+" "+token.Raw)

			rec := httptest.NewRecorder()

			ctx := e.NewContext(req, rec)
			ctx.Set("user", token)

			_ = as.FindById(ctx)

			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		})

		t.Run("Not Found", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/item/3a35452e-957c-4588-8d40-c88f370067d20", nil)

			token, rawToken := createJwtToken(username, user_id)

			if err != nil {
				t.Errorf("Error creating JWT token: %v", err)
			}

			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set(echo.HeaderAuthorization, "Bearer"+" "+rawToken)

			rec := httptest.NewRecorder()

			ctx := e.NewContext(req, rec)
			ctx.Set("user", token)

			_ = as.FindById(ctx)

			assert.Equal(t, http.StatusNotFound, rec.Code)
		})

		t.Run("Internal Server error - Cannot find the item belongs to user", func(t *testing.T) {
			rows := mock.NewRows([]string{"id", "name", "description", "due_date", "priority", "created_at", "updated_at", "is_completed", "is_deleted", "user_id"}).
				AddRow("3a35452e-957c-4588-8d40-c88f370067d2", "Todo list item 1", "This is item 1", dueDate, "HIGH", createdTime, updatedAt, true, false, "ed6caeda-1fa9-442e-a41d-dd2b135cea67")

			mock.ExpectQuery(`SELECT \* FROM todo_items WHERE id='3a35452e-957c-4588-8d40-c88f370067d2'`).
				WillReturnRows(rows)

			req := httptest.NewRequest(http.MethodGet, "/item/3a35452e-957c-4588-8d40-c88f370067d2", nil)

			token, rawToken := createJwtToken(username, user_id)

			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set(echo.HeaderAuthorization, "Bearer"+" "+rawToken)

			rec := httptest.NewRecorder()

			ctx := e.NewContext(req, rec)
			ctx.Set("user", token)
			ctx.SetParamNames("id")
			ctx.SetParamValues("3a35452e-957c-4588-8d40-c88f370067d2")

			err := as.FindById(ctx)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusInternalServerError, rec.Code)
		})

		t.Run("Forbidden", func(t *testing.T) {
			rows := mock.NewRows([]string{"id", "name", "description", "due_date", "priority", "created_at", "updated_at", "is_completed", "is_deleted", "user_id"}).
				AddRow("3a35452e-957c-4588-8d40-c88f370067d2", "Todo list item 1", "This is item 1", dueDate, "HIGH", createdTime, updatedAt, true, false, "ed6caeda-1fa9-442e-a41d-dd2b135cea67")

			mock.ExpectQuery(`SELECT \* FROM todo_items WHERE id='3a35452e-957c-4588-8d40-c88f370067d2'`).
				WillReturnRows(rows)

			expectedQuery := `SELECT EXISTS \(SELECT 1 FROM todo_items WHERE id='3a35452e-957c-4588-8d40-c88f370067d2' AND user_id='ed6caeda-1fa9-442e-a41d-dd2b135cea67'\) AS is_current_user`
			rows = mock.NewRows([]string{"is_current_user"}).AddRow(false)
			mock.ExpectQuery(expectedQuery).WillReturnRows(rows)

			req := httptest.NewRequest(http.MethodGet, "/item/3a35452e-957c-4588-8d40-c88f370067d2", nil)

			token, rawToken := createJwtToken(username, user_id)

			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set(echo.HeaderAuthorization, "Bearer"+" "+rawToken)

			rec := httptest.NewRecorder()

			ctx := e.NewContext(req, rec)
			ctx.Set("user", token)
			ctx.SetParamNames("id")
			ctx.SetParamValues("3a35452e-957c-4588-8d40-c88f370067d2")

			err := as.FindById(ctx)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusForbidden, rec.Code)
		})
		t.Run("Success", func(t *testing.T) {
			rows := mock.NewRows([]string{"id", "name", "description", "due_date", "priority", "created_at", "updated_at", "is_completed", "is_deleted", "user_id"}).
				AddRow("3a35452e-957c-4588-8d40-c88f370067d2", "Todo list item 1", "This is item 1", dueDate, "HIGH", createdTime, updatedAt, true, false, "ed6caeda-1fa9-442e-a41d-dd2b135cea67")

			mock.ExpectQuery(`SELECT \* FROM todo_items WHERE id='3a35452e-957c-4588-8d40-c88f370067d2'`).
				WillReturnRows(rows)

			expectedQuery := `SELECT EXISTS \(SELECT 1 FROM todo_items WHERE id='3a35452e-957c-4588-8d40-c88f370067d2' AND user_id='ed6caeda-1fa9-442e-a41d-dd2b135cea67'\) AS is_current_user`
			rows = mock.NewRows([]string{"is_current_user"}).AddRow(true)
			mock.ExpectQuery(expectedQuery).WillReturnRows(rows)

			req := httptest.NewRequest(http.MethodGet, "/item/3a35452e-957c-4588-8d40-c88f370067d2", nil)

			token, rawToken := createJwtToken(username, user_id)

			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set(echo.HeaderAuthorization, "Bearer"+" "+rawToken)

			rec := httptest.NewRecorder()

			ctx := e.NewContext(req, rec)
			ctx.Set("user", token)
			ctx.SetParamNames("id")
			ctx.SetParamValues("3a35452e-957c-4588-8d40-c88f370067d2")

			err := as.FindById(ctx)
			assert.NoError(t, err)

			var result entity.TodoItem
			err = json.Unmarshal(rec.Body.Bytes(), &result)
			if err != nil {
				t.Errorf("Error unmarshalling response: %v", err)
			}
			assert.NoError(t, err)

			want := entity.TodoItem{
				Id: "3a35452e-957c-4588-8d40-c88f370067d2",
				Item: entity.TodoItemDto{
					Name: "Todo list item 1",
					Details: entity.TodoItemDetailsDto{
						Description: "This is item 1",
						Priority:    "HIGH",
						DueDate:     dueDate,
					},
				},
				IsCompleted: true,
				IsDeleted:   false,
				CreatedAt:   createdTime,
				UpdatedAt:   updatedAt,
			}

			assert.Equal(t, want, result)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)
		})
	})

	t.Run("Test Get all items of the user", func(t *testing.T) {
		t.Run("Cannot get jwt token", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/item/list", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set(echo.HeaderAuthorization, "Bearer INVALID_JWT_TOKEN")
			rec := httptest.NewRecorder()

			ctx := e.NewContext(req, rec)
			_ = as.GetAllItems(ctx)

			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		})
		t.Run("Unauthorized", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/item/list", nil)
			token := &jwt.Token{
				Claims: jwt.MapClaims{},
			}

			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set(echo.HeaderAuthorization, "Bearer"+" "+token.Raw)

			rec := httptest.NewRecorder()

			ctx := e.NewContext(req, rec)
			ctx.Set("user", token)

			_ = as.GetAllItems(ctx)

			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		})

		t.Run("Internal server error - bind error", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/item/list", strings.NewReader(`{
				"limit":1
			}`))

			token, rawToken := createJwtToken(username, user_id)

			if err != nil {
				t.Errorf("Error creating JWT token: %v", err)
			}

			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set(echo.HeaderAuthorization, "Bearer"+" "+rawToken)

			rec := httptest.NewRecorder()

			ctx := e.NewContext(req, rec)
			ctx.Set("user", token)
			ctx.Bind(nil)

			_ = as.GetAllItems(ctx)

			assert.Equal(t, http.StatusInternalServerError, rec.Code)
		})

		t.Run("Bad request", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/item/list", nil)

			token, rawToken := createJwtToken(username, user_id)

			if err != nil {
				t.Errorf("Error creating JWT token: %v", err)
			}

			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set(echo.HeaderAuthorization, "Bearer"+" "+rawToken)

			rec := httptest.NewRecorder()

			ctx := e.NewContext(req, rec)
			ctx.Set("user", token)

			_ = as.GetAllItems(ctx)

			assert.Equal(t, http.StatusBadRequest, rec.Code)
		})

		t.Run("Success", func(t *testing.T) {
			rows := mock.NewRows([]string{"id", "name", "description", "due_date", "priority", "created_at", "updated_at", "is_completed", "is_deleted", "user_id"}).
				AddRow("3a35452e-957c-4588-8d40-c88f370067d2", "Todo list item 1", "This is item 1", dueDate, "HIGH", createdTime, updatedAt, true, false, "ed6caeda-1fa9-442e-a41d-dd2b135cea67")
			GetItemsQuery := `SELECT \* FROM todo_items WHERE is_deleted = false AND user_id='ed6caeda-1fa9-442e-a41d-dd2b135cea67' LIMIT 1`
			mock.ExpectQuery(GetItemsQuery).WillReturnRows(rows)

			req := httptest.NewRequest(http.MethodPost, "/item/list", strings.NewReader(
				`{"limit":1}`,
			))

			token, rawToken := createJwtToken(username, user_id)

			if err != nil {
				t.Errorf("Error creating JWT token: %v", err)
			}

			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set(echo.HeaderAuthorization, "Bearer"+" "+rawToken)

			rec := httptest.NewRecorder()

			ctx := e.NewContext(req, rec)
			ctx.Set("user", token)

			_ = as.GetAllItems(ctx)

			var result []entity.TodoItem
			err = json.Unmarshal(rec.Body.Bytes(), &result)
			if err != nil {
				t.Errorf("Error unmarshalling response: %v", err)
			}
			assert.NoError(t, err)

			item := entity.TodoItem{
				Id: "3a35452e-957c-4588-8d40-c88f370067d2",
				Item: entity.TodoItemDto{
					Name: "Todo list item 1",
					Details: entity.TodoItemDetailsDto{
						Description: "This is item 1",
						Priority:    "HIGH",
						DueDate:     dueDate,
					},
				},
				IsCompleted: true,
				IsDeleted:   false,
				CreatedAt:   createdTime,
				UpdatedAt:   updatedAt,
			}

			assert.Len(t, result, 1)
			assert.Equal(t, item, result[0])
		})
	})

	t.Run("Test Update item by id", func(t *testing.T) {
		t.Run("Cannot get jwt token", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/item/3a35452e-957c-4588-8d40-c88f370067d20", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set(echo.HeaderAuthorization, "Bearer INVALID_JWT_TOKEN") 
			rec := httptest.NewRecorder()

			ctx := e.NewContext(req, rec)
			_ = as.UpdateItemById(ctx)
			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		})
		t.Run("Unauthorized", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/item/3a35452e-957c-4588-8d40-c88f370067d20", nil)
			token := &jwt.Token{
				Claims: jwt.MapClaims{},
			}

			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set(echo.HeaderAuthorization, "Bearer"+" "+token.Raw)

			rec := httptest.NewRecorder()

			ctx := e.NewContext(req, rec)
			ctx.Set("user", token)

			_ = as.UpdateItemById(ctx)

			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		})
		t.Run("Internal Server error - Cannot find the item belongs to user", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/item/3a35452e-957c-4588-8d40-c88f370067d2", nil)

			token, rawToken := createJwtToken(username, user_id)

			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set(echo.HeaderAuthorization, "Bearer"+" "+rawToken)

			rec := httptest.NewRecorder()

			ctx := e.NewContext(req, rec)
			ctx.Set("user", token)
			ctx.SetParamNames("id")
			ctx.SetParamValues("3a35452e-957c-4588-8d40-c88f370067d2")

			err := as.UpdateItemById(ctx)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusInternalServerError, rec.Code)
		})
		t.Run("Forbidden", func(t *testing.T) {
			expectedQuery := `SELECT EXISTS \(SELECT 1 FROM todo_items WHERE id='3a35452e-957c-4588-8d40-c88f370067d2' AND user_id='ed6caeda-1fa9-442e-a41d-dd2b135cea67'\) AS is_current_user`
			rows := mock.NewRows([]string{"is_current_user"}).AddRow(false)
			mock.ExpectQuery(expectedQuery).WillReturnRows(rows)

			req := httptest.NewRequest(http.MethodGet, "/item/3a35452e-957c-4588-8d40-c88f370067d2", nil)

			token, rawToken := createJwtToken(username, user_id)

			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set(echo.HeaderAuthorization, "Bearer"+" "+rawToken)

			rec := httptest.NewRecorder()

			ctx := e.NewContext(req, rec)
			ctx.Set("user", token)
			ctx.SetParamNames("id")
			ctx.SetParamValues("3a35452e-957c-4588-8d40-c88f370067d2")

			err := as.UpdateItemById(ctx)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusForbidden, rec.Code)
		})
		t.Run("Bad Request - bind error", func(t *testing.T) {
			expectedQuery := `SELECT EXISTS \(SELECT 1 FROM todo_items WHERE id='3a35452e-957c-4588-8d40-c88f370067d2' AND user_id='ed6caeda-1fa9-442e-a41d-dd2b135cea67'\) AS is_current_user`
			rows := mock.NewRows([]string{"is_current_user"}).AddRow(true)
			mock.ExpectQuery(expectedQuery).WillReturnRows(rows)

			req := httptest.NewRequest(http.MethodGet, "/item/3a35452e-957c-4588-8d40-c88f370067d2", strings.NewReader(`{
					"description": "This is description for the todo list item.",
					"due_date": "2023-05-22T09:38:24.405027Z",
					"priority": "HIGH"
				}`))

			token, rawToken := createJwtToken(username, user_id)

			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set(echo.HeaderAuthorization, "Bearer"+" "+rawToken)

			rec := httptest.NewRecorder()

			ctx := e.NewContext(req, rec)
			ctx.Set("user", token)
			ctx.SetParamNames("id")
			ctx.SetParamValues("3a35452e-957c-4588-8d40-c88f370067d2")
			ctx.Bind(nil)

			err := as.UpdateItemById(ctx)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		})

		t.Run("Bad Request - validation error", func(t *testing.T) {
			expectedQuery := `SELECT EXISTS \(SELECT 1 FROM todo_items WHERE id='3a35452e-957c-4588-8d40-c88f370067d2' AND user_id='ed6caeda-1fa9-442e-a41d-dd2b135cea67'\) AS is_current_user`
			rows := mock.NewRows([]string{"is_current_user"}).AddRow(true)
			mock.ExpectQuery(expectedQuery).WillReturnRows(rows)

			req := httptest.NewRequest(http.MethodGet, "/item/3a35452e-957c-4588-8d40-c88f370067d2", nil)

			token, rawToken := createJwtToken(username, user_id)

			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set(echo.HeaderAuthorization, "Bearer"+" "+rawToken)

			rec := httptest.NewRecorder()

			ctx := e.NewContext(req, rec)
			ctx.Set("user", token)
			ctx.SetParamNames("id")
			ctx.SetParamValues("3a35452e-957c-4588-8d40-c88f370067d2")

			err := as.UpdateItemById(ctx)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		})
		t.Run("Internal server error", func(t *testing.T) {
			expectedQuery := `SELECT EXISTS \(SELECT 1 FROM todo_items WHERE id='3a35452e-957c-4588-8d40-c88f370067d2' AND user_id='ed6caeda-1fa9-442e-a41d-dd2b135cea67'\) AS is_current_user`
			rows := mock.NewRows([]string{"is_current_user"}).AddRow(true)
			mock.ExpectQuery(expectedQuery).WillReturnRows(rows)

			req := httptest.NewRequest(http.MethodGet, "/item/3a35452e-957c-4588-8d40-c88f370067d2", strings.NewReader(`{
					  "description": "This is description for the todo list item.",
					  "due_date": "2023-05-22T09:38:24.405027Z",
					  "priority": "HIGH"
				  }`))

			token, rawToken := createJwtToken(username, user_id)

			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set(echo.HeaderAuthorization, "Bearer"+" "+rawToken)

			rec := httptest.NewRecorder()

			ctx := e.NewContext(req, rec)
			ctx.Set("user", token)
			ctx.SetParamNames("id")
			ctx.SetParamValues("3a35452e-957c-4588-8d40-c88f370067d2")

			err := as.UpdateItemById(ctx)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusInternalServerError, rec.Code)
		})
		t.Run("Success", func(t *testing.T) {
			expectedQuery := `SELECT EXISTS \(SELECT 1 FROM todo_items WHERE id='3a35452e-957c-4588-8d40-c88f370067d2' AND user_id='ed6caeda-1fa9-442e-a41d-dd2b135cea67'\) AS is_current_user`
			rows := mock.NewRows([]string{"is_current_user"}).AddRow(true)
			mock.ExpectQuery(expectedQuery).WillReturnRows(rows)

			newUpdatedAtString := time.Now().Format("2006-01-02T15:04:05Z07:00")
			newUpdatedDate, _ := time.Parse(newUpdatedAtString, "2006-01-02T15:04:05Z07:00")

			rows = mock.NewRows([]string{"id", "name", "description", "due_date", "priority", "created_at", "updated_at", "is_completed", "is_deleted", "user_id"}).
				AddRow("3a35452e-957c-4588-8d40-c88f370067d2", "Todo list item 1", "This is description for the todo list item.", dueDate, "HIGH", createdTime, newUpdatedDate, true, false, "ed6caeda-1fa9-442e-a41d-dd2b135cea67")
			newUpdatedQuery := `UPDATE todo_items SET description=\$1, due_date=\$2, priority=\$3, updated_at=\$4 WHERE id=\$5 RETURNING \*;`

			mock.ExpectQuery(newUpdatedQuery).
				WithArgs("This is description for the todo list item.", dueDate, "HIGH", newUpdatedAtString, "3a35452e-957c-4588-8d40-c88f370067d2").
				WillReturnRows(rows)

			req := httptest.NewRequest(http.MethodGet, "/item/3a35452e-957c-4588-8d40-c88f370067d2", strings.NewReader(`{
					  "description": "This is description for the todo list item.",
					  "due_date": "2023-10-15T09:38:24Z",
					  "priority": "HIGH"
				  }`))

			token, rawToken := createJwtToken(username, user_id)

			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set(echo.HeaderAuthorization, "Bearer"+" "+rawToken)

			rec := httptest.NewRecorder()

			ctx := e.NewContext(req, rec)
			ctx.Set("user", token)
			ctx.SetParamNames("id")
			ctx.SetParamValues("3a35452e-957c-4588-8d40-c88f370067d2")

			err := as.UpdateItemById(ctx)

			assert.NoError(t, err)

			want := entity.TodoItem{
				Id: "3a35452e-957c-4588-8d40-c88f370067d2",
				Item: entity.TodoItemDto{
					Name: "Todo list item 1",
					Details: entity.TodoItemDetailsDto{
						Description: "This is description for the todo list item.",
						Priority:    "HIGH",
						DueDate:     dueDate,
					},
				},
				IsCompleted: true,
				IsDeleted:   false,
				CreatedAt:   createdTime,
				UpdatedAt:   newUpdatedDate,
			}

			var result entity.TodoItem
			err = json.Unmarshal(rec.Body.Bytes(), &result)
			if err != nil {
				t.Errorf("Error unmarshalling response: %v", err)
			}
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, want, result)
		})
	})

	t.Run("Test Delete by id", func(t *testing.T) {
		t.Run("Cannot get jwt token", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/item/3a35452e-957c-4588-8d40-c88f370067d20", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set(echo.HeaderAuthorization, "Bearer INVALID_JWT_TOKEN") 
			rec := httptest.NewRecorder()

			ctx := e.NewContext(req, rec)
			_ = as.DeleteById(ctx)

			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		})
		t.Run("Unauthorized", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/item/3a35452e-957c-4588-8d40-c88f370067d20", nil)
			token := &jwt.Token{
				Claims: jwt.MapClaims{},
			}

			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set(echo.HeaderAuthorization, "Bearer"+" "+token.Raw)

			rec := httptest.NewRecorder()

			ctx := e.NewContext(req, rec)
			ctx.Set("user", token)

			_ = as.DeleteById(ctx)
			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		})

		t.Run("Internal Server error - Cannot find the item belongs to user", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/item/3a35452e-957c-4588-8d40-c88f370067d2", nil)

			token, rawToken := createJwtToken(username, user_id)

			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set(echo.HeaderAuthorization, "Bearer"+" "+rawToken)

			rec := httptest.NewRecorder()

			ctx := e.NewContext(req, rec)
			ctx.Set("user", token)
			ctx.SetParamNames("id")
			ctx.SetParamValues("3a35452e-957c-4588-8d40-c88f370067d2")

			err := as.DeleteById(ctx)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusInternalServerError, rec.Code)
		})

		t.Run("Forbidden", func(t *testing.T) {
			expectedQuery := `SELECT EXISTS \(SELECT 1 FROM todo_items WHERE id='3a35452e-957c-4588-8d40-c88f370067d2' AND user_id='ed6caeda-1fa9-442e-a41d-dd2b135cea67'\) AS is_current_user`
			rows := mock.NewRows([]string{"is_current_user"}).AddRow(false)
			mock.ExpectQuery(expectedQuery).WillReturnRows(rows)

			req := httptest.NewRequest(http.MethodGet, "/item/3a35452e-957c-4588-8d40-c88f370067d2", nil)

			token, rawToken := createJwtToken(username, user_id)

			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set(echo.HeaderAuthorization, "Bearer"+" "+rawToken)

			rec := httptest.NewRecorder()

			ctx := e.NewContext(req, rec)
			ctx.Set("user", token)
			ctx.SetParamNames("id")
			ctx.SetParamValues("3a35452e-957c-4588-8d40-c88f370067d2")

			err := as.DeleteById(ctx)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusForbidden, rec.Code)
		})
		t.Run("Internal server error-Delete Todo Item", func(t *testing.T) {
			expectedQuery := `SELECT EXISTS \(SELECT 1 FROM todo_items WHERE id='3a35452e-957c-4588-8d40-c88f370067d2' AND user_id='ed6caeda-1fa9-442e-a41d-dd2b135cea67'\) AS is_current_user`
			rows := mock.NewRows([]string{"is_current_user"}).AddRow(true)
			mock.ExpectQuery(expectedQuery).WillReturnRows(rows)

			req := httptest.NewRequest(http.MethodGet, "/item/3a35452e-957c-4588-8d40-c88f370067d2", nil)

			token, rawToken := createJwtToken(username, user_id)

			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set(echo.HeaderAuthorization, "Bearer"+" "+rawToken)

			rec := httptest.NewRecorder()

			ctx := e.NewContext(req, rec)
			ctx.Set("user", token)
			ctx.SetParamNames("id")
			ctx.SetParamValues("3a35452e-957c-4588-8d40-c88f370067d2")

			err := as.DeleteById(ctx)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusInternalServerError, rec.Code)
		})
		t.Run("Success", func(t *testing.T) {
			expectedQuery := `SELECT EXISTS \(SELECT 1 FROM todo_items WHERE id='3a35452e-957c-4588-8d40-c88f370067d2' AND user_id='ed6caeda-1fa9-442e-a41d-dd2b135cea67'\) AS is_current_user`
			rows := mock.NewRows([]string{"is_current_user"}).AddRow(true)
			mock.ExpectQuery(expectedQuery).WillReturnRows(rows)

			newUpdatedAtString := time.Now().Format("2006-01-02T15:04:05Z07:00")
			newDeleteQuery := `UPDATE todo_items SET is_deleted = 'true', updated_at=\$1 WHERE id=\$2;`

			mock.ExpectExec(newDeleteQuery).WithArgs(newUpdatedAtString, "3a35452e-957c-4588-8d40-c88f370067d2").
				WillReturnResult(sqlmock.NewResult(1, 1))

			req := httptest.NewRequest(http.MethodGet, "/item/3a35452e-957c-4588-8d40-c88f370067d2", nil)

			token, rawToken := createJwtToken(username, user_id)

			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set(echo.HeaderAuthorization, "Bearer"+" "+rawToken)

			rec := httptest.NewRecorder()

			ctx := e.NewContext(req, rec)
			ctx.Set("user", token)
			ctx.SetParamNames("id")
			ctx.SetParamValues("3a35452e-957c-4588-8d40-c88f370067d2")

			err := as.DeleteById(ctx)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, constants.DELETE_ITEM_SUCCESSFULL, rec.Body.String())
		})
	})
	t.Run("Test Update status to completed", func(t *testing.T) {
		t.Run("Cannot get jwt token", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/item/3a35452e-957c-4588-8d40-c88f370067d20", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set(echo.HeaderAuthorization, "Bearer INVALID_JWT_TOKEN") 
			rec := httptest.NewRecorder()

			ctx := e.NewContext(req, rec)
			_ = as.UpdateStatustoCompleted(ctx)

			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		})
		t.Run("Unauthorized", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/item/3a35452e-957c-4588-8d40-c88f370067d20", nil)
			token := &jwt.Token{
				Claims: jwt.MapClaims{},
			}

			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set(echo.HeaderAuthorization, "Bearer"+" "+token.Raw)

			rec := httptest.NewRecorder()

			ctx := e.NewContext(req, rec)
			ctx.Set("user", token)

			_ = as.UpdateStatustoCompleted(ctx)

			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		})

		t.Run("Internal Server error - Cannot find the item belongs to user", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/item/3a35452e-957c-4588-8d40-c88f370067d2", nil)

			token, rawToken := createJwtToken(username, user_id)

			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set(echo.HeaderAuthorization, "Bearer"+" "+rawToken)

			rec := httptest.NewRecorder()

			ctx := e.NewContext(req, rec)
			ctx.Set("user", token)
			ctx.SetParamNames("id")
			ctx.SetParamValues("3a35452e-957c-4588-8d40-c88f370067d2")

			err := as.UpdateStatustoCompleted(ctx)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusInternalServerError, rec.Code)
		})

		t.Run("Forbidden", func(t *testing.T) {
			expectedQuery := `SELECT EXISTS \(SELECT 1 FROM todo_items WHERE id='3a35452e-957c-4588-8d40-c88f370067d2' AND user_id='ed6caeda-1fa9-442e-a41d-dd2b135cea67'\) AS is_current_user`
			rows := mock.NewRows([]string{"is_current_user"}).AddRow(false)
			mock.ExpectQuery(expectedQuery).WillReturnRows(rows)

			req := httptest.NewRequest(http.MethodGet, "/item/3a35452e-957c-4588-8d40-c88f370067d2", nil)

			token, rawToken := createJwtToken(username, user_id)

			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set(echo.HeaderAuthorization, "Bearer"+" "+rawToken)

			rec := httptest.NewRecorder()

			ctx := e.NewContext(req, rec)
			ctx.Set("user", token)
			ctx.SetParamNames("id")
			ctx.SetParamValues("3a35452e-957c-4588-8d40-c88f370067d2")

			err := as.UpdateStatustoCompleted(ctx)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusForbidden, rec.Code)
		})
		t.Run("Internal server error-Delete Todo Item", func(t *testing.T) {
			expectedQuery := `SELECT EXISTS \(SELECT 1 FROM todo_items WHERE id='3a35452e-957c-4588-8d40-c88f370067d2' AND user_id='ed6caeda-1fa9-442e-a41d-dd2b135cea67'\) AS is_current_user`
			rows := mock.NewRows([]string{"is_current_user"}).AddRow(true)
			mock.ExpectQuery(expectedQuery).WillReturnRows(rows)

			req := httptest.NewRequest(http.MethodGet, "/item/3a35452e-957c-4588-8d40-c88f370067d2", nil)

			token, rawToken := createJwtToken(username, user_id)

			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set(echo.HeaderAuthorization, "Bearer"+" "+rawToken)

			rec := httptest.NewRecorder()

			ctx := e.NewContext(req, rec)
			ctx.Set("user", token)
			ctx.SetParamNames("id")
			ctx.SetParamValues("3a35452e-957c-4588-8d40-c88f370067d2")

			err := as.UpdateStatustoCompleted(ctx)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusInternalServerError, rec.Code)
		})
		t.Run("Success", func(t *testing.T) {
			expectedQuery := `SELECT EXISTS \(SELECT 1 FROM todo_items WHERE id='3a35452e-957c-4588-8d40-c88f370067d2' AND user_id='ed6caeda-1fa9-442e-a41d-dd2b135cea67'\) AS is_current_user`
			rows := mock.NewRows([]string{"is_current_user"}).AddRow(true)
			mock.ExpectQuery(expectedQuery).WillReturnRows(rows)

			rows = mock.NewRows([]string{"id", "name", "description", "due_date", "priority", "created_at", "updated_at", "is_completed", "is_deleted", "user_id"}).
				AddRow("3a35452e-957c-4588-8d40-c88f370067d2", "Todo list item 1", "This is item 1", dueDate, "HIGH", createdTime, updatedAt, true, false, "ed6caeda-1fa9-442e-a41d-dd2b135cea67")
			newUpdatedAtString := time.Now().Format("2006-01-02T15:04:05Z07:00")
			newCompleteQuery := `UPDATE todo_items SET is_completed = 'true', updated_at=\$1 WHERE id=\$2 RETURNING \*;`

			mock.ExpectQuery(newCompleteQuery).WithArgs(newUpdatedAtString, "3a35452e-957c-4588-8d40-c88f370067d2").
				WillReturnRows(rows)

			req := httptest.NewRequest(http.MethodGet, "/item/3a35452e-957c-4588-8d40-c88f370067d2", nil)

			token, rawToken := createJwtToken(username, user_id)

			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set(echo.HeaderAuthorization, "Bearer"+" "+rawToken)

			rec := httptest.NewRecorder()

			ctx := e.NewContext(req, rec)
			ctx.Set("user", token)
			ctx.SetParamNames("id")
			ctx.SetParamValues("3a35452e-957c-4588-8d40-c88f370067d2")

			err := as.UpdateStatustoCompleted(ctx)
			assert.NoError(t, err)

			var result entity.TodoItem
			err = json.Unmarshal(rec.Body.Bytes(), &result)
			if err != nil {
				t.Errorf("Error unmarshalling response: %v", err)
			}
			assert.NoError(t, err)

			want := entity.TodoItem{
				Id: "3a35452e-957c-4588-8d40-c88f370067d2",
				Item: entity.TodoItemDto{
					Name: "Todo list item 1",
					Details: entity.TodoItemDetailsDto{
						Description: "This is item 1",
						Priority:    "HIGH",
						DueDate:     dueDate,
					},
				},
				IsCompleted: true,
				IsDeleted:   false,
				CreatedAt:   createdTime,
				UpdatedAt:   updatedAt,
			}

			assert.Equal(t, want, result)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)
		})
	})
}
