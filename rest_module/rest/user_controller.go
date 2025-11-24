package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	. "rest_module/service"
)

type RequestSignUp struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type ResponseAuth struct {
	JwtToken string `json:"token"`
}

// API пользователей
type UsersController struct {
	userManager *UserManager // сервис пользователей
}

// Конструктор API пользователей
func UsersControllerNewInstance(userManager *UserManager) *UsersController {
	api := UsersController{}
	api.userManager = userManager
	return &api
}

// Endpoint для регистрации
func (api *UsersController) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// Читаем тело запроса с помощью io.ReadAll
	body, err := io.ReadAll(r.Body)

	// Закрываем тело запроса
	defer r.Body.Close()

	// Проверяем наличие ошибок
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Выводим тело запроса в ответ
	request := RequestSignUp{}
	err = json.Unmarshal(body, &request)

	// Проверяем наличие ошибок
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := api.userManager.AddUser(request.Username, request.Password, request.Email)
	// Проверяем наличие ошибок
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, _ := GenerateJWTToken(fmt.Sprint(user.ID))
	responseDTO := ResponseAuth{JwtToken: token}
	response, _ := json.Marshal(&responseDTO)
	w.Write(response)
}

// Endpoint для аутентификации
func (api *UsersController) LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Читаем тело запроса с помощью io.ReadAll
	body, err := io.ReadAll(r.Body)

	// Закрываем тело запроса
	defer r.Body.Close()

	// Проверяем наличие ошибок
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Выводим тело запроса в ответ
	request := RequestSignUp{}
	err = json.Unmarshal([]byte(body), &request)

	// Проверяем наличие ошибок
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := api.userManager.FindUserByName(request.Username)
	// Проверяем наличие ошибок
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = CheckPasswordForUser(user, request.Password)
	// Проверяем наличие ошибок
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, _ := GenerateJWTToken(fmt.Sprint(user.ID))
	responseDTO := ResponseAuth{JwtToken: token}
	response, _ := json.Marshal(&responseDTO)
	w.Write(response)
}

// Endpoint информации о пользователе
func (api *UsersController) UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.Context().Value("id").(string))
	user, err := api.userManager.FindUserById(int64(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sender, _ := InitMailSender()
	sender.SendEmailMessage(user.Email, 0.)

	json, _ := json.Marshal(&user)
	w.Write(json)
}
