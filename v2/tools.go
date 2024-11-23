package toolkit

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const randomCharacterSet = "abcdefghijklijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ01223456789"

// Tools is the type to instantiate the module. Any variable of this type have the access to all
// the methods with the receiver *Tools.
type Tools struct {
	MaxFileSize        int
	AllowedFileTypes   []string
	MaxJSONSize        int
	AllowUnknownFields bool
}

// RandomString returns a string of random character of length n.
func (t *Tools) RandomString(n int) string {
	s := make([]rune, n)
	r := []rune(randomCharacterSet)
	for i := range s {
		p, _ := rand.Prime(rand.Reader, 64)
		index := p.Uint64() % uint64(len(r))
		s[i] = r[index]
	}
	return string(s)
}

// UploadedFile is a struct to save information about uploaded file
type UploadedFile struct {
	NewFileName      string
	OriginalFileName string
	FileSize         int64
}

func (t *Tools) UploadFile(r *http.Request, uploadDir string, rename ...bool) (*UploadedFile, error) {
	renameFile := true
	if len(rename) > 0 {
		renameFile = rename[0]
	}

	files, err := t.UploadFiles(r, uploadDir, renameFile)
	if err != nil {
		return nil, err
	}
	return files[0], nil
}

func (t *Tools) UploadFiles(r *http.Request, uploadDir string, rename ...bool) ([]*UploadedFile, error) {
	renameFile := true
	if len(rename) > 0 {
		renameFile = rename[0]
	}

	var uploadedFiles []*UploadedFile

	if t.MaxFileSize == 0 {
		t.MaxFileSize = 1 << 30 // 1 GB
	}

	err := t.CreateDirIfNotExists(uploadDir)
	if err != nil {
		return nil, err
	}

	err = r.ParseMultipartForm(int64(t.MaxFileSize))
	if err != nil {
		return nil, errors.New("uploaded file is too big")
	}

	for _, fileHeaders := range r.MultipartForm.File {
		for _, fileHeader := range fileHeaders {
			var uploadedFile UploadedFile
			infile, err := fileHeader.Open()
			if err != nil {
				return uploadedFiles, err
			}
			defer infile.Close()

			buff := make([]byte, 512)
			_, err = infile.Read(buff)
			if err != nil {
				return uploadedFiles, err
			}

			// check if file type is permitted
			allowed := false
			fileType := http.DetectContentType(buff)
			if len(t.AllowedFileTypes) > 0 {
				for _, x := range t.AllowedFileTypes {
					if strings.EqualFold(x, fileType) {
						allowed = true
						break
					}
				}
			} else {
				allowed = true
			}

			if !allowed {
				return uploadedFiles, errors.New("uploaded file type is not permitted")
			}

			_, err = infile.Seek(0, 0)
			if err != nil {
				return uploadedFiles, err
			}

			uploadedFile.OriginalFileName = fileHeader.Filename
			if renameFile {
				uploadedFile.NewFileName = fmt.Sprintf("%s%s", t.RandomString(32), filepath.Ext(fileHeader.Filename))
			} else {
				uploadedFile.NewFileName = fileHeader.Filename
			}

			var outfile *os.File
			defer outfile.Close()

			outfile, err = os.Create(filepath.Join(uploadDir, uploadedFile.NewFileName))
			if err != nil {
				return uploadedFiles, err
			} else {
				fileSize, err := io.Copy(outfile, infile)
				if err != nil {
					return uploadedFiles, err
				}
				uploadedFile.FileSize = fileSize
			}

			uploadedFiles = append(uploadedFiles, &uploadedFile)
		}
	}

	return uploadedFiles, nil
}

// CreateDirIfNotExists creates a directory if not exist
func (t *Tools) CreateDirIfNotExists(path string) error {
	const mode = 0755
	_, err := os.Stat(path)
	log.Println(err)
	if os.IsNotExist(err) {
		err = os.MkdirAll(path, mode)
		log.Println(err)
		if err != nil {
			return err
		}
	}
	return err
}

// Slugify is very simple function to generate slug from a string
func (t *Tools) Slugify(s string) (string, error) {
	if s == "" {
		return "", errors.New("empty string not permitted")
	}

	s = strings.ToLower(s)
	re := regexp.MustCompile(`[^a-z\d]+`)
	s = re.ReplaceAllString(s, "-")
	slug := strings.Trim(s, "-")
	if slug == "" {
		return "", errors.New("after removing charachter slug is empty")
	}

	return slug, nil
}

// DownloadStaticFile downloads a file, and tries to force the browser to avoid displaying it
// in the browser window by setting content disposition. It also allows specification fo the
// display name
func (t *Tools) DownloadStaticFile(w http.ResponseWriter, r *http.Request, filePath, displayName string) {
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", displayName))
	http.ServeFile(w, r, filePath)
}

type JSONResponse struct {
	Error   bool        `json:"error"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (t *Tools) ReadJSON(w http.ResponseWriter, r *http.Request, data interface{}) error {
	maxBytes := 1 << 20 // 1 mega byte
	if t.MaxJSONSize != 0 {
		maxBytes = t.MaxJSONSize
	}

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	dec := json.NewDecoder(r.Body)

	if !t.AllowUnknownFields {
		dec.DisallowUnknownFields()
	}

	err := dec.Decode(data)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at charachter %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case strings.HasPrefix(err.Error(), "json: unknown field"):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)

		case errors.As(err, &invalidUnmarshalError):
			return fmt.Errorf("error unmarshaling JSON %s", err.Error())

		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body contains more than one json value")
	}

	return nil
}

func (t *Tools) WriteJSON(w http.ResponseWriter, status int, data interface{}, headers ...http.Header) error {
	out, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(out)
	if err != nil {
		return err
	}

	return nil
}

func (t *Tools) ErrorJSON(w http.ResponseWriter, err error, status ...int) error {
	statusCode := http.StatusBadRequest
	if len(status) > 0 {
		statusCode = status[0]
	}

	payload := JSONResponse{
		Error:   true,
		Message: err.Error(),
	}
	return t.WriteJSON(w, statusCode, payload)
}

func (t *Tools) PushJSONToRemote(uri string, data interface{}, client ...*http.Client) (*http.Response, int, error) {
	// create json
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, 0, err
	}

	// check for custom http client
	httClient := &http.Client{}
	if len(client) > 0 {
		httClient = client[0]
	}

	// build request
	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	// call remote uri
	res, err := httClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	// send the response back
	return res, res.StatusCode, nil
}
