package api

const IfItemBelongToUserQuery = `SELECT EXISTS (SELECT 1 FROM todo_items WHERE id='%s' AND user_id='%s') AS is_current_user`

//POST
const CreateTodoItemQuery = `INSERT INTO todo_items (id,name,description,due_date,priority,created_at,updated_at,user_id) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING *;`

//GET
const FindByIdQuery = `SELECT * FROM todo_items WHERE id='%s'`
const GetAllItemsQuery = `SELECT * FROM todo_items WHERE is_deleted = false AND user_id='%s' LIMIT %d`

//UPDATE
const UpdateItemQuery = `UPDATE todo_items SET description=$1, due_date=$2, priority=$3, updated_at=$4 WHERE id=$5 RETURNING *;`

//DELETE 
const DeleteItemQuery = `UPDATE todo_items SET is_deleted = 'true', updated_at=$1 WHERE id=$2;`

//PATCH
const SetItemStatusAsCompletedQuery = `UPDATE todo_items SET is_completed = 'true', updated_at=$1 WHERE id=$2 RETURNING *;`