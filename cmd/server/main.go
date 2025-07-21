package main

import (
	"fmt"
	"go.uber.org/zap"

	"AccountService/internal/app/account"
	"AccountService/internal/config"
	"AccountService/internal/db"
	"AccountService/internal/token"
	"AccountService/internal/transport/rest"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	sugar := logger.Sugar()

	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		sugar.Fatalf("ошибка загрузки конфига: %v", err)
	}

	// Подключение к БД
	dbCfg := db.DBConfig{
		Host:     cfg.DB.Host,
		Port:     cfg.DB.Port,
		User:     cfg.DB.User,
		Password: cfg.DB.Password,
		Name:     cfg.DB.Name,
	}
	pg, err := db.NewPostgres(dbCfg)
	if err != nil {
		sugar.Fatalf("DB error: %v", err)
	}
	defer pg.Close()

	// Применение миграций
	migrationURL := "file://migrations"
	dsn := "postgres://" + cfg.DB.User + ":" + cfg.DB.Password + "@" +
		cfg.DB.Host + ":" + fmt.Sprint(cfg.DB.Port) + "/" + cfg.DB.Name + "?sslmode=disable"

	migrator, err := migrate.New(migrationURL, dsn)
	if err != nil {
		sugar.Fatalf("ошибка инициализации мигратора: %v", err)
	}
	if err := migrator.Up(); err != nil && err != migrate.ErrNoChange {
		sugar.Fatalf("ошибка применения миграций: %v", err)
	} else {
		sugar.Info("✅ Миграции успешно применены (или нет изменений)")
	}

	// Инициализация JWT-менеджера
	tokens := token.New(cfg.JWT.Secret, cfg.JWT.AccessTokenTTL, cfg.JWT.RefreshTokenTTL)

	// handler и сервисы
	accountService := account.NewService(pg, tokens)
	accountHandler := account.NewHandler(accountService)

	// роутинги
	router := rest.SetupRouter(accountHandler)

	sugar.Infof("🚀 Server running on port %s", cfg.Server.Port)
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		sugar.Fatalf("server failed: %v", err)
	}
}
