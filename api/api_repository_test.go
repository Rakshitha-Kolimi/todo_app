package api

import (
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"todo-project/entity"
)

func TestApiRepo(t *testing.T) {
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

	t.Run("Test Get Item from query", func(t *testing.T) {
		rows := mock.NewRows([]string{"id", "name", "description", "due_date", "priority", "created_at", "updated_at", "is_completed", "is_deleted", "user_id"}).
			AddRow("3a35452e-957c-4588-8d40-c88f370067d2", "Todo list item 1", "This is item 1", dueDate, "HIGH", createdTime, updatedAt, true, false, "ed6caeda-1fa9-442e-a41d-dd2b135cea67")

		t.Run("Success", func(t *testing.T) {
			mock.ExpectQuery("SELECT").WillReturnRows(rows)

			row := as.R.DB.QueryRow("SELECT * FROM todo_items WHERE id='3a35452e-957c-4588-8d40-c88f370067d2'")

			got, err := getItemFromQuery(row, "3a35452e-957c-4588-8d40-c88f370067d2")

			if err != nil {
				t.Fatalf("Error getting the item from the query: %v", err)
			}

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

			assert.Equal(t, got, want)
		})

		t.Run("Error", func(t *testing.T) {
			mock.ExpectQuery("SELECT").WillReturnRows(rows)
			row := as.R.DB.QueryRow("SELECT * FROM todo_items WHERE id='3a35452e-957c-4588-8d40-c88f370067d2'")

			_, err := getItemFromQuery(row, "")

			assert.Error(t, err)
		})
	})

	t.Run("Test Create an item", func(t *testing.T) {
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
			WithArgs(sqlmock.AnyArg(), item.Name, item.Details.Description, sqlmock.AnyArg(), item.Details.Priority, sqlmock.AnyArg(), sqlmock.AnyArg(), "user_id").
			WillReturnRows(rows)

		got, err := as.R.CreateTodoItem(item, "user_id")

		if err != nil {
			t.Fatalf("Error getting the item from the query: %v", err)
		}

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

		assert.Equal(t, got, want)
	})

	t.Run("Test find item by id", func(t *testing.T) {
		rows := mock.NewRows([]string{"id", "name", "description", "due_date", "priority", "created_at", "updated_at", "is_completed", "is_deleted", "user_id"}).
			AddRow("3a35452e-957c-4588-8d40-c88f370067d2", "Todo list item 1", "This is item 1", dueDate, "HIGH", createdTime, updatedAt, true, false, "ed6caeda-1fa9-442e-a41d-dd2b135cea67")

		mock.ExpectQuery(`SELECT \* FROM todo_items WHERE id='3a35452e-957c-4588-8d40-c88f370067d2'`).
			WillReturnRows(rows)

		got, err := as.R.FindItemById("3a35452e-957c-4588-8d40-c88f370067d2")

		if err != nil {
			t.Fatalf("Error getting the item from the query: %v", err)
		}

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

		assert.Equal(t, got, want)
	})

	t.Run("Test update todo item", func(t *testing.T) {
		newUpdatedAtString := time.Now().Format("2006-01-02T15:04:05Z07:00")
		newUpdatedDate, _ := time.Parse(newUpdatedAtString, "2006-01-02T15:04:05Z07:00")

		rows := mock.NewRows([]string{"id", "name", "description", "due_date", "priority", "created_at", "updated_at", "is_completed", "is_deleted", "user_id"}).
			AddRow("3a35452e-957c-4588-8d40-c88f370067d2", "Todo list item 1", "This is item 2", dueDate, "HIGH", createdTime, newUpdatedDate, true, false, "ed6caeda-1fa9-442e-a41d-dd2b135cea67")

		newUpdatedQuery := `UPDATE todo_items SET description=\$1, due_date=\$2, priority=\$3, updated_at=\$4 WHERE id=\$5 RETURNING \*;`

		mock.ExpectQuery(newUpdatedQuery).
			WithArgs("This is item 2", dueDate, "HIGH", newUpdatedAtString, "3a35452e-957c-4588-8d40-c88f370067d2").
			WillReturnRows(rows)

		details := entity.TodoItemDetailsDto{
			Description: "This is item 2",
			Priority:    "HIGH",
			DueDate:     dueDate,
		}

		got, err := as.R.UpdateTodoItem(&details, "3a35452e-957c-4588-8d40-c88f370067d2")
		if err != nil {
			t.Fatalf("Error getting the item from the query: %v", err)
		}

		want := entity.TodoItem{
			Id: "3a35452e-957c-4588-8d40-c88f370067d2",
			Item: entity.TodoItemDto{
				Name: "Todo list item 1",
				Details: entity.TodoItemDetailsDto{
					Description: "This is item 2",
					Priority:    "HIGH",
					DueDate:     dueDate,
				},
			},
			IsCompleted: true,
			IsDeleted:   false,
			CreatedAt:   createdTime,
			UpdatedAt:   newUpdatedDate,
		}

		assert.Equal(t, got, want)
	})

	t.Run("Test Delete todo item", func(t *testing.T) {
		newUpdatedAtString := time.Now().Format("2006-01-02T15:04:05Z07:00")
		newDeleteQuery := `UPDATE todo_items SET is_deleted = 'true', updated_at=\$1 WHERE id=\$2;`

		t.Run("SUCCESS", func(t *testing.T) {
			mock.ExpectExec(newDeleteQuery).WithArgs(newUpdatedAtString, "3a35452e-957c-4588-8d40-c88f370067d2").
				WillReturnResult(sqlmock.NewResult(1, 1))

			err := as.R.DeleteTodoItem("3a35452e-957c-4588-8d40-c88f370067d2")

			assert.NoError(t, err)
		})

		t.Run("ERROR", func(t *testing.T) {
			mock.ExpectExec(newDeleteQuery).WithArgs(newUpdatedAtString, "3a35452e-957c-4588-8d40-c88f370067d2").
				WillReturnError(errors.New("internal server error"))

			err := as.R.DeleteTodoItem("3a35452e-957c-4588-8d40-c88f370067d2")

			assert.Error(t, err)
		})
	})

	t.Run("Test Set item status to complete", func(t *testing.T) {
		rows := mock.NewRows([]string{"id", "name", "description", "due_date", "priority", "created_at", "updated_at", "is_completed", "is_deleted", "user_id"}).
			AddRow("3a35452e-957c-4588-8d40-c88f370067d2", "Todo list item 1", "This is item 1", dueDate, "HIGH", createdTime, updatedAt, true, false, "ed6caeda-1fa9-442e-a41d-dd2b135cea67")

		newUpdatedAtString := time.Now().Format("2006-01-02T15:04:05Z07:00")
		newCompleteQuery := `UPDATE todo_items SET is_completed = 'true', updated_at=\$1 WHERE id=\$2 RETURNING \*;`

		mock.ExpectQuery(newCompleteQuery).WithArgs(newUpdatedAtString, "3a35452e-957c-4588-8d40-c88f370067d2").
			WillReturnRows(rows)

		got, err := as.R.SetItemStatusToComplete("3a35452e-957c-4588-8d40-c88f370067d2")
		if err != nil {
			t.Fatalf("Error getting the item from the query: %v", err)
		}

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

		assert.Equal(t, got, want)
	})

	t.Run("Test Get todo list items", func(t *testing.T) {
		rows := mock.NewRows([]string{"id", "name", "description", "due_date", "priority", "created_at", "updated_at", "is_completed", "is_deleted", "user_id"}).
			AddRow("3a35452e-957c-4588-8d40-c88f370067d2", "Todo list item 1", "This is item 1", dueDate, "HIGH", createdTime, updatedAt, true, false, "ed6caeda-1fa9-442e-a41d-dd2b135cea67")

		t.Run("SUCCESS", func(t *testing.T) {
			GetItemsQuery := `SELECT \* FROM todo_items WHERE is_deleted = false AND user_id='ed6caeda-1fa9-442e-a41d-dd2b135cea67' LIMIT 1`
			mock.ExpectQuery(GetItemsQuery).WillReturnRows(rows)

			got, err := as.R.GetTodoListItems(1, "ed6caeda-1fa9-442e-a41d-dd2b135cea67")
			if err != nil {
				t.Fatalf("Error getting the item from the query: %v", err)
			}

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

			assert.Len(t, got, 1)
			assert.Equal(t, item, got[0])

			assert.NoError(t, mock.ExpectationsWereMet())
		})

		t.Run("FAILED", func(t *testing.T) {
			GetItemsQuery := `SELECT \* FROM todo_items WHERE is_deleted = false AND user_id='ed6caeda-1fa9-442e-a41d-dd2b135cea67' LIMIT 1`
			mock.ExpectQuery(GetItemsQuery).WillReturnError(errors.ErrUnsupported)

			_, err := as.R.GetTodoListItems(1, "ed6caeda-1fa9-442e-a41d-dd2b135cea67")
			assert.Error(t, err)
		})
	})
}
