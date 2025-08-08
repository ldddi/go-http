package server

import (
	"errors"
	"html/template"
	resp "httpserver/internal/response"
	"httpserver/pkg/utils"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
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

// 修改模板路径
var dirTemplate *template.Template

func init() {
	rootDir, err := utils.GetProjectRoot()
	if err != nil {
		log.Fatal(err)
	}

	dirTemplate = template.Must(template.ParseFiles(filepath.Join(rootDir, "index.html")))
}

func (s *Server) BrowserGetHandler(w http.ResponseWriter, r *http.Request) resp.Response {
	reqPath := mux.Vars(r)["path"]
	reqPath = strings.TrimPrefix(reqPath, "/")
	localPath := filepath.Join(s.WorkDir, reqPath)

	info, err := os.Stat(localPath)
	if err != nil {
		log.Println("file not exit or no auth", err)
		return errorResponse(http.StatusBadRequest, errors.New("file not exit or no auth"))
	}

	if info.IsDir() {
		files, err := os.ReadDir(localPath)
		if err != nil {
			log.Println("failed to open directory", err)
			return errorResponse(http.StatusBadRequest, errors.New("failed to open directory"))
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

		sort.Slice(items, func(i, j int) bool {
			return items[i].IsDir && !items[j].IsDir
		})

		err = dirTemplate.Execute(w, PageData{
			Path:  reqPath,
			Items: items,
		})
		if err != nil {
			log.Println("failed to execute template: ", err)
			return errorResponse(http.StatusInternalServerError, errors.New("failed to execute template"))
		}
	} else {
		// 提供文件下载
		w.Header().Set("Content-Disposition", "attachment;")
		http.ServeFile(w, r, localPath)
	}
	return nil
}
