package bot

import (
	"fmt"
	"immich-telegramm-uploader-bot/uploader/immich"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/joho/godotenv/autoload"
	tele "gopkg.in/telebot.v3"
)

func Handle() {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("Не задан TELEGRAM_BOT_TOKEN")
	}
	immichToken := os.Getenv("IMMICH_TOKEN")
	if immichToken == "" {
		log.Fatal("Не задан IMMICH_TOKEN")
	}
	immichServer := os.Getenv("IMMICH_SERVER")
	if immichServer == "" {
		log.Fatal("Не задан IMMICH_SERVER")
	}

	settings := tele.Settings{
		Token:  botToken,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := tele.NewBot(settings)
	if err != nil {
		log.Fatalf("Ошибка создания бота: %v", err)
	}

	i := immich.Immich{Server: immichServer, Token: immichToken}

	// --- 1. Обработка фото ---
	bot.Handle(tele.OnPhoto, func(c tele.Context) error {
		photo := c.Message().Photo
		if photo == nil {
			return c.Send("Нет фото.")
		}

		file, err := bot.FileByID(photo.FileID)
		if err != nil {
			return c.Send("Не удалось загрузить фото.")
		}

		reader, err := bot.File(&file)
		if err != nil {
			return c.Send("Ошибка чтения файла.")
		}

		filename := fmt.Sprintf("photo_%s.jpg", photo.FileID)

		_, err = i.Upload(reader, filename, nil)
		if err != nil {
			return c.Send(fmt.Sprintf("Ошибка при отправке на API: %v", err))
		}

		return nil
	})

	// --- 2. Обработка видео ---
	bot.Handle(tele.OnVideo, func(c tele.Context) error {
		video := c.Message().Video
		if video == nil {
			return c.Send("Нет видео.")
		}

		file, err := bot.FileByID(video.FileID)
		if err != nil {
			return c.Send("Не удалось получить видео.")
		}

		reader, err := bot.File(&file)
		if err != nil {
			return c.Send("Ошибка чтения видео.")
		}

		filename := fmt.Sprintf("video_%s.mp4", file.FileID)

		_, err = i.Upload(reader, filename, nil)
		if err != nil {
			return c.Send(fmt.Sprintf("Ошибка при отправке на API: %v", err))
		}

		return nil
	})

	// --- 3. Обработка документов (фото, видео, другие файлы) ---
	bot.Handle(tele.OnDocument, func(c tele.Context) error {
		doc := c.Message().Document
		if doc == nil {
			return c.Send("Нет документа.")
		}

		// Если это изображение — обрабатываем как фото
		if strings.HasPrefix(doc.MIME, "image/") {
			file, err := bot.FileByID(doc.FileID)
			if err != nil {
				return c.Send("Не удалось получить файл.")
			}

			reader, err := bot.File(&file)
			if err != nil {
				return c.Send("Ошибка чтения файла.")
			}

			filename := doc.FileName
			if filename == "" {
				ext := ".jpg"
				if strings.HasPrefix(doc.MIME, "image/png") {
					ext = ".png"
				} else if strings.HasPrefix(doc.MIME, "image/webp") {
					ext = ".webp"
				} else if strings.HasPrefix(doc.MIME, "image/heic") || strings.HasPrefix(doc.MIME, "image/heif") {
					ext = ".heic"
				}
				filename = fmt.Sprintf("image_%s%s", file.FileID, ext)
			}

			_, err = i.Upload(reader, filename, nil)
			if err != nil {
				return c.Send(fmt.Sprintf("Ошибка при отправке на API: %v", err))
			}

			return nil
		}

		// Если это видео — обрабатываем как видео
		if strings.HasPrefix(doc.MIME, "video/") {
			file, err := bot.FileByID(doc.FileID)
			if err != nil {
				return c.Send("Не удалось получить видео.")
			}

			reader, err := bot.File(&file)
			if err != nil {
				return c.Send("Ошибка чтения видео.")
			}

			filename := doc.FileName
			if filename == "" {
				ext := ".mp4"
				if strings.HasPrefix(doc.MIME, "video/quicktime") {
					ext = ".mov"
				} else if strings.HasPrefix(doc.MIME, "video/x-msvideo") {
					ext = ".avi"
				}
				filename = fmt.Sprintf("video_%s%s", file.FileID, ext)
			}

			_, err = i.Upload(reader, filename, nil)
			if err != nil {
				return c.Send(fmt.Sprintf("Ошибка при отправке на API: %v", err))
			}

			return nil
		}

		// Остальные типы — игнорируем
		return nil
	})

	log.Println("Бот запущен...")
	bot.Start()
}
