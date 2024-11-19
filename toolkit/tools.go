package toolkit

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

const randomCharacterSet = "abcdefghijklijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ01223456789"

// Tools is the type to instantiate the module. Any variable of this type have the access to all
// the methods with the receiver *Tools.
type Tools struct {
	MaxFileSize      int
	AllowedFileTypes []string
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
func (t *Tools) DownloadStaticFile(w http.ResponseWriter, r *http.Request, p, file, displayName string) {
	fp := path.Join(p, file)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", displayName))
	http.ServeFile(w, r, fp)
}
