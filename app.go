package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	_ "github.com/lib/pq"
)

func main() {
	// Укажите учетные данные AWS
	accessKey := os.Getenv("YOUR_AWS_ACCESS_KEY")
	secretKey := os.Getenv("YOUR_AWS_SECRET_KEY")

	bucketName := os.Getenv("YOUR_BUCKET_NAME") // Укажите имя вашего S3-бакета
	endpointURL := os.Getenv("S3ENDURL")
	regionS3 := os.Getenv("REGION") // Укажите свой регион

	// Обработчик POST запроса
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Проверяем метод запроса
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Чтение тела запроса
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		// Проверяем значение тела запроса
		if string(body) != `{"START"}` {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		// Генерация пути к объектам в S3-бакете
		objectPrefix := os.Getenv("PREFIX")

		// запуск функции
		err = DownloadMXFFilesFromS3(bucketName, objectPrefix, regionS3, accessKey, secretKey, endpointURL)
		if err != nil {
			log.Fatal(err)
		}

		// Отправка успешного ответа
		w.WriteHeader(http.StatusOK)
	})

	// Запуск веб-сервера на порту 8080
	fmt.Println("Server listening on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// функция скачивает MXF файлы из S3-бакета
func DownloadMXFFilesFromS3(bucketName, objectPrefix, region, accessKeyID, secretAccessKey, endpointURL string) error {
	// Создаем сессию AWS с указанными учетными данными и регионом
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Endpoint:    aws.String(endpointURL),
		Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
	})
	if err != nil {
		return fmt.Errorf("failed to create AWS session: %v", err)
	}

	// Создаем новый клиент S3
	svc := s3.New(sess)

	// Создаем новый пул загрузчиков S3
	downloader := s3manager.NewDownloaderWithClient(svc, func(d *s3manager.Downloader) {
		d.Concurrency = 5 // Установите желаемое значение конкурентности загрузок
	})

	// Определяем путь каталога для сохранения файлов
	outputDirectory := "shared/"

	// Функция для обработки каждой страницы объектов
	pageHandler := func(page *s3.ListObjectsV2Output, lastPage bool) bool {
		// Перебираем каждый объект и скачиваем файлы с расширением .mxf
		for _, obj := range page.Contents {
			key := *obj.Key
			if filepath.Ext(key) == ".mxf" {
				// Формируем путь и имя файла для сохранения
				filePath := filepath.Join(outputDirectory, filepath.Base(key))

				// Получаем имя файла без расширения
				filename := filepath.Base(key[:len(key)-len(filepath.Ext(key))])

				// Проверяем наличие значения в базе данных
				if existsInDatabase(filename) {
					fmt.Printf("Skipping file (already exists): %s\n", filePath)
					continue // Пропускаем файл и переходим к следующей итерации цикла
				}

				// Создаем файл для записи
				file, err := os.Create(filePath)
				if err != nil {
					log.Printf("failed to create file: %v", err)
				}
				defer file.Close()

				// Скачиваем объект из S3
				_, err = downloader.Download(file, &s3.GetObjectInput{
					Bucket: aws.String(bucketName),
					Key:    aws.String(key),
				})
				if err != nil {
					fmt.Errorf("failed to download object from S3: %v", err)
				}

				// Здесь происходит запись имени файла без расширения в базу данных
				err = writeToDatabase(filepath.Base(key[:len(key)-len(filepath.Ext(key))]))
				if err != nil {
					fmt.Errorf("failed to write to database: %v", err)
				}

				fmt.Printf("Downloaded file: %s\n", filePath)
			}
		}

		// Возвращаем true, чтобы продолжить обработку следующей страницы
		return true
	}

	// Получаем список объектов в S3 с использованием пагинации
	err = svc.ListObjectsV2Pages(&s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(objectPrefix),
	}, pageHandler)

	if err != nil {
		return fmt.Errorf("failed to list objects in S3: %v", err)
	}

	return nil
}

func writeToDatabase(filename string) error {
	// Подключение к базе данных PostgreSQL
	dbConnectionString := os.Getenv("DB_CONNECTION_STRING")
	if dbConnectionString == "" {
		fmt.Println("DB_CONNECTION_STRING environment variable is not set")
	}

	db, err := sql.Open("postgres", dbConnectionString)
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)

	}
	defer db.Close()

	// Вставка данных в таблицу
	_, err = db.Exec("INSERT INTO transcode_jobs (id) VALUES ($1)", filename)
	if err != nil {
		log.Printf("Failed to insert data into table: %v", err)
	}

	log.Printf("Data saved successfully: ID=%s", filename)

	return nil
}

func getEnvVariable(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Fatal(name + " environment variable is not set")
	}
	return value
}
func existsInDatabase(filename string) bool {
	// Подключение к базе данных PostgreSQL
	dbConnectionString := os.Getenv("DB_CONNECTION_STRING")
	if dbConnectionString == "" {
		fmt.Println("DB_CONNECTION_STRING environment variable is not set")
	}

	db, err := sql.Open("postgres", dbConnectionString)
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		return false
	}
	defer db.Close()

	// Проверка наличия значения в таблице
	var count int
	row := db.QueryRow("SELECT COUNT(*) FROM transcode_jobs WHERE id = $1", filename)
	if err := row.Scan(&count); err != nil {
		log.Printf("Failed to check value in table: %v", err)
		return false
	}

	return count > 0
}
