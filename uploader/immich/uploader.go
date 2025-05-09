package immich

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

type Immich struct {
	Server string
	Token  string
}

type ApiResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func (im *Immich) Upload(fileReader io.Reader, filename string, tags []string) (string, error) {
	// Читаем файл в буфер
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, fileReader); err != nil {
		return "", fmt.Errorf("ошибка чтения файла: %v", err)
	}

	// Вычисляем SHA-1 для deviceAssetId и x-immich-checksum
	hasher := sha1.New()
	hasher.Write(buf.Bytes())
	checksum := hex.EncodeToString(hasher.Sum(nil))

	// Создаем multipart/form-data
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Добавляем файл
	filePart, err := writer.CreateFormFile("assetData", filename)
	if err != nil {
		return "", fmt.Errorf("ошибка создания поля asset: %v", err)
	}
	if _, err := io.Copy(filePart, bytes.NewReader(buf.Bytes())); err != nil {
		return "", fmt.Errorf("ошибка копирования данных файла: %v", err)
	}

	// Добавляем текстовые поля
	if err := writer.WriteField("deviceAssetId", checksum); err != nil {
		return "", fmt.Errorf("ошибка записи deviceAssetId: %v", err)
	}
	if err := writer.WriteField("deviceId", "telegram"); err != nil {
		return "", fmt.Errorf("ошибка записи deviceId: %v", err)
	}

	now := time.Now()
	if err := writer.WriteField("fileCreatedAt", now.Format(time.RFC3339)); err != nil {
		return "", fmt.Errorf("ошибка записи fileCreatedAt: %v", err)
	}
	if err := writer.WriteField("fileModifiedAt", now.Format(time.RFC3339)); err != nil {
		return "", fmt.Errorf("ошибка записи fileModifiedAt: %v", err)
	}

	// Добавляем теги (если есть)
	if len(tags) > 0 {
		tagJSON, _ := json.Marshal(tags)
		if err := writer.WriteField("tags", string(tagJSON)); err != nil {
			return "", fmt.Errorf("ошибка записи tags: %v", err)
		}
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("ошибка закрытия multipart writer: %v", err)
	}

	// Формируем запрос
	req, err := http.NewRequest("POST", im.Server+"/api/assets", &body)
	if err != nil {
		return "", fmt.Errorf("ошибка создания запроса: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-immich-checksum", checksum)
	req.Header.Set("x-api-key", im.Token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ошибка выполнения запроса: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("не удалось прочитать ответ: %v", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("ошибка API: статус %d, тело: %s", resp.StatusCode, respBody)
	}

	// Парсим JSON
	var apiResp struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}

	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return "", fmt.Errorf("не удалось распарсить ответ от API: %v", err)
	}

	return apiResp.ID, nil
}
