package rest

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/time/rate"

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

type uploadRequest struct {
	Bucket      string `json:"bucket,omitempty"`
	ObjectName  string `json:"object_name"`
	Content     string `json:"content"`
	ContentType string `json:"content_type,omitempty"`
}

type presignRequest struct {
	Bucket        string `json:"bucket,omitempty"`
	ObjectName    string `json:"object_name"`
	ExpirySeconds int    `json:"expiry_seconds,omitempty"`
}

// API приложения.
type API struct {
	r               *mux.Router  // маршрутизатор запросов
	userManager     *UserManager // сервис пользователей
	integration     *IntegrationService
	totalRequests   *prometheus.CounterVec   // счетчик запросов
	requestDuration *prometheus.HistogramVec // метрика длительности запросов
	limiter         *rate.Limiter
}

// Конструктор API.
func ApiNewInstance(userManager *UserManager, integration *IntegrationService) *API {
	api := API{}
	api.userManager = userManager
	api.integration = integration
	api.r = mux.NewRouter()
	api.endpoints()
	api.totalRequests = prometheus.NewCounterVec( // Consistent имя
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint"},
	)
	api.requestDuration = prometheus.NewHistogramVec( // Добавлен для latency
		prometheus.HistogramOpts{
			Name: "http_request_duration_seconds",
			Help: "Request duration",
		},
		[]string{"method", "endpoint"},
	)
	prometheus.MustRegister(api.totalRequests)
	prometheus.MustRegister(api.requestDuration)
	api.limiter = rate.NewLimiter(rate.Limit(1000), 5000) // 1000 req/s + burst 5000 для стабильности
	return &api
}

// Endpoint для проверки работы сервиса
func (api *API) healthHandler(w http.ResponseWriter, r *http.Request) {
	response := ResponseHealth{Status: "UP"}
	json, _ := json.Marshal(&response)
	w.Write(json)
}

func (api *API) metricsMiddleware(next http.Handler) http.Handler { // Для Gorilla Mux (http.Handler)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now() // Таймер для latency
		go api.totalRequests.WithLabelValues(r.Method, r.URL.Path).Inc()
		next.ServeHTTP(w, r)
		// Фиксируем latency
		go api.requestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(time.Since(start).Seconds())
	})
}

func (api *API) rateLimitMiddleware(next http.Handler) http.Handler { // Для Gorilla Mux (http.Handler)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !api.limiter.Allow() {
			go http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Регистрация методов API в маршрутизаторе запросов.
func (api *API) endpoints() {
	// Protected routes
	router := api.Router().PathPrefix("/").Subrouter()
	router.Use(api.metricsMiddleware)
	router.Use(api.rateLimitMiddleware)

	// Public routes
	router.HandleFunc("/health", api.healthHandler).Methods(http.MethodGet)
	router.Handle("/prometheus", promhttp.Handler()).Methods(http.MethodGet)

	router.HandleFunc("/api/users", api.UserListHandler).Methods(http.MethodGet)
	router.HandleFunc("/api/users/{id:[0-9]+}", api.UserInfoHandler).Methods(http.MethodGet)
	router.HandleFunc("/api/users", api.RegisterUserHandler).Methods(http.MethodPost)
	router.HandleFunc("/api/users/{id:[0-9]+}", api.UserUpdateHandler).Methods(http.MethodPut)
	router.HandleFunc("/api/users/{id:[0-9]+}", api.UserDeleteHandler).Methods(http.MethodDelete)

	router.HandleFunc("/storage/objects", api.UploadObject).Methods(http.MethodPost)
	router.HandleFunc("/storage/presign", api.GetPresignedURL).Methods(http.MethodPost)
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
		go http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Выводим тело запроса в ответ
	request := RequestSignUp{}
	err = json.Unmarshal(body, &request)

	// Проверяем наличие ошибок
	if err != nil {
		go http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := api.userManager.AddUser(request.Username, request.Password, request.Email)
	// Проверяем наличие ошибок
	if err != nil {
		go http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response, _ := json.Marshal(&user)
	w.Write(response)
}

// Endpoint списка счетов пользователя
func (api *API) UserListHandler(w http.ResponseWriter, r *http.Request) {
	users, err := api.userManager.FindAllUsers()
	if err != nil {
		go http.Error(w, err.Error(), http.StatusBadRequest)
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
		go http.Error(w, err.Error(), http.StatusBadRequest)
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
		go http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Читаем тело запроса с помощью io.ReadAll
	body, err := io.ReadAll(r.Body)

	// Закрываем тело запроса
	defer r.Body.Close()

	// Проверяем наличие ошибок
	if err != nil {
		go http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Выводим тело запроса в ответ
	request := RequestSignUp{}
	err = json.Unmarshal(body, &request)

	// Проверяем наличие ошибок
	if err != nil {
		go http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err = api.userManager.UpdateUser(int64(id), request.Username, request.Password, request.Email)
	// Проверяем наличие ошибок
	if err != nil {
		go http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response, _ := json.Marshal(&user)
	w.Write(response)
}

// Endpoint удаления пользователя
func (api *API) UserDeleteHandler(w http.ResponseWriter, r *http.Request) {
	// api.totalRequests.WithLabelValues("delete_user_label").Inc()
	id, _ := strconv.Atoi(r.Context().Value("id").(string))
	err := api.userManager.DeleteUserById(int64(id))
	if err != nil {
		go http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Write(nil)
}

func (api *API) UploadObject(w http.ResponseWriter, r *http.Request) {
	var req uploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		go log.Println("UploadObject", err)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	info, err := api.integration.UploadObject(ctx, req.Bucket, req.ObjectName, []byte(req.Content), req.ContentType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		go log.Println("UploadObject", err)
		return
	}
	go log.Println("UPLOAD_OBJECT", 0)
	response := map[string]interface{}{
		"bucket":      info.Bucket,
		"object_name": req.ObjectName,
		"etag":        info.ETag,
		"size":        info.Size,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(response)
}

func (api *API) GetPresignedURL(w http.ResponseWriter, r *http.Request) {
	var req presignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		go log.Println("GetPresignedURL", err)
		return
	}
	expiry := time.Duration(req.ExpirySeconds) * time.Second
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	url, err := api.integration.PresignedURL(ctx, req.Bucket, req.ObjectName, expiry)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		go log.Println("GetPresignedURL", err)
		return
	}
	go log.Println("PRESIGN_OBJECT", 0)
	response := map[string]interface{}{
		"url":            url,
		"expiry_seconds": req.ExpirySeconds,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}
