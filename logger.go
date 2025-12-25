package main

import (
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"log"
	"os"
	"time"
)

func initLogger() {
	// базовое имя лога (на него будет ссылка)
	logPath := "logs/chat.log"

	// writer с ротацией по времени
	writer, err := rotatelogs.New(
		"logs/%Y-%m-%d-chat.log",         // итоговые файлы: chat.log.2025-02-20
		rotatelogs.WithLinkName(logPath), // симлинк "chat.log" -> текущий лог
		//rotatelogs.WithMaxAge(7*24*time.Hour),     // храним 7 дней
		rotatelogs.WithMaxAge(0),                  // не удалять никогда
		rotatelogs.WithRotationTime(24*time.Hour), // ротация раз в сутки
	)
	if err != nil {
		log.Fatalf("failed to init log file rotator: %v", err)
	}

	log.SetOutput(writer)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func logMessage(msg Message) {
	// Открываем файл в режиме добавления
	f, err := os.OpenFile("chat.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("log file error:", err)
		return
	}
	defer f.Close()

	// Формат лога
	line := msg.Time + " | " + msg.User + " | " + msg.Text + " | " + msg.ClientID + "\n"

	if _, err := f.WriteString(line); err != nil {
		log.Println("write log error:", err)
	}
}
