package main

import (
	"net/http"

	log "github.com/sirupsen/logrus"

	. "rest_module/repository"
	. "rest_module/rest"
	. "rest_module/service"
)

func main() {
	// Инициализация соединения с БД
	var dbManager = NewDBManager()
	defer dbManager.CloseConnection()
	err := dbManager.InitDB()
	if err != nil {
		log.Fatal(err)
	}

	// Создание объектов API пользователя
	var userRepository = InitUserRepository(dbManager)
	var userManager = UserManagerNewInstance(userRepository)
	var usersController = UsersControllerNewInstance(userManager)

	// Главный контроллер приложения
	api := ApiNewInstance(usersController)
	// Запуск сетевой службы и HTTP-сервера
	// на всех локальных IP-адресах на порту 8080.
	err = http.ListenAndServe(":8080", api.Router())
	if err != nil {
		log.Fatal(err)
	}
}
