package entity

import "time"

type TodoItem struct {
	Id          string `json:"id" validate:"required" example:"3604fa26-5ee8-428f-a6dd-c742455e8148"`
	Item        TodoItemDto `json:"item" validate:"required"`
	IsCompleted bool   `json:"is_completed" validate:"required"`
	IsDeleted   bool   `json:"is_deleted" validate:"required"`
	CreatedAt   time.Time `json:"created_at" validate:"required"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type TodoItemDto struct {
	Name        string `json:"name" validate:"required,min=3" example:"Todo list 1"`
    Details    TodoItemDetailsDto `json:"details" validate:"required"`
}

type TodoItemDetailsDto struct {
	Description string `json:"description" validate:"required" example:"This is description for the todo list item."`
	DueDate     time.Time `json:"due_date" validate:"required" example:"2023-05-22T09:38:24.405027Z"` 
	Priority    string `json:"priority" validate:"required,oneof=HIGH LOW MEDIUM" example:"HIGH"`
}

type GetListDto struct {
	Limit int `json:"limit" validate:"required" example:"5"`
}