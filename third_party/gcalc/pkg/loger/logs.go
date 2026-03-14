package loger

import (
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
)

// Logger Структура лога
type Logger struct {
	logger *log.Logger
	file   *os.File
}

// Start Функция создания лога
func Start(name string) (*Logger, error) {
	execPath, err := os.Executable()
	if err != nil {
		return nil, err
	}
	execDir := filepath.Dir(execPath)

	logDir := filepath.Join(execDir, "logs")
	logFile := filepath.Join(logDir, name+".log")

	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, err
	}

	multiWriter := io.MultiWriter(os.Stdout, file)
	logger := log.New(multiWriter, "["+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	logger.Println(logFile)
	return &Logger{logger: logger, file: file}, nil
}

// NewDiscard creates a logger that drops all output.
func NewDiscard() *Logger {
	return &Logger{
		logger: log.New(io.Discard, "", 0),
		file:   nil,
	}
}

// Add Println Метод логирования (обёртка)
func (l *Logger) Add(v ...interface{}) {
	l.logger.Println(v...)
}

// AddF Println Метод логирования (обёртка)
func (l *Logger) AddF(v ...interface{}) {
	l.logger.Fatalln(v...)
}

// CheckError обертка обработки ошибки
func (l *Logger) CheckError(err error) bool {
	if err != nil {
		l.AddF("Ошибка: ", err) // Пишем основную ошибку

		var unwrapped = err
		for errors.Unwrap(unwrapped) != nil {
			unwrapped = errors.Unwrap(unwrapped)
			l.AddF("Вложенная ошибка: ", unwrapped) // Пишем вложенные ошибки
		}
		return true
	}
	return false
}

// Stop Метод завершения лога
func (l *Logger) Stop() {
	l.logger.Println("Лог завершён.") // Финальное сообщение
	if l.file == nil {
		return
	}
	if err := l.file.Close(); err != nil {
		log.Printf("Ошибка при закрытии лога: %v", err)
	}
}
