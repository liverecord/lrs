package handlers

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/liverecord/lrs"
)

// File stucture defines upload
type File struct {
	Name         string    `json:"name"`
	Size         int       `json:"size"`
	Uploaded     int       `json:"uploaded"`
	Type         string    `json:"type"`
	LastModified time.Time `json:"lastModifiedDate"`
	PublicPath   string    `json:"path"`
}

// Upload Handler
func (Ctx *AppContext) Upload(frame lrs.Frame) {
	var f File
	frame.BindJSON(&f)
	fmt.Println(frame, f)
	Ctx.File = &f
}

// CancelUpload handler
func (Ctx *AppContext) CancelUpload(frame lrs.Frame) {
	Ctx.File = nil
}

// Uploader reads bytes from socket and writes them into file
func (Ctx *AppContext) Uploader(reader io.Reader) {
	if Ctx.File == nil {
		Ctx.Logger.Println("File is not set")
		return
	}
	file, err := ioutil.TempFile(Ctx.Cfg.UploadDir, "upload_")
	if err != nil {
		Ctx.Logger.WithError(err).Errorf("Unable to create temporary file. Is '%s' writable?", Ctx.Cfg.UploadDir)
	}
	fmt.Println("we've got the file!")
	bufferSize := Ctx.File.Size / 100
	if bufferSize < 4096 {
		bufferSize = 4096
	}
	if bufferSize > 2<<20 {
		bufferSize = 2 << 20
	}
	buf := make([]byte, bufferSize)
	total := 0
	Ctx.File.Uploaded = total
	Ctx.Pool.Write(Ctx.Ws, lrs.NewFrame(lrs.CancelUploadFrame, Ctx.File, ""))
	for {
		if Ctx.File == nil {
			Ctx.Logger.Println("File is not set or Upload was cancelled!")
			file.Close()
			os.Remove(file.Name())
			Ctx.Pool.Write(Ctx.Ws, lrs.NewFrame(lrs.CancelUploadFrame, Ctx.File, ""))
			return
		}
		fmt.Print(".")
		n, err := reader.Read(buf)
		fmt.Print(":")
		if err != nil {
			break
		}
		if n == 0 {
			break
		}
		file.Write(buf[:n])
		total += n
		Ctx.File.Uploaded = total
		Ctx.Pool.Write(Ctx.Ws, lrs.NewFrame(lrs.FileUploadFrame, Ctx.File, ""))
		if total == Ctx.File.Size {
			// we done with upload!
			break
		}
		time.Sleep(time.Millisecond * time.Duration(rand.Int63n(100)+50))
	}
	h := md5.New()
	if _, err := io.Copy(h, file); err != nil {
		Ctx.Logger.WithError(err).Errorln("Unable to read file")
	}

	t := time.Now()
	publicDir := path.Join(
		"files",
		fmt.Sprintf("%4d/%02d", t.Year(), t.Month()),
	)
	fp := path.Join(Ctx.Cfg.DocumentRoot, publicDir)
	err = os.MkdirAll(fp, 0777)
	if err != nil {
		Ctx.Logger.WithError(err).Errorln("Unable to create path")
		return
	}
	//
	ext := path.Ext(Ctx.File.Name)
	base := strings.TrimSuffix(path.Base(Ctx.File.Name), ext)
	safeFileName := fmt.Sprintf("%s-%x%s", slug.Make(base), h.Sum(nil)[:3], ext)
	Ctx.File.PublicPath = path.Join("/", publicDir, safeFileName)
	fp = path.Join(fp, safeFileName)
	file.Close()
	os.Rename(file.Name(), fp)
	Ctx.Pool.Write(Ctx.Ws, lrs.NewFrame(lrs.FileUploadFrame, Ctx.File, ""))
	Ctx.File = nil
}
