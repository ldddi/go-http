package server

import (
	"html/template"
	"httpserver/pkg/utils"
	"log"
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

// 修改路径：从 cmd 目录执行时的相对路径
const baseDir = "../files"

// 修改模板路径
var dirTemplate *template.Template

func init() {
	rootDir, err := utils.GetProjectRoot()
	if err != nil {
		log.Fatal(err)
	}

	dirTemplate = template.Must(template.ParseFiles(filepath.Join(rootDir, "index.html")))
}

func BrowserGetHandler(w http.ResponseWriter, r *http.Request) {
	reqPath := mux.Vars(r)["path"]

	if reqPath == "" {
		reqPath = "/"
	}

	reqPath = strings.TrimPrefix(reqPath, "/")

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

		// 如果不是根目录，添加返回上级目录的链接
		if reqPath != "" && reqPath != "/" {
			parentPath := filepath.Dir(reqPath)
			if parentPath == "." {
				parentPath = ""
			}
			href := "/" + strings.ReplaceAll(parentPath, "\\", "/")
			items = append(items, FileItem{
				Name:  "..",
				Href:  href,
				IsDir: true,
			})
		}

		for _, f := range files {
			name := f.Name()
			var href string
			if reqPath == "" {
				href = "/" + name
			} else {
				href = "/" + filepath.Join(reqPath, name)
			}
			href = strings.ReplaceAll(href, "\\", "/")
			items = append(items, FileItem{
				Name:  name,
				Href:  href,
				IsDir: f.IsDir(),
			})
		}

		displayPath := "/" + reqPath
		if displayPath == "/" {
			displayPath = "/files"
		}

		err = dirTemplate.Execute(w, PageData{
			Path:  displayPath,
			Items: items,
		})
		if err != nil {
			http.Error(w, "模板渲染失败: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// 提供文件下载
		http.ServeFile(w, r, localPath)
	}
}
