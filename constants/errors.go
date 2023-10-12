package constants

const (
	BAD_REQUEST                  = `Bad request: %v`
	CANNOT_CHECK_IF_EMAIL_EXISTS = `Email validation failed: %v`
	CANNOT_CREATE_TABLE_ERROR    = `Cannot create the table`
	CANNOT_FETCH_THE_USER_ID     = `Cannot find the user id`
	CANNOT_PROCESS_THE_REQUEST   = `Cannot process the request`
	CREATE_ITEM_ERROR            = `Cannot create the todo item: %v`
	DATABASE_CONNECTION_ERROR    = `Cannot connect to data base: %v`
	DELETE_ITEM_ERROR            = `Cannot delete the todo item: %v`
	DOES_NOT_BELONG_TO_USER      = `Item does not belong to the current user`
	EMAIL_ADDRESS_ALREADY_EXISTS = `User with the email id already exists`
	EMAIL_NOT_REGISTERED         = `Email id not registered`
	FIND_ITEM_BY_ITEM_ERROR      = `Cannot find the item: %v`
	GET_ALL_ITEMS_ERROR          = `Cannot fetch todo item list: %v`
	INTERNAL_SERVER_ERROR        = `Internal server error: %v`
	INVALID_PASSWORD             = `Invalid password`
	SET_STATUS_ITEM_ERROR        = `Cannot set staus to complete: %v`
	UPDATE_ITEM_ERROR            = `Cannot update the todo item: %v`
	INVALID_USERNAME_OR_USER_ID  = `invalid username or user_id`
)
