package entity

import "cloud.google.com/go/storage"

type SttProcessFormat struct {
	JwtSecret  string
	BucketName string
	Bucket     *storage.BucketHandle
	Client     *storage.Client
}
