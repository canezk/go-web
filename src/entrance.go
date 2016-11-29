/**
支持上传，查看图片
默认文件夹：uploads
*/
package main

import (
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"runtime/debug"
)

const (
	UPLOAD_DIR   = "./uploads"
	TEMPLATE_DIR = "./views"
)

var templates = make(map[string]*template.Template)

func uploadHanlder(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		if err := renderHtml(w, "upload", nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}
	if r.Method == "POST" {
		f, h, err := r.FormFile("image")
		check(err)
		filename := h.Filename
		defer f.Close()
		t, err := os.Create(UPLOAD_DIR + "/" + filename)
		check(err)
		defer t.Close()
		if _, err := io.Copy(t, f); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/view?id="+filename,
			http.StatusFound)
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	imgId := r.FormValue("id")
	imgPath := UPLOAD_DIR + "/" + imgId
	log.Println("Image id:", imgId)
	if exist := isExist(imgPath); !exist {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "image")
	http.ServeFile(w, r, imgPath)
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	fileInfoArr, err := ioutil.ReadDir(UPLOAD_DIR)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	locals := make(map[string]interface{})
	images := []string{}
	for _, fileInfo := range fileInfoArr {
		images = append(images, fileInfo.Name())
	}
	if err = renderHtml(w, "list", locals); err != nil {
		http.Error(w, err.Error(),
			http.StatusInternalServerError)
	}
}

func safeHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if e, ok := recover().(error); ok {
				http.Error(w, e.Error(), http.StatusInternalServerError)
				log.Println(string(debug.Stack()))
			}
		}()
		fn(w, r)
	}
}

func init() {
	fileInfoArr, err := ioutil.ReadDir(TEMPLATE_DIR)
	check(err)

	var templateName, templatePath string
	for _, fileInfo := range fileInfoArr {
		templateName = fileInfo.Name()
		if ext := path.Ext(templateName); ext != ".html" {
			continue
		}
		templatePath = TEMPLATE_DIR + "/" + templateName
		log.Println("Template path is ", templatePath)
		t := template.Must(template.ParseFiles(templatePath))
		templates[templateName] = t
		log.Println("Template  is ", t)
	}
}

func main() {
	http.Handle("/", safeHandler(listHandler))
	http.Handle("/upload", safeHandler(uploadHanlder))
	http.Handle("/view", safeHandler(viewHandler))
	err := http.ListenAndServe(":8080", nil)
	check(err)
}

/**
utils function
*/
func isExist(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return os.IsExist(err)
}

func renderHtml(w http.ResponseWriter, tmpPath string, locals map[string]interface{}) (err error) {
	log.Println("Template is ", templates[tmpPath + ".html"])
	err = templates[tmpPath + ".html"].Execute(w, locals)
	return
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
