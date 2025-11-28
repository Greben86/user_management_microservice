package service

import (
	"context"
	"fmt"
	"net/mail"
	"rest_module/repository"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"golang.org/x/crypto/bcrypt"

	. "rest_module/model"
)

type UserManager struct {
	m           sync.Mutex                 // мьютекс для синхронизации доступа
	repository  *repository.UserRepository // репозиторий пользователей
	integration *IntegrationService
}

// Конструктор сервиса
func UserManagerNewInstance(repository *repository.UserRepository, integration *IntegrationService) *UserManager {
	manager := UserManager{}
	manager.repository = repository
	manager.integration = integration
	return &manager
}

// Создание пользователя
func (manager *UserManager) AddUser(Username, Password, Email string) (*User, error) {
	go log.Println("Создание пользователя")
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
	manager.exportUserSnapshot(&user)
	manager.repository.Db.CommitTransaction()
	return &user, nil
}

// Обновление пользователя
func (manager *UserManager) UpdateUser(id int64, Username, Password, Email string) (*User, error) {
	go log.Println("Обновление пользователя")
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
	err = manager.repository.UpdateUser(id, &user, string(hashedPassword))
	if err != nil {
		manager.repository.Db.RollbackTransaction()
		return nil, fmt.Errorf("Ошибка обновления пользователя %s", err.Error())
	}
	manager.exportUserSnapshot(&user)
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
	go log.Println("Поиск пользователя по идентификатору")
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
	go log.Println("Поиск пользователя по имени")
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

// Поиск пользователей
func (manager *UserManager) FindAllUsers() (*[]User, error) {
	go log.Println("Чтение пользователей")
	manager.m.Lock()
	defer manager.m.Unlock()

	manager.repository.Db.BeginTransaction()
	users, _ := manager.repository.GetAllUsers()
	if users == nil {
		manager.repository.Db.RollbackTransaction()
		return nil, fmt.Errorf("Пользователи не найдены")
	}
	manager.repository.Db.CommitTransaction()

	return users, nil
}

// Удаление пользователя
func (manager *UserManager) DeleteUserById(id int64) error {
	go log.Println("Удаление пользователя")
	manager.m.Lock()
	defer manager.m.Unlock()

	manager.repository.Db.BeginTransaction()
	err := manager.repository.DeleteUserById(id)
	if err != nil {
		manager.repository.Db.RollbackTransaction()
		return fmt.Errorf("Ошибка удаления пользователя %s", err.Error())
	}
	manager.repository.Db.CommitTransaction()

	return nil
}

func (manager *UserManager) exportUserSnapshot(user *User) {
	if manager.integration == nil || user == nil {
		return
	}
	userCopy := *user
	go func(u User) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := manager.integration.ExportUserSnapshot(ctx, manager.integration.GetBucket(), &u); err != nil {
			go log.Println("ExportUserSnapshot", err)
		}
	}(userCopy)
}
