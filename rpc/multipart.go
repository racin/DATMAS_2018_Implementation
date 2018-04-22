package rpc

import (
	"mime/multipart"
	"os"
	"bytes"
	"io"
)

func GetMultipartValues(values *map[string]io.Reader) (buffer *bytes.Buffer, boundary string){
	var b bytes.Buffer
	var err error
	w := multipart.NewWriter(&b)
	defer w.Close()

	for index, element := range *values {
		var writer io.Writer
		// If file has a close method.
		if file, ok := element.(io.Closer); ok {
			defer file.Close()
		}

		// Check if a file is added. Else add it as a regular data element.
		if file, ok := element.(*os.File); ok {
			writer, err = w.CreateFormFile(index, file.Name());
		} else {
			writer, err = w.CreateFormField(index);
		}

		// If there was no errors creating the form element, try to copy the element to it
		if err == nil {
			io.Copy(writer, element)
		}
	}

	return &b, w.FormDataContentType()
}
