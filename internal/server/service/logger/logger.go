// Package logger предоставляет инициализацию и глобальный экземпляр логгера.
package logger

import (
	"go.uber.org/zap"
)

// Logger - глобальный экземпляр логгера.
var Logger zap.SugaredLogger

// Initialize инициализирует логгер с указанным уровнем логирования.
//
// Параметр level определяет уровень логирования (например: "debug", "info", "warn", "error").
// Возвращает ошибку, если уровень логирования невалиден или произошла ошибка инициализации.
func Initialize(level string) error {

	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	cfg := zap.NewProductionConfig()

	cfg.Level = lvl

	zl, err := cfg.Build()
	if err != nil {
		return err
	}

	Logger = *zl.Sugar()

	return nil
}