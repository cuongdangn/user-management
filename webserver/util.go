package webserver

import (
	"mime/multipart"
	"net/http"
)

func GetFileContentType(out multipart.File) (string, error) {

	buffer := make([]byte, 512)

	_, err := out.Read(buffer)
	if err != nil {
		return "", err
	}
	contentType := http.DetectContentType(buffer)

	return contentType, nil
}
