package repository

import (
	"database/sql"
	"fmt"

	log "github.com/sirupsen/logrus"

	// "strconv"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	dbname   = "database"
	user     = "admin"
	password = "admin"
)

type DBManager struct {
	database           *sql.DB
	currentTransaction *sql.Tx
}

// Конструктор БД.
func NewDBManager() *DBManager {
	// Получение параметров из переменных окружения
	// host := os.Getenv("DB_HOST")
	// port, err := strconv.Atoi(os.Getenv("DB_PORT"))
	// if err != nil {
	// 	panic(err)
	// }
	// dbname := os.Getenv("DB_NAME")
	// user := os.Getenv("DB_USER")
	// password := os.Getenv("DB_PASS")

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
