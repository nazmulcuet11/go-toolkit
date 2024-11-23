package toolkit

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
)

func TestTools_RandomString(t *testing.T) {
	var testTools Tools
	s := testTools.RandomString(10)
	if len(s) != 10 {
		t.Error("Wrong length random string")
	}
}

var uploadTests = []struct {
	name          string
	allowedTypes  []string
	renameFile    bool
	errorExpected bool
}{
	{
		name:          "allowed no rename",
		allowedTypes:  []string{"image/jpeg", "image/png"},
		renameFile:    false,
		errorExpected: false,
	},
	{
		name:          "allowed Rename",
		allowedTypes:  []string{"image/jpeg", "image/png"},
		renameFile:    true,
		errorExpected: false,
	},
	{
		name:          "not allowed",
		allowedTypes:  []string{"image/jpeg"},
		renameFile:    true,
		errorExpected: true,
	},
}

func TestTools_UploadFiles(t *testing.T) {
	for _, e := range uploadTests {
		// setup pipe to avoid buffering
		pr, pw := io.Pipe()
		writer := multipart.NewWriter(pw)
		wg := sync.WaitGroup{}
		wg.Add(1)

		go func() {
			defer writer.Close()
			defer wg.Done()

			part, err := writer.CreateFormFile("file", "./testdata/img.png")
			if err != nil {
				t.Error(err)
			}

			f, err := os.Open("./testdata/img.png")
			if err != nil {
				t.Error(err)
			}
			defer f.Close()

			img, _, err := image.Decode(f)
			if err != nil {
				t.Error(err)
			}

			err = png.Encode(part, img)
			if err != nil {
				t.Error(err)
			}
		}()

		request := httptest.NewRequest("POST", "/", pr)
		request.Header.Add("Content-Type", writer.FormDataContentType())

		var testTools Tools
		testTools.AllowedFileTypes = e.allowedTypes

		uploadedFiles, err := testTools.UploadFiles(request, "./testdata/uploads", e.renameFile)
		if err != nil && !e.errorExpected {
			t.Error(err)
		}

		if !e.errorExpected {
			filepath := fmt.Sprintf("./testdata/uploads/%s", uploadedFiles[0].NewFileName)
			_, err = os.Stat(filepath)
			if os.IsNotExist(err) {
				t.Errorf("%s: expected file to exist", e.name)
			}

			// cleanup
			_ = os.Remove(filepath)
		}

		if e.errorExpected && err == nil {
			t.Errorf("%s: expected error but none received", e.name)
		}

		wg.Wait()
	}
}

func TestTools_UploadFile(t *testing.T) {
	for _, e := range uploadTests {
		// setup pipe to avoid buffering
		pr, pw := io.Pipe()
		writer := multipart.NewWriter(pw)
		wg := sync.WaitGroup{}
		wg.Add(1)

		go func() {
			defer writer.Close()
			defer wg.Done()

			part, err := writer.CreateFormFile("file", "./testdata/img.png")
			if err != nil {
				t.Error(err)
			}

			f, err := os.Open("./testdata/img.png")
			if err != nil {
				t.Error(err)
			}
			defer f.Close()

			img, _, err := image.Decode(f)
			if err != nil {
				t.Error(err)
			}

			err = png.Encode(part, img)
			if err != nil {
				t.Error(err)
			}
		}()

		request := httptest.NewRequest("POST", "/", pr)
		request.Header.Add("Content-Type", writer.FormDataContentType())

		var testTools Tools
		testTools.AllowedFileTypes = e.allowedTypes

		uploadedFile, err := testTools.UploadFile(request, "./testdata/uploads", e.renameFile)
		if err != nil && !e.errorExpected {
			t.Error(err)
		}

		if !e.errorExpected {
			filepath := fmt.Sprintf("./testdata/uploads/%s", uploadedFile.NewFileName)
			_, err = os.Stat(filepath)
			if os.IsNotExist(err) {
				t.Errorf("%s: expected file to exist", e.name)
			}

			// cleanup
			_ = os.Remove(filepath)
		}

		if e.errorExpected && err == nil {
			t.Errorf("%s: expected error but none received", e.name)
		}

		wg.Wait()
	}
}

func TestTools_CreateDirIfNotExists(t *testing.T) {
	var testTools Tools
	err := testTools.CreateDirIfNotExists("./testdata/test-dir/subdir")
	if err != nil {
		t.Error(err)
	}

	err = testTools.CreateDirIfNotExists("./testdata/test-dir/subdir")
	if err != nil {
		t.Error(err)
	}

	// cleanup
	os.RemoveAll("./testdata/test-dir/")
}

var slugTests = []struct {
	name          string
	s             string
	expected      string
	errorExpected bool
}{
	{
		name:          "valid string",
		s:             "now is the time",
		expected:      "now-is-the-time",
		errorExpected: false,
	},
	{
		name:          "empty string",
		s:             "",
		expected:      "",
		errorExpected: true,
	},
	{
		name:          "complex string",
		s:             "NOW!!! is the time --+ &^??123",
		expected:      "now-is-the-time-123",
		errorExpected: false,
	},
	{
		name:          "empty result",
		s:             "--+ &^??",
		expected:      "",
		errorExpected: true,
	},
}

func TestTools_Slugify(t *testing.T) {
	var testTools Tools
	for _, e := range slugTests {
		slugg, err := testTools.Slugify(e.s)
		if err != nil && !e.errorExpected {
			t.Errorf("%s: error received when none expected: %s", e.name, err)
		}
		if !e.errorExpected && slugg != e.expected {
			t.Errorf("%s: wrong slugg returned. expected %s but got %s", e.name, e.expected, slugg)
		}
	}
}

func TestTools_DownloadStaticFile(t *testing.T) {
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	testTools := Tools{}
	testTools.DownloadStaticFile(rr, req, "./testdata", "pic.jpg", "puppy.jpg")
	res := rr.Result()
	defer res.Body.Close()
	if res.Header["Content-Length"][0] != "98827" {
		t.Error("wrong content length of", res.Header["Content-Length"][0])
	}

	if res.Header["Content-Disposition"][0] != "attachment; filename=\"puppy.jpg\"" {
		t.Error("wrong content disposition of", res.Header["Content-Disposition"][0])
	}

	_, err := io.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
	}
}

var jsonTests = []struct {
	name          string
	json          string
	errorExpected bool
	maxSize       int
	allowUnknown  bool
}{
	{
		name:          "good json",
		json:          `{"foo": "bar"}`,
		errorExpected: false,
		maxSize:       1 << 10,
		allowUnknown:  false,
	},
	{
		name:          "boadly-formatted json",
		json:          `{"foo":}`,
		errorExpected: true,
		maxSize:       1 << 10,
		allowUnknown:  false,
	},
	{
		name:          "incorrect type",
		json:          `{"foo": 1}`,
		errorExpected: true,
		maxSize:       1 << 10,
		allowUnknown:  false,
	},
	{
		name:          "more than one json",
		json:          `{"foo": "bar"}{"alpha": "beta"}`,
		errorExpected: true,
		maxSize:       1 << 10,
		allowUnknown:  false,
	},
	{
		name:          "empty body",
		json:          ``,
		errorExpected: true,
		maxSize:       1 << 10,
		allowUnknown:  false,
	},
	{
		name:          "syntax error",
		json:          `{"foo": "bar}`,
		errorExpected: true,
		maxSize:       1 << 10,
		allowUnknown:  false,
	},
	{
		name:          "unknown field not allowed",
		json:          `{"fooo": "bar"}`,
		errorExpected: true,
		maxSize:       1 << 10,
		allowUnknown:  false,
	},
	{
		name:          "unknown field allowed",
		json:          `{"fooo": "bar"}`,
		errorExpected: false,
		maxSize:       1 << 10,
		allowUnknown:  true,
	},
	{
		name:          "missing field name",
		json:          `{jack: "bar"}`,
		errorExpected: true,
		maxSize:       1 << 10,
		allowUnknown:  true,
	},
	{
		name:          "file too large",
		json:          `{"foo": "bar"}`,
		errorExpected: true,
		maxSize:       5,
		allowUnknown:  true,
	},
	{
		name:          "not a json",
		json:          `"hello world"`,
		errorExpected: true,
		maxSize:       1 << 10,
		allowUnknown:  true,
	},
}

func TestTools_ReadJSON(t *testing.T) {
	testTool := Tools{}
	for _, e := range jsonTests {
		testTool.MaxJSONSize = e.maxSize
		testTool.AllowUnknownFields = e.allowUnknown

		decodedJson := struct {
			Foo string `json:"foo"`
		}{}

		req, err := http.NewRequest("POST", "/", bytes.NewReader([]byte(e.json)))
		if err != nil {
			t.Log(err)
		}

		rr := httptest.NewRecorder()
		err = testTool.ReadJSON(rr, req, &decodedJson)

		if e.errorExpected && err == nil {
			t.Errorf("%s: error expected but none received", e.name)
		}

		if !e.errorExpected && err != nil {
			t.Errorf("%s: error not expected but one received %s", e.name, err.Error())
		}

		req.Body.Close()
	}
}
