package toolkit

import (
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
)

func TestTools_RandomString(t *testing.T) {
	var testTools Tools

	s := testTools.RandomString(10)

	if len(s) != 10 {
		t.Errorf("RandomString(10) returned string of length %d, expected 10", len(s))
	}
}

var uploadTests = []struct {
	name             string
	allowedMimeTypes []string
	renameFile       bool
	errorExpected    bool
}{
	{name: "allowed no rename", allowedMimeTypes: []string{"image/jpeg", "image/png"}, renameFile: false, errorExpected: false},
	{name: "allowed with rename", allowedMimeTypes: []string{"image/jpeg", "image/png"}, renameFile: true, errorExpected: false},
	{name: "not allowed mime type", allowedMimeTypes: []string{"image/png"}, renameFile: false, errorExpected: true},
}

func TestTools_UploadFiles(t *testing.T) {
	for _, e := range uploadTests {
		pr, pw := io.Pipe()
		writer := multipart.NewWriter(pw)
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer writer.Close()
			defer wg.Done()

			part, err := writer.CreateFormFile("file", "./testdata/img.jpg")
			if err != nil {
				t.Errorf("Error creating form file: %v", err)
				return
			}

			file, err := os.Open("./testdata/img.jpg")
			if err != nil {
				t.Errorf("Error opening test file: %v", err)
				return
			}
			defer file.Close()

			img, _, err := image.Decode(file)
			if err != nil {
				t.Errorf("Error decoding image: %v", err)
				return
			}

			jpeg.Encode(part, img, nil)
		}()

		request := httptest.NewRequest("POST", "/", pr)
		request.Header.Add("Content-Type", writer.FormDataContentType())

		var testTools Tools
		testTools.AllowedFileTypes = e.allowedMimeTypes
		uploadedFiles, err := testTools.UploadFiles(request, "./testdata/uploads/", e.renameFile)
		if err != nil && !e.errorExpected {
			t.Errorf("%s: UploadFiles returned error: %v", e.name, err)
		}

		if !e.errorExpected {
			if _, err := os.Stat(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles[0].NewFileName)); os.IsNotExist(err) {
				t.Errorf("%s: Uploaded file not found: %v", e.name, err)
			} else {
				os.Remove(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles[0].NewFileName))
			}
		}

		if !e.errorExpected && err != nil {
			t.Errorf("%s: UploadFiles returned unexpected error: %v", e.name, err)
		}
		wg.Wait()
	}
}

func TestTools_UploadOneFile(t *testing.T) {
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)
	go func() {
		defer writer.Close()

		part, err := writer.CreateFormFile("file", "./testdata/img.jpg")
		if err != nil {
			t.Errorf("Error creating form file: %v", err)
			return
		}

		file, err := os.Open("./testdata/img.jpg")
		if err != nil {
			t.Errorf("Error opening test file: %v", err)
			return
		}
		defer file.Close()

		img, _, err := image.Decode(file)
		if err != nil {
			t.Errorf("Error decoding image: %v", err)
			return
		}

		jpeg.Encode(part, img, nil)
	}()

	request := httptest.NewRequest("POST", "/", pr)
	request.Header.Add("Content-Type", writer.FormDataContentType())

	var testTools Tools
	uploadedFiles, err := testTools.UploadOneFile(request, "./testdata/uploads/", true)
	if err != nil {
		t.Errorf("Upload one file returned error: %v", err)
	}

	if _, err := os.Stat(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles.NewFileName)); os.IsNotExist(err) {
		t.Errorf("Uploaded file not found: %v", err.Error())
	}
	_ = os.Remove(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles.NewFileName))
}
