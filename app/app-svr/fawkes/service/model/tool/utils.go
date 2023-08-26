package tool

import "mime/multipart"

type ContextValues struct {
	Username    string
	FromOpenAPI bool
	OpenAPIUser string
	File        multipart.File
	FileHeader  *multipart.FileHeader
}
