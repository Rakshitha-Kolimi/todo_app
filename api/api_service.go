package api

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"

	"todo-project/authentication"
	"todo-project/constants"
	"todo-project/entity"
)

type Service interface {
	//Create an item
	CreateTodoItem(item *entity.TodoItemDto, userId string) (todoItem entity.TodoItem, err error)

	//Find a todo item by id
	FindItemById(id string) (todoItem entity.TodoItem, err error)

	//Get all the todo list item
	GetTodoListItems(limit int, userId string) (todoItemList []entity.TodoItem, err error)

	//Update an todo list item
	UpdateTodoItem(item *entity.TodoItemDetailsDto, id string) (todoItem entity.TodoItem, err error)

	//Delete an todo Item
	DeleteTodoItem(id string) error

	//Update status to Complete by Id
	SetItemStatusToComplete(id string) (todoItem entity.TodoItem, err error)
}

type ApiService struct {
	R ApiRepository
}

// Create an item to do
// @Summary creates a new todo item
// @Tags Item
// @Accept json
// @Produce json
// @Security JWT
// @Param Authorization header string false "Bearer"
// @Param dto body entity.TodoItemDto true "Item details"
// @Success 201 {object} entity.TodoItem
// @Failure 400 string constants.BAD_REQUEST
// @Failure 401 string constants.CANNOT_FETCH_THE_USER_ID
// @Failure 500 string constants.INTERNAL_SERVER_ERROR
// @Router /item/create [post]
func(as ApiService) CreateTodoItem(c echo.Context) error {
	t := new(entity.TodoItemDto)
	token,ok := c.Get("user").(*jwt.Token)
	
	if !ok {
		return c.String(http.StatusUnauthorized, constants.CANNOT_FETCH_THE_USER_ID)
	}

	userId, ok := authentication.GetUserFromToken(token)

	if !ok {
		return c.String(http.StatusUnauthorized, constants.CANNOT_FETCH_THE_USER_ID)
	}

	if err := c.Bind(t); err != nil {
		errMessage := fmt.Sprintf(constants.INTERNAL_SERVER_ERROR, err)
		return c.String(http.StatusInternalServerError, errMessage)
	}

	validate := validator.New()
	err := validate.Struct(t)

	if err != nil {
		errMessage := fmt.Sprintf(constants.BAD_REQUEST, err)
		return c.String(http.StatusBadRequest, errMessage)
	}

	item, err := as.R.CreateTodoItem(t, userId)
	if err != nil {
		errMessage := fmt.Sprintf(constants.CREATE_ITEM_ERROR, err)
		return c.String(http.StatusBadRequest, errMessage)
	}

	return c.JSONPretty(http.StatusCreated, item, " ")
}

// Finds an item to do by id
// @Summary finds a todo item by id
// @Tags Item
// @Produce json
// @Security JWT
// @Param Authorization header string false "Bearer"
// @Param id path string true "item Id"
// @Success 200 {object} entity.TodoItem
// @Failure 400 string constants.BAD_REQUEST
// @Failure 401 string constants.CANNOT_FETCH_THE_USER_ID
// @Failure 403 string constants.DOES_NOT_BELONG_TO_USER
// @Failure 500 string constants.INTERNAL_SERVER_ERROR
// @Router /item/{id} [get]
func(as ApiService) FindById(c echo.Context) error {
	token,ok := c.Get("user").(*jwt.Token)
	if !ok {
		return c.String(http.StatusUnauthorized, constants.CANNOT_FETCH_THE_USER_ID)
	}

	userId, ok := authentication.GetUserFromToken(token)

	if !ok {
		return c.String(http.StatusUnauthorized, constants.CANNOT_FETCH_THE_USER_ID)
	}

	param:= c.Param("id")

	item, err := as.R.FindItemById(param)

	if err != nil {
		errMessage := fmt.Sprintf(constants.BAD_REQUEST, err)
		return c.String(http.StatusNotFound, errMessage)
	}

	isCurrentUser, err := as.R.IsItemOftheUser(param, userId)

	if err != nil {
		errMessage := fmt.Sprintf(constants.BAD_REQUEST, err)
		return c.String(http.StatusInternalServerError, errMessage)
	}

	if !isCurrentUser {
		return c.String(http.StatusForbidden, constants.DOES_NOT_BELONG_TO_USER)
	}

	return c.JSONPretty(http.StatusOK, item, " ")
}

// Find all items to do by user_id
// @Summary find all todo items by user_id
// @Tags Item
// @Accept json
// @Produce json
// @Security JWT
// @Param Authorization header string false "Bearer"
// @Param dto body entity.GetListDto true "Item List Criteria"
// @Success 200 {array} entity.TodoItem
// @Failure 400 string constants.BAD_REQUEST
// @Failure 401 string constants.CANNOT_FETCH_THE_USER_ID
// @Failure 403 string constants.DOES_NOT_BELONG_TO_USER
// @Failure 500 string constants.INTERNAL_SERVER_ERROR
// @Router /item/list [post]
func (as ApiService) GetAllItems(c echo.Context) error {
	token,ok := c.Get("user").(*jwt.Token)
	if !ok {
		return c.String(http.StatusUnauthorized, constants.CANNOT_FETCH_THE_USER_ID)
	}

	userId, ok := authentication.GetUserFromToken(token)
	if !ok {
		return c.String(http.StatusUnauthorized, constants.CANNOT_FETCH_THE_USER_ID)
	}

	getListBody := new(entity.GetListDto)
	if err := c.Bind(getListBody); err != nil {
		errMessage := fmt.Sprintf(constants.BAD_REQUEST, err)
		return c.String(http.StatusInternalServerError, errMessage)
	}

	
	items, err := as.R.GetTodoListItems(getListBody.Limit, userId)

	if err != nil {
		errMessage := fmt.Sprintf(constants.GET_ALL_ITEMS_ERROR, err)
		return c.String(http.StatusBadRequest, errMessage)
	}

	return c.JSONPretty(http.StatusOK, items, " ")
}

// Updates an item to do by id
// @Summary updates a todo item by id
// @Tags Item
// @Accept json
// @Produce json
// @Security JWT
// @Param Authorization header string false "Bearer"
// @Param id path string true "item Id"
// @Param dto body entity.TodoItemDetailsDto true "Item update criteria"
// @Success 200 {object} entity.TodoItem
// @Failure 400 string constants.BAD_REQUEST
// @Failure 401 string constants.CANNOT_FETCH_THE_USER_ID
// @Failure 403 string constants.DOES_NOT_BELONG_TO_USER
// @Failure 500 string constants.INTERNAL_SERVER_ERROR
// @Router /item/update/{id} [put]
func (as ApiService) UpdateItemById(c echo.Context) error {
	param := c.Param("id")

	token,ok := c.Get("user").(*jwt.Token)
	if !ok {
		return c.String(http.StatusUnauthorized, constants.CANNOT_FETCH_THE_USER_ID)
	}

	userId, ok := authentication.GetUserFromToken(token)

	if !ok {
		return c.String(http.StatusUnauthorized, constants.CANNOT_FETCH_THE_USER_ID)
	}

	isCurrentUser, err := as.R.IsItemOftheUser(param, userId)

	if err != nil {
		errMessage := fmt.Sprintf(constants.BAD_REQUEST, err)
		return c.String(http.StatusInternalServerError, errMessage)
	}

	if !isCurrentUser {
		return c.String(http.StatusForbidden, constants.DOES_NOT_BELONG_TO_USER)
	}

	u := new(entity.TodoItemDetailsDto)
	if err := c.Bind(u); err != nil {
		errMessage := fmt.Sprintf(constants.UPDATE_ITEM_ERROR, err)
		return c.String(http.StatusBadRequest, errMessage)
	}

	validate := validator.New()
	err = validate.Struct(u)

	if err != nil {
		errMessage := fmt.Sprintf(constants.BAD_REQUEST, err)
		return c.String(http.StatusBadRequest, errMessage)
	}

	item, err := as.R.UpdateTodoItem(u, param)
	if err != nil {
		errMessage := fmt.Sprintf(constants.UPDATE_ITEM_ERROR, err)
		return c.String(http.StatusInternalServerError, errMessage)
	}

	return c.JSONPretty(http.StatusOK, item, " ")
}

// Deletes an item to do by id
// @Summary deletes a todo item by id
// @Tags Item
// @Accept json
// @Produce plain
// @Security JWT
// @Param Authorization header string false "Bearer"
// @Param id path string true "item Id"
// @Success 200 string constants.DELETE_ITEM_SUCCESSFULL
// @Failure 400 string constants.BAD_REQUEST
// @Failure 401 string constants.CANNOT_FETCH_THE_USER_ID
// @Failure 403 string constants.DOES_NOT_BELONG_TO_USER
// @Failure 500 string constants.INTERNAL_SERVER_ERROR
// @Router /item/delete/{id} [delete]
func (as ApiService) DeleteById(c echo.Context) error {
	token,ok := c.Get("user").(*jwt.Token)
	if !ok {
		return c.String(http.StatusUnauthorized, constants.CANNOT_FETCH_THE_USER_ID)
	}

	userId, ok := authentication.GetUserFromToken(token)
	if !ok {
		return c.String(http.StatusUnauthorized, constants.CANNOT_FETCH_THE_USER_ID)
	}

	param := c.Param("id")

	isCurrentUser, err := as.R.IsItemOftheUser(param, userId)

	if err != nil {
		errMessage := fmt.Sprintf(constants.BAD_REQUEST, err)
		return c.String(http.StatusInternalServerError, errMessage)
	}

	if !isCurrentUser {
		return c.String(http.StatusForbidden, constants.DOES_NOT_BELONG_TO_USER)
	}

	err = as.R.DeleteTodoItem(param)

	if err != nil {
		errMessage := fmt.Sprintf(constants.DELETE_ITEM_ERROR, err)
		return c.String(http.StatusInternalServerError, errMessage)
	}

	return c.String(http.StatusOK, constants.DELETE_ITEM_SUCCESSFULL)
}

// Sets the item status to complete by id
// @Summary sets the item status to complete by id
// @Tags Item
// @Accept json
// @Produce json
// @Security JWT
// @Param Authorization header string false "Bearer"
// @Param id path string true "item Id"
// @Success 200 {object} entity.TodoItem
// @Failure 400 string constants.BAD_REQUEST
// @Failure 401 string constants.CANNOT_FETCH_THE_USER_ID
// @Failure 403 string constants.DOES_NOT_BELONG_TO_USER
// @Failure 500 string constants.INTERNAL_SERVER_ERROR
// @Router /item/complete/{id} [patch]
func (as ApiService) UpdateStatustoCompleted(c echo.Context) error {
	param := c.Param("id")
	
	token,ok := c.Get("user").(*jwt.Token)
	if !ok {
		return c.String(http.StatusUnauthorized, constants.CANNOT_FETCH_THE_USER_ID)
	}

	userId, ok := authentication.GetUserFromToken(token)
	if !ok {
		return c.String(http.StatusUnauthorized, constants.CANNOT_FETCH_THE_USER_ID)
	}

	isCurrentUser, err := as.R.IsItemOftheUser(param, userId)

	if err != nil {
		errMessage := fmt.Sprintf(constants.BAD_REQUEST, err)
		return c.String(http.StatusInternalServerError, errMessage)
	}

	if !isCurrentUser {
		return c.String(http.StatusForbidden, constants.DOES_NOT_BELONG_TO_USER)
	}

	item, err := as.R.SetItemStatusToComplete(param)
	if err != nil {
		errMessage := fmt.Sprintf(constants.SET_STATUS_ITEM_ERROR, err)
		return c.String(http.StatusInternalServerError, errMessage)
	}

	return c.JSONPretty(http.StatusOK, item, " ")
}
