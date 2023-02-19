package entity

import "mime/multipart"

type UploadFormat struct {
	Token     string
	AudioFile *multipart.FileHeader
}

// type AudioFile struct {
// 	File *multipart.FileHeader
// }
