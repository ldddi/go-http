package httpfileserver

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
)

type FileItem struct {
	Name  string
	Href  string
	IsDir bool
}

type PageData struct {
	Path  string
	Items []FileItem
}

const baseDir = "./files"

var dirTemplate = template.Must(template.ParseFiles("index.html"))

func BrowserGetHandler(w http.ResponseWriter, r *http.Request) {
	reqPath := mux.Vars(r)["path"]
	localPath := filepath.Join(baseDir, reqPath)

	info, err := os.Stat(localPath)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if info.IsDir() {
		files, err := os.ReadDir(localPath)
		if err != nil {
			http.Error(w, "无法读取目录", http.StatusInternalServerError)
			return
		}

		var items []FileItem
		for _, f := range files {
			name := f.Name()
			href := filepath.Join("/", reqPath, name)
			href = strings.ReplaceAll(href, "\\", "/")
			items = append(items, FileItem{
				Name:  name,
				Href:  href,
				IsDir: f.IsDir(),
			})
		}

		err = dirTemplate.Execute(w, PageData{
			Path:  "/" + reqPath,
			Items: items,
		})
		if err != nil {
			http.Error(w, "模板渲染失败: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		http.ServeFile(w, r, localPath)
	}
}
