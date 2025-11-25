package rest

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	. "rest_module/service"
)

type RequestSignUp struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type ResponseHealth struct {
	Status string `json:"status"`
}

// API приложения.
type API struct {
	r           *mux.Router  // маршрутизатор запросов
	userManager *UserManager // сервис пользователей

}

// Конструктор API.
func ApiNewInstance(userManager *UserManager) *API {
	api := API{}
	api.userManager = userManager
	api.r = mux.NewRouter()
	api.endpoints()
	return &api
}

// Endpoint для проверки работы сервиса
func (api *API) healthHandler(w http.ResponseWriter, r *http.Request) {
	response := ResponseHealth{Status: "UP"}
	json, _ := json.Marshal(&response)
	w.Write(json)
}

// Регистрация методов API в маршрутизаторе запросов.
func (api *API) endpoints() {
	// Public routes
	api.Router().HandleFunc("/health", api.healthHandler).Methods(http.MethodGet)

	api.Router().HandleFunc("/api/users", api.UserListHandler).Methods(http.MethodGet)
	api.Router().HandleFunc("/api/users/{id}", api.UserInfoHandler).Methods(http.MethodGet)
	api.Router().HandleFunc("/api/users", api.RegisterUserHandler).Methods(http.MethodPost)
	api.Router().HandleFunc("/api/users/{id}", api.UserUpdateHandler).Methods(http.MethodPut)
	api.Router().HandleFunc("/api/users/{id}", api.UserDeleteHandler).Methods(http.MethodDelete)
}

// Router возвращает маршрутизатор запросов.
func (api *API) Router() *mux.Router {
	return api.r
}

// Endpoint для регистрации
func (api *API) RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
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

	response, _ := json.Marshal(&user)
	w.Write(response)
}

// Endpoint списка счетов пользователя
func (api *API) UserListHandler(w http.ResponseWriter, r *http.Request) {
	users, err := api.userManager.FindAllUsers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response, _ := json.Marshal(&users)
	w.Write(response)
}

// Endpoint информации о пользователе
func (api *API) UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.Context().Value("id").(string))
	user, err := api.userManager.FindUserById(int64(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json, _ := json.Marshal(&user)
	w.Write(json)
}

// Endpoint обновления информации о пользователе
func (api *API) UserUpdateHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.Context().Value("id").(string))
	user, err := api.userManager.FindUserById(int64(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

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

	user, err = api.userManager.UpdateUser(int64(id), request.Username, request.Password, request.Email)
	// Проверяем наличие ошибок
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response, _ := json.Marshal(&user)
	w.Write(response)
}

// Endpoint удаления пользователя
func (api *API) UserDeleteHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.Context().Value("id").(string))
	err := api.userManager.DeleteUserById(int64(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Write(nil)
}
