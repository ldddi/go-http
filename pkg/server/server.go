package server

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
	ReadTimeout int `json:"read_timeout"`
	// write timeout
	WriteTimeout int `json:"write_timeout"`
	// enable auth
	EnableAuth bool `json:"enable_auth"`
	// enable CORS
	EnableCORS bool `json:"enable_cors"`
	// file Naming strategy
	FileNamingStrategy string `json:"file_naming_strategy"`
}

type Server struct {
	ServerConfig
	// fs afero.Fs
}
