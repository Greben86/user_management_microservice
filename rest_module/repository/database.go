package repository

import (
	"database/sql"
	"fmt"
	. "rest_module/utils"
	"strconv"

	log "github.com/sirupsen/logrus"

	_ "github.com/lib/pq"
)

type DBManager struct {
	database           *sql.DB
	currentTransaction *sql.Tx
}

// Конструктор БД.
func NewDBManager() *DBManager {
	// Получение параметров из переменных окружения
	host := GetEnv("DB_HOST", "localhost")
	port, err := strconv.Atoi(GetEnv("DB_PORT", "5432"))
	if err != nil {
		panic(err)
	}
	dbname := GetEnv("DB_NAME", "database")
	user := GetEnv("DB_USER", "admin")
	password := GetEnv("DB_PASS", "admin")

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	log.Println("База данных подключена!")

	manager := DBManager{}
	manager.database = db
	manager.currentTransaction = nil
	return &manager
}

// Закрытие соединения
func (manager *DBManager) CloseConnection() {
	manager.database.Close()
}

// Старт транзакции
func (manager *DBManager) BeginTransaction() error {
	tx, err := manager.database.Begin()
	if err != nil {
		return fmt.Errorf("Ошибка открытия транзакции %s", err.Error())
	}

	manager.currentTransaction = tx
	return nil
}

// Подтверждение транзакции
func (manager *DBManager) CommitTransaction() error {
	if manager.currentTransaction == nil {
		return fmt.Errorf("Транзакция не была открыта!")
	}

	manager.currentTransaction.Commit()
	manager.currentTransaction = nil
	return nil
}

// Подтверждение транзакции
func (manager *DBManager) RollbackTransaction() error {
	if manager.currentTransaction == nil {
		return fmt.Errorf("Транзакция не была открыта!")
	}

	manager.currentTransaction.Rollback()
	manager.currentTransaction = nil
	return nil
}
