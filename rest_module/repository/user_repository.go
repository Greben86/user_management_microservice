package repository

import (
	"database/sql"
	. "rest_module/model"
)

type UserRepository struct {
	Db *DBManager // база данных
}

func InitUserRepository(db *DBManager) *UserRepository {
	repo := UserRepository{}
	repo.Db = db
	return &repo
}

func (repo *UserRepository) Database() *sql.DB {
	if repo.Db == nil {
		panic("База данных не подключена!")
	}

	return repo.Db.database
}

// Сохранение нового пользователя в БД
func (repo *UserRepository) InsertUser(user *User) (int64, error) {
	insertStmt := `insert into "users" ("username", "password", "email") values($1, $2, $3) returning "id"`

	var id int64 = 0
	err := repo.Database().QueryRow(insertStmt, user.Username, user.Password, user.Email).Scan(&id)
	if err != nil {
		return -1, err
	}

	return id, nil
}

// Обновление пользователя
func (repo *UserRepository) UpdateUser(user *User) error {
	insertStmt := `update "users" set "username"=$1, "password"=$2, "email"=$3 where "id" = $1`

	_, err := repo.Database().Exec(insertStmt, user.Username, user.Password, user.Email, user.ID)
	if err != nil {
		return err
	}

	return nil
}

// Поиск пользователя по идентификатору
func (repo *UserRepository) GetUserByID(id int64) (*User, error) {
	selectStmt := `select "id", "username", "password", "email" from "users" where "id" = $1`
	rows, err := repo.Database().Query(selectStmt, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		var id int64
		var username string
		var password string
		var email string

		err = rows.Scan(&id, &username, &password, &email)
		if err != nil {
			return nil, err
		}

		return &User{
			ID:       id,
			Username: username,
			Password: password,
			Email:    email,
		}, nil
	}

	return nil, nil
}

// Поиск пользователя по имени
func (repo *UserRepository) GetUserByName(name string) (*User, error) {
	selectStmt := `select "id", "username", "password", "email" from "users" where "username" ~ $1`
	rows, err := repo.Database().Query(selectStmt, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		var id int64
		var username string
		var password string
		var email string

		err = rows.Scan(&id, &username, &password, &email)
		if err != nil {
			return nil, err
		}

		return &User{
			ID:       id,
			Username: username,
			Password: password,
			Email:    email,
		}, nil
	}

	return nil, nil
}
