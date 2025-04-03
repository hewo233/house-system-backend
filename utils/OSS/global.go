package OSS

import "github.com/minio/minio-go/v7"

var minioClient *minio.Client
var bucket string
var endpoint string
var useSSL bool
