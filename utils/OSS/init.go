package OSS

import (
	"context"
	"github.com/hewo233/house-system-backend/shared/consts"
	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"log"
	"os"
)

func connectOSS() {
	if err := godotenv.Load(consts.OSSEnvFIle); err != nil {
		log.Fatal("Error loading .env file")
	}

	endpoint = os.Getenv("ENDPOINT")
	accessKey := os.Getenv("ACCESS_KEY")
	secretKey := os.Getenv("SECRET_KEY")
	bucket = os.Getenv("BUCKET")
	useSSL = os.Getenv("USE_SSL") == "true"

	var err error
	minioClient, err = minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: true,
	})
	if err != nil {
		log.Fatalf("Failed to create MinIO client: %v", err)
	}

	ctx := context.Background()
	exists, err := minioClient.BucketExists(ctx, bucket)
	if err != nil {
		log.Fatal("failed to check bucket:", err)
	}
	if !exists {
		log.Fatal("bucket does not exist")
	}

	log.Println("\033[32mMinIO client initialized successfully\033[0m")
}

func Init() {
	connectOSS()
}
