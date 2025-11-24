package service

import (
	"fmt"
	"net/mail"
	"rest_module/repository"
	"sync"

	log "github.com/sirupsen/logrus"

	"golang.org/x/crypto/bcrypt"

	. "rest_module/model"
)

type UserManager struct {
	m          sync.Mutex                 // мьютекс для синхронизации доступа
	repository *repository.UserRepository // репозиторий пользователей
}

// Конструктор сервиса
func UserManagerNewInstance(repository *repository.UserRepository) *UserManager {
	manager := UserManager{}
	manager.repository = repository
	return &manager
}

// Создание пользователя
func (manager *UserManager) AddUser(Username, Password, Email string) (*User, error) {
	log.Println("Создание пользователя")
	manager.m.Lock()
	defer manager.m.Unlock()

	// Проверяем корректность Email
	err := validEmail(Email)
	if err != nil {
		return nil, fmt.Errorf("Не валидный Email %s", err.Error())
	}

	if len(Password) < 8 {
		return nil, fmt.Errorf("Пароль должен содержать не менее 8 символов")
	}

	manager.repository.Db.BeginTransaction()
	exist, _ := manager.repository.GetUserByName(Username)
	if exist != nil {
		manager.repository.Db.RollbackTransaction()
		return nil, fmt.Errorf("Пользователь с таким логином уже есть")
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(Password), bcrypt.DefaultCost)
	user := User{Username: Username, Email: Email, Password: string(hashedPassword)}
	user.ID, err = manager.repository.InsertUser(&user)
	if err != nil {
		manager.repository.Db.RollbackTransaction()
		return nil, fmt.Errorf("Ошибка добавления пользователя %s", err.Error())
	}
	manager.repository.Db.CommitTransaction()
	return &user, nil
}

// Проверка валидности Email
func validEmail(email string) error {
	_, err := mail.ParseAddress(email)
	return err
}

// Поиск пользователя по идентификатору
func (manager *UserManager) FindUserById(id int64) (*User, error) {
	log.Println("Поиск пользователя по идентификатору")
	manager.m.Lock()
	defer manager.m.Unlock()

	manager.repository.Db.BeginTransaction()
	user, _ := manager.repository.GetUserByID(id)
	if user == nil {
		manager.repository.Db.RollbackTransaction()
		return nil, fmt.Errorf("Пользователь с таким идентификатором не найден")
	}
	manager.repository.Db.CommitTransaction()

	return user, nil
}

// Поиск пользователя по имени
func (manager *UserManager) FindUserByName(Username string) (*User, error) {
	log.Println("Поиск пользователя по имени")
	manager.m.Lock()
	defer manager.m.Unlock()

	manager.repository.Db.BeginTransaction()
	user, _ := manager.repository.GetUserByName(Username)
	if user == nil {
		manager.repository.Db.RollbackTransaction()
		return nil, fmt.Errorf("Пользователь с таким логином не найден")
	}
	manager.repository.Db.CommitTransaction()

	return user, nil
}
