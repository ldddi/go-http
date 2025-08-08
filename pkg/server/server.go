package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
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
	// enable auth
	EnableAuth *bool `json:"enable_auth"`
	// enable CORS
	EnableCORS *bool `json:"enable_cors"`
	// file Naming strategy
	FileNamingStrategy string `json:"file_naming_strategy"`
}

type Server struct {
	ServerConfig
	// fs afero.Fs
}

type ErrorMsg struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}

type SuccessMsg struct {
	Ok   bool   `json:"ok"`
	Data string `json:"data"`
}

func NewServer(config ServerConfig) *Server {
	return &Server{ServerConfig: config}
}

func (s *Server) handle(f func(http.ResponseWriter, *http.Request) (int, any)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status, result := f(w, r)
		var respBody []byte

		// to json
		if result != nil {
			switch v := result.(type) {
			case error:
				result = ErrorMsg{Ok: false, Error: v.Error()}
			}

			respBytes, err := json.Marshal(result)
			if err != nil {
				log.Printf("failed to marshal response: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			respBody = respBytes
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)

		if _, err := w.Write(respBody); err != nil {
			log.Printf("failed to write response: %v", err)
		}

	}
}

var (
	overwriteKey = "overwrite"
	fileKey      = "file"
)

// query params:
// - overwrite: if true, allows overwriting the existing file
func (s *Server) uploadFileHandler(w http.ResponseWriter, r *http.Request) (int, any) {
	overwrite := r.URL.Query().Get(overwriteKey)
	if ok, err := strconv.ParseBool(overwrite); err != nil {
		log.Printf("invalid overwrite parameter: %v\n", err)
		return http.StatusBadRequest, errors.New("invalid overwrite parameter, please use true or false")
	}

	file, info, err := r.FormFile(fileKey)
	if err != nil {
		log.Printf("failed to get file from request: %v\n", err)
		return http.StatusBadRequest, errors.New("failed to get file from request")
	}
	defer file.Close()

	return http.StatusOK, SuccessMsg{Ok: true, Data: "File uploaded successfully"}
}

func (s *Server) downloadFileHandler(w http.ResponseWriter, r *http.Request) (int, any) {
	return http.StatusOK, SuccessMsg{Ok: true, Data: "File downloaded successfully"}
}

func (s *Server) getFileHandler(w http.ResponseWriter, r *http.Request) (int, any) {
	return http.StatusOK, SuccessMsg{Ok: true, Data: "File retrieved successfully"}
}

func (s *Server) deleteFileHandler(w http.ResponseWriter, r *http.Request) (int, any) {
	return http.StatusOK, SuccessMsg{Ok: true, Data: "File deleted successfully"}
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Header().Set("Content-Type", "application/json")
	resp := ErrorMsg{Ok: false, Error: "Not Found"}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}

func methodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Header().Set("Content-Type", "application/json")
	resp := ErrorMsg{Ok: false, Error: "Method not allowed"}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}

// Start starts the HTTP server and listens for shutdown signals
// stop: channel to receive termination signals for graceful shutdown
// ready: channel to signal when server is ready to accept connections
func (s *Server) Start(stop chan os.Signal, ready chan struct{}) error {
	r := mux.NewRouter()
	r.HandleFunc("/upload", s.handle(s.uploadFileHandler)).Methods("POST")
	r.HandleFunc("/download", s.handle(s.downloadFileHandler)).Methods("GET")
	r.HandleFunc("/files", s.handle(s.getFileHandler)).Methods("GET")
	r.HandleFunc("/delete", s.handle(s.deleteFileHandler)).Methods("DELETE")

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
		log.Printf("server start to: %v", srv.Addr)
		if err := srv.Serve(l); err != nil && err != http.ErrServerClosed {
			ret <- fmt.Errorf("failed to start server: %w", err)
		}
		log.Printf("server successful shut down")
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
