package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	resp "httpserver/internal/response"
	logger "httpserver/pkg/log"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type ServerConfig struct {
	// server address 127.0.0.1:8888
	Addr string `json:"addr"`
	// server run root dir
	WorkDir string `json:"work_dir"`
	// file upload max size bytes
	MaxUploadSize int64 `json:"max_upload_size"`
	// shutdown timeout
	ShutdownTimeout int `json:"shutdown_time"`
	// read timeout
	ReadTimeout time.Duration `json:"read_timeout"`
	// write timeout
	WriteTimeout time.Duration `json:"write_timeout"`
}

type Server struct {
	ServerConfig
	// fs afero.Fs
}

func NewServer(config ServerConfig) *Server {
	return &Server{ServerConfig: config}
}

func errorResponse(status int, message error) resp.Response {
	return resp.NewErrorMsgBuilder().WithStatus(status).WithMessage(message.Error()).Build()
}

func successResponse(status int, message string, data any) resp.Response {
	if status < 200 || status >= 300 {
		logger.Warn(fmt.Sprintf("success response with non-2xx status: %d", status))
	}
	return resp.NewSuccessMsgBuilder().WithStatus(status).WithMessage(message).WithData(data).Build()
}

func (s *Server) handle(f func(http.ResponseWriter, *http.Request) resp.Response) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		result := f(w, r)
		var respBody []byte
		logger.Info(result)
		// to json
		respBody, err := json.Marshal(result)
		if err != nil {
			logger.Error(fmt.Sprintf("failed to marshal response: %v", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")

		w.WriteHeader(result.GetStatus())

		if _, err := w.Write(respBody); err != nil {
			logger.Error(fmt.Sprintf("failed to write response: %v", err))
		}
	}
}

// query params:
// - overwrite: if true, allows overwriting the existing file
// -distPath: save file to distPath, default to workDir
func (s *Server) uploadFileHandler(w http.ResponseWriter, r *http.Request) resp.Response {
	distPath := r.FormValue("distPath")
	file, info, err := r.FormFile("file")
	if err != nil {
		logger.Error(fmt.Sprintf("failed to get file from request: %v\n", err))
		return errorResponse(http.StatusBadRequest, errors.New("failed to get file from request"))
	}
	defer file.Close()

	if distPath == "" {
		distPath = filepath.Join(s.WorkDir, info.Filename)
	} else {
		distPath = strings.TrimSpace(distPath)
		distPath = strings.TrimLeft(distPath, "/\\")
		distPath = filepath.Join(s.WorkDir, distPath, info.Filename)
	}

	if _, err := os.Stat(distPath); err == nil {
		logger.Error("file already exist")
		return errorResponse(http.StatusBadRequest, errors.New("file already exist"))
	}

	if _, err := os.Stat(filepath.Dir(distPath)); err != nil {
		if err = os.MkdirAll(filepath.Dir(distPath), 0755); err != nil {
			logger.Error(fmt.Sprintf("failed to make dir: %v", err))
			return errorResponse(http.StatusInternalServerError, errors.New("failed to make dir"))
		}
	}

	distFile, err := os.Create(distPath)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to create dist file: %v", err))
		return errorResponse(http.StatusInternalServerError, errors.New("failed to create dist file"))
	}
	defer distFile.Close()

	srcFile := http.MaxBytesReader(w, file, s.MaxUploadSize)
	if _, err := io.Copy(distFile, srcFile); err != nil {
		logger.Error(fmt.Sprintf("failed to upload file: %v", err))
		return errorResponse(http.StatusInternalServerError, errors.New("failed to upload file"))
	}

	return successResponse(http.StatusOK, "File uploaded successfully", nil)
}

// query params:
// - path: the path of the file to download
func (s *Server) downloadFileHandler(w http.ResponseWriter, r *http.Request) resp.Response {
	path := r.URL.Query().Get("path")
	path = strings.TrimPrefix(path, "/")
	localPath := filepath.Join(s.WorkDir, path)
	info, err := os.Stat(localPath)
	if err != nil {
		logger.Error(fmt.Sprintf("file not found: %v", err))
		return errorResponse(http.StatusNotFound, errors.New("file not found"))
	}

	if info.IsDir() {
		logger.Error("cannot download a directory")
		return errorResponse(http.StatusBadRequest, errors.New("cannot download a directory"))
	}

	file, err := os.Open(localPath)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to open file: %v", err))
		return errorResponse(http.StatusInternalServerError, errors.New("failed to open file"))
	}
	defer file.Close()

	_, err = io.Copy(w, file)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to write file to response: %v", err))
		return errorResponse(http.StatusInternalServerError, errors.New("failed to write file to response"))
	}

	return successResponse(http.StatusOK, "File download successfully", nil)
}

// query params:
// - path: the path of the file to delete
func (s *Server) deleteFileHandler(w http.ResponseWriter, r *http.Request) resp.Response {
	path := r.URL.Query().Get("path")
	path = strings.TrimPrefix(path, "/")
	localPath := filepath.Join(s.WorkDir, path)
	info, err := os.Stat(localPath)
	if err != nil {
		logger.Error(fmt.Sprintf("file not found: %v", err))
		return errorResponse(http.StatusNotFound, errors.New("file not found"))
	}

	if info.IsDir() {
		if err = os.RemoveAll(localPath); err != nil {
			logger.Error(fmt.Sprintf("failed to delete directory: %v", err))
			return errorResponse(http.StatusInternalServerError, errors.New("failed to delete directory"))
		}
		return successResponse(http.StatusOK, "Directory delete successfully", nil)
	} else {
		if err = os.Remove(localPath); err != nil {
			logger.Error(fmt.Sprintf("failed to delete file: %v", err))
			return errorResponse(http.StatusInternalServerError, errors.New("failed to delete file"))
		}
		return successResponse(http.StatusOK, "File delete successfully", nil)
	}
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Header().Set("Content-Type", "application/json")
	resp := errorResponse(http.StatusNotFound, errors.New("not found"))
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.Error(fmt.Sprintf("failed to write response: %v", err))
	}
}

func methodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	resp := errorResponse(http.StatusMethodNotAllowed, errors.New("method not allowed"))
	respBody, err := json.Marshal(resp)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to marshal response: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)

	if _, err := w.Write(respBody); err != nil {
		logger.Error(fmt.Sprintf("failed to write response: %v", err))
	}
}

// Start starts the HTTP server and listens for shutdown signals
// stop: channel to receive termination signals for graceful shutdown
// ready: channel to signal when server is ready to accept connections
func (s *Server) Start(stop chan os.Signal, ready chan struct{}) error {
	r := mux.NewRouter()
	r.HandleFunc("/upload", s.handle(s.uploadFileHandler)).Methods("POST")
	r.HandleFunc("/download", s.handle(s.downloadFileHandler)).Methods("GET")
	r.HandleFunc("/delete", s.handle(s.deleteFileHandler)).Methods("DELETE")

	r.HandleFunc("/files/{path:.*}", s.BrowserGetHandler).Methods("GET")

	r.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	r.MethodNotAllowedHandler = http.HandlerFunc(methodNotAllowedHandler)

	srv := http.Server{
		Addr:         s.Addr,
		Handler:      r,
		ReadTimeout:  s.ReadTimeout,
		WriteTimeout: s.WriteTimeout,
	}

	l, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return fmt.Errorf("fail to create Listener: %w", err)
	}

	ret := make(chan error, 1)
	go func() {
		logger.Info(fmt.Sprintf("server start to: %v", srv.Addr))
		if err := srv.Serve(l); err != nil && err != http.ErrServerClosed {
			ret <- fmt.Errorf("failed to start server: %w", err)
		}
		logger.Info("server successful shut down")
		ret <- nil
	}()

	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shut down server: %w", err)
	}

	return <-ret
}
