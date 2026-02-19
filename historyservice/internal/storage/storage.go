package storage

import (
	"log"
	"log/slog"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"gitlab.roy9.ru/roy9/backend/core/historyservice/internal/config"
	"gitlab.roy9.ru/roy9/backend/core/historyservice/internal/storage/repo"
	"gitlab.roy9.ru/roy9/backend/core/historyservice/internal/storage/repo/history"
)

type Storage interface {
	History() repo.History
}

type storage struct {
	history repo.History
}

func NewStorage(cfg *config.Config, logger *slog.Logger) Storage {
	dsn := "postgres://" + cfg.DbConfig.DbUser + ":" + cfg.DbConfig.DbPassword + "@" + cfg.DbConfig.DbHost +
		":" + cfg.DbConfig.DbPort + "/" + cfg.DbConfig.DbName + "?sslmode=disable"
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	// Колонки created_at/deleted_at в rent хранятся как TIMESTAMP без таймзоны (Europe/Moscow при записи).
	// Чтобы даты не сдвигались на фронте (особенно «сегодня»), интерпретируем их в той же таймзоне.
	_, err = db.Exec("SET timezone = 'Europe/Moscow'")
	if err != nil {
		log.Fatal("set session timezone: " + err.Error())
	}

	logger.Info("db Connection init")

	return &storage{
		history: history.New(db, logger),
	}

}
func (s storage) History() repo.History {
	return s.history
}
