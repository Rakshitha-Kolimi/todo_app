package api

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"todo-project/entity"
)

type Repository interface {
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

type ApiRepository struct {
	DB *sql.DB
}

//Get todo item from a sql query
func getItemFromQuery(row *sql.Row, id string) (todoItem entity.TodoItem, err error) {
	var name, description, priority, user_id string
	var is_completed, is_deleted bool
	var due_date, created_at, updated_at time.Time

	err = row.Scan(&id, &name, &description, &due_date, &priority, &created_at, &updated_at, &is_completed, &is_deleted, &user_id)

	if err != nil {
		return todoItem, err
	}

	details := entity.TodoItemDetailsDto{
		Description: description,
		DueDate:     due_date,
		Priority:    priority,
	}

	item := entity.TodoItemDto{
		Name:    name,
		Details: details,
	}

	todoItem = entity.TodoItem{
		Id:          id,
		Item:        item,
		IsCompleted: is_completed,
		IsDeleted:   is_deleted,
		CreatedAt:   created_at,
		UpdatedAt:   updated_at}

	return todoItem, nil
}

//Checks whether the item belongs to current user
func(a ApiRepository)IsItemOftheUser(id, userId string) (bool, error) {
	var is_current_user bool
	isItemOftheUser := fmt.Sprintf(IfItemBelongToUserQuery, id, userId)

	err := a.DB.QueryRow(isItemOftheUser).Scan(&is_current_user)
	if err != nil {
		return false, err
	}

	return is_current_user, nil
}

// Create a to do list item
func (r ApiRepository) CreateTodoItem(item *entity.TodoItemDto, userId string) (todoItem entity.TodoItem, err error) {
	id := (uuid.New()).String()
	due_date := item.Details.DueDate.Format("2006-01-02T15:04:05Z07:00")
	created_at := time.Now().Format("2006-01-02T15:04:05Z07:00")

	row := r.DB.QueryRow(CreateTodoItemQuery, id, item.Name, item.Details.Description, due_date, item.Details.Priority, created_at, created_at, userId)

	return getItemFromQuery(row, id)
}

// Find a todo item by id
func (r ApiRepository) FindItemById(id string) (todoItem entity.TodoItem, err error) {
	findItem := fmt.Sprintf(FindByIdQuery, id)
	row := r.DB.QueryRow(findItem)

	return getItemFromQuery(row, id)
}

// Get all the todo list item
func (r ApiRepository) GetTodoListItems(limit int, userId string) (todoItemList []entity.TodoItem, err error) {
	getAllItems := fmt.Sprintf(GetAllItemsQuery, userId, limit)

	rows, err := r.DB.Query(getAllItems)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []entity.TodoItem

	for rows.Next() {
		var name, description, priority, user_id, id string
		var is_completed, is_deleted bool
		var due_date, created_at, updated_at time.Time

		err = rows.Scan(&id, &name, &description, &due_date, &priority, &created_at, &updated_at, &is_completed, &is_deleted, &user_id)

		details := entity.TodoItemDetailsDto{
			Description: description,
			DueDate:     due_date,
			Priority:    priority,
		}

		item := entity.TodoItemDto{
			Name:    name,
			Details: details,
		}

		todoItem := entity.TodoItem{
			Id:          id,
			Item:        item,
			IsCompleted: is_completed,
			IsDeleted:   is_deleted,
			CreatedAt:   created_at,
			UpdatedAt:   updated_at}

		items = append(items, todoItem)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

// Update an todo list item
func (r ApiRepository) UpdateTodoItem(item *entity.TodoItemDetailsDto, id string) (todoItem entity.TodoItem, err error) {
	updated_at := time.Now().Format("2006-01-02T15:04:05Z07:00")
	row := r.DB.QueryRow(UpdateItemQuery, item.Description, item.DueDate, item.Priority, updated_at, id)

	return getItemFromQuery(row, id)
}

// Delete an todo Item
func (r ApiRepository) DeleteTodoItem(id string) error {
	updated_at := time.Now().Format("2006-01-02T15:04:05Z07:00")

	_, err := r.DB.Exec(DeleteItemQuery, updated_at, id)

	if err != nil {
		return err
	}

	return nil
}

// Update status to Complete by Id
func (r ApiRepository) SetItemStatusToComplete(id string) (todoItem entity.TodoItem, err error) {
	updated_at := time.Now().Format("2006-01-02T15:04:05Z07:00")
	row := r.DB.QueryRow(SetItemStatusAsCompletedQuery, updated_at, id)

	return getItemFromQuery(row, id)
}
