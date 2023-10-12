package authentication

//GET
const IfEmailExistsQuery = `SELECT EXISTS (SELECT 1 FROM users WHERE email = '%s') AS email_exists`

//POST
const CreateUserQuery = "INSERT INTO users (id, email, password) VALUES ($1, $2, $3)"
const LoginQuery = `SELECT password AS hashPassword, id AS userId FROM users WHERE email = '%s'`