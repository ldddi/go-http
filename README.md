<p>
  <img src="https://github.com/ldddi/http-file-server/blob/main/pic.png">
</p>

# Go HTTP File Server 📁

This is a lightweight HTTP file server written in Go that provides a simple web interface for file management. It supports file uploads, downloads, and deletions through a clean REST API and web interface. 

You can specify the MaxUploadSize, WorkDir, ConfigFilePath, ShutdownTimeout, ReadTimeout, and WriteTimeout through configuration files or command-line arguments

## Features ✨

- **File Upload**: Upload single or multiple files via web interface or API 📤
- **File Download**: Download files directly from the browser 📥  
- **File Management**: Delete unwanted files easily 🗑️
- **Directory Browsing**: Navigate through directories with a user-friendly interface 📂
- **Cross-Platform**: Works on Windows, macOS, and Linux 🌍
- **Lightweight**: Minimal dependencies and fast performance ⚡

## Installation 🚀

### Prerequisites
- Go 1.18 or higher installed on your system

### From Source
```bash
# Clone the repository
git clone https://github.com/ldddi/http-file-server.git
cd http-file-server\cmd

# Build the application
go build -o http-file-server
./main.exe
```
