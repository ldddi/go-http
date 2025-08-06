package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
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

func NewServer(config ServerConfig) *Server {
	return &Server{ServerConfig: config}
}

func f(w http.ResponseWriter, r *http.Request) {

}

func (s *Server) Start(stop chan os.Signal, ready chan struct{}) error {
	r := mux.NewRouter()
	r.HandleFunc("/upload", f)

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
