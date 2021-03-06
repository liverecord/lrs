package lrs

import (
	"bytes"
	"encoding/json"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

func setStatic(b []byte, title string, body string) []byte {
	b = bytes.Replace(b, S2BA("<title>LiveRecord"), S2BA("<title>"+title), 1)
	return bytes.Replace(
		b,
		S2BA("<!-- content -->"),
		S2BA(
			"<h1>"+title+"</h1>\n\n"+
				body+"\n<hr>\n<p><a href=\"/\">Home</a></p><hr>",
		),
		1,
	)
}

func inlineTopic(b []byte, t Topic) []byte {
	return setStatic(b, t.Title, t.Body)
}

func inlineCategory(b []byte, c Category) []byte {
	return setStatic(b, c.Name, c.Description)
}

func maxDate(times ...time.Time) time.Time {
	var ct time.Time
	if len(times) < 1 {
		return ct
	}
	ct = times[0]
	for _, v := range times {
		if v.Unix() > ct.Unix() {
			ct = v
		}
	}
	return ct
}

func containsDotDot(v string) bool {
	if !strings.Contains(v, "..") {
		return false
	}
	for _, ent := range strings.FieldsFunc(v, isSlashRune) {
		if ent == ".." {
			return true
		}
	}
	return false
}

func isSlashRune(r rune) bool { return r == '/' || r == '\\' }

func serveVFS(w http.ResponseWriter, r *http.Request) {

}


func respond(w http.ResponseWriter, logger *logrus.Logger, code int, html []byte) {
	w.WriteHeader(404)
	_, err := w.Write(html)
	if err != nil {
		logger.WithError(err)
	}
}

func handleDynamicContent(cfg *Config, db *gorm.DB, logger *logrus.Logger) func(w http.ResponseWriter, r *http.Request) {
	// read file

	shtml, err := ioutil.ReadFile(path.Join(cfg.DocumentRoot, "app-dist/index.html"))
	logger.Info(shtml)

	if err != nil {
		logger.WithError(err)
	}
	cfgJson, _ := json.Marshal(cfg)
	cfgJson = bytes.Join([][]byte{S2BA("liveRecordConfig = "), cfgJson, S2BA(";//")}, S2BA(" "))



	return func(w http.ResponseWriter, r *http.Request) {

		if cfg.Debug {
			shtml, err = ioutil.ReadFile(path.Join(cfg.DocumentRoot, "app-dist/index.html"))
			logger.Info(shtml)

			if err != nil {
				logger.WithError(err)
			}
			cfgJson, _ = json.Marshal(cfg)
			cfgJson = bytes.Join([][]byte{S2BA("liveRecordConfig = "), cfgJson, S2BA(";//")}, S2BA(" "))
		}

		logger.Infof("%s %s", r.Method, r.RequestURI)
		// serve it
		html := bytes.Replace(
			shtml,
			S2BA("liveRecordConfig = "),
			cfgJson,
			1,
		)

		upath := r.URL.Path
		if !strings.HasPrefix(upath, "/") {
			upath = "/" + upath
			r.URL.Path = upath
		}
		if containsDotDot(r.URL.Path) {
			respond(w, logger, 404, html)
			return
		}
		elements := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		logger.Infof("%v", elements)
		if len(elements[0]) == 0 || elements[0] == "index.html" {
			respond(w, logger, 200, html)
			return
		}
		fname := path.Join(cfg.DocumentRoot, path.Clean(upath))
		logger.Debug(fname)

		f, err := os.Open(fname)
		statusCode := 200
		if err != nil {
			switch len(elements) {
			case 2:
				var found Topic
				db.
					Where("private = 0").
					Where("slug = ?", elements[1]).
					First(&found)
				if found.ID > 0 {
					html = inlineTopic(html, found)
					w.Header().Set("Last-Modified",
						maxDate(found.CreatedAt, found.UpdatedAt, found.CommentedAt).Format(time.RFC1123),
					)
				} else {
					// w.WriteHeader(404)
					statusCode = 404
				}
			case 1:

				var found Category
				db.
					Where("slug = ?", elements[0]).
					First(&found)
				if found.ID > 0 {
					html = inlineCategory(html, found)
				} else {
					statusCode = 404
				}
			default:
				statusCode = 200
			}
			respond(w, logger, statusCode, html)
			return
		}
		defer f.Close()
		d, err := f.Stat()
		if err != nil {
			respond(w, logger, 404, html)
			return
		}
		if d.IsDir() {
			respond(w, logger, 403, html)
			return
		}
		http.ServeFile(w, r, fname)
		return
	}
}

//RegisterStaticHandlers registers handlers to serve static content
func RegisterStaticHandlers(cfg *Config, db *gorm.DB, logger *logrus.Logger) {
	http.HandleFunc("/", handleDynamicContent(cfg, db, logger))
}
