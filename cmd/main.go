package main

import (
	"immich-telegramm-uploader-bot/bot"
	"immich-telegramm-uploader-bot/metrics"
	"log"
)

func main() {
	// Запускаем метрики в отдельной горутине
	go func() {
		addr := ":8080" // порт можно вынести в .env
		log.Printf("Запуск Prometheus метрик на %s", addr)
		if err := metrics.StartMetricsServer(addr); err != nil {
			log.Fatalf("Ошибка запуска метрик: %v", err)
		}
	}()
	bot.Handle()
}
