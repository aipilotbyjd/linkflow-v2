package integrations

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"path/filepath"
	"time"

	"github.com/jlaffaye/ftp"
	"github.com/linkflow-ai/linkflow/internal/worker/core"
)

// FTPNode handles FTP operations
type FTPNode struct{}

func (n *FTPNode) Type() string {
	return "integrations.ftp"
}

func (n *FTPNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	config := execCtx.Config

	operation := core.GetString(config, "operation", "list")

	conn, err := n.connect(config)
	if err != nil {
		return nil, fmt.Errorf("FTP connection failed: %w", err)
	}
	defer func() { _ = conn.Quit() }()

	switch operation {
	case "list":
		return n.list(conn, config)
	case "download":
		return n.download(conn, config)
	case "upload":
		return n.upload(conn, config, execCtx.Input)
	case "delete":
		return n.delete(conn, config)
	case "rename":
		return n.rename(conn, config)
	case "mkdir":
		return n.mkdir(conn, config)
	case "rmdir":
		return n.rmdir(conn, config)
	default:
		return n.list(conn, config)
	}
}

func (n *FTPNode) connect(config map[string]interface{}) (*ftp.ServerConn, error) {
	host := core.GetString(config, "host", "")
	port := core.GetInt(config, "port", 21)
	username := core.GetString(config, "username", "anonymous")
	password := core.GetString(config, "password", "")
	timeout := core.GetInt(config, "timeout", 30)

	if host == "" {
		return nil, fmt.Errorf("host is required")
	}

	address := fmt.Sprintf("%s:%d", host, port)

	conn, err := ftp.Dial(address,
		ftp.DialWithTimeout(time.Duration(timeout)*time.Second),
		ftp.DialWithDisabledEPSV(true),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	if err := conn.Login(username, password); err != nil {
		_ = conn.Quit()
		return nil, fmt.Errorf("login failed: %w", err)
	}

	return conn, nil
}

func (n *FTPNode) list(conn *ftp.ServerConn, config map[string]interface{}) (map[string]interface{}, error) {
	path := core.GetString(config, "path", "/")

	entries, err := conn.List(path)
	if err != nil {
		return nil, fmt.Errorf("list failed: %w", err)
	}

	var files []map[string]interface{}
	for _, entry := range entries {
		files = append(files, map[string]interface{}{
			"name":     entry.Name,
			"size":     entry.Size,
			"type":     entryTypeToString(entry.Type),
			"time":     entry.Time.Format(time.RFC3339),
			"path":     filepath.Join(path, entry.Name),
			"isDir":    entry.Type == ftp.EntryTypeFolder,
		})
	}

	return map[string]interface{}{
		"files": files,
		"count": len(files),
		"path":  path,
	}, nil
}

func (n *FTPNode) download(conn *ftp.ServerConn, config map[string]interface{}) (map[string]interface{}, error) {
	path := core.GetString(config, "path", "")
	if path == "" {
		return nil, fmt.Errorf("path is required")
	}

	resp, err := conn.Retr(path)
	if err != nil {
		return nil, fmt.Errorf("download failed: %w", err)
	}
	defer resp.Close()

	data, err := io.ReadAll(resp)
	if err != nil {
		return nil, fmt.Errorf("read failed: %w", err)
	}

	return map[string]interface{}{
		"content":  string(data),
		"size":     len(data),
		"path":     path,
		"filename": filepath.Base(path),
	}, nil
}

func (n *FTPNode) upload(conn *ftp.ServerConn, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	path := core.GetString(config, "path", "")
	if path == "" {
		return nil, fmt.Errorf("path is required")
	}

	content := core.GetString(config, "content", "")
	if content == "" {
		if c, ok := input["content"].(string); ok {
			content = c
		} else if c, ok := input["data"].(string); ok {
			content = c
		}
	}

	if content == "" {
		return nil, fmt.Errorf("content is required")
	}

	reader := bytes.NewReader([]byte(content))
	err := conn.Stor(path, reader)
	if err != nil {
		return nil, fmt.Errorf("upload failed: %w", err)
	}

	return map[string]interface{}{
		"uploaded": true,
		"path":     path,
		"size":     len(content),
		"filename": filepath.Base(path),
	}, nil
}

func (n *FTPNode) delete(conn *ftp.ServerConn, config map[string]interface{}) (map[string]interface{}, error) {
	path := core.GetString(config, "path", "")
	if path == "" {
		return nil, fmt.Errorf("path is required")
	}

	err := conn.Delete(path)
	if err != nil {
		return nil, fmt.Errorf("delete failed: %w", err)
	}

	return map[string]interface{}{
		"deleted":  true,
		"path":     path,
		"filename": filepath.Base(path),
	}, nil
}

func (n *FTPNode) rename(conn *ftp.ServerConn, config map[string]interface{}) (map[string]interface{}, error) {
	oldPath := core.GetString(config, "oldPath", "")
	newPath := core.GetString(config, "newPath", "")

	if oldPath == "" || newPath == "" {
		return nil, fmt.Errorf("oldPath and newPath are required")
	}

	err := conn.Rename(oldPath, newPath)
	if err != nil {
		return nil, fmt.Errorf("rename failed: %w", err)
	}

	return map[string]interface{}{
		"renamed": true,
		"oldPath": oldPath,
		"newPath": newPath,
	}, nil
}

func (n *FTPNode) mkdir(conn *ftp.ServerConn, config map[string]interface{}) (map[string]interface{}, error) {
	path := core.GetString(config, "path", "")
	if path == "" {
		return nil, fmt.Errorf("path is required")
	}

	err := conn.MakeDir(path)
	if err != nil {
		return nil, fmt.Errorf("mkdir failed: %w", err)
	}

	return map[string]interface{}{
		"created": true,
		"path":    path,
	}, nil
}

func (n *FTPNode) rmdir(conn *ftp.ServerConn, config map[string]interface{}) (map[string]interface{}, error) {
	path := core.GetString(config, "path", "")
	if path == "" {
		return nil, fmt.Errorf("path is required")
	}

	err := conn.RemoveDir(path)
	if err != nil {
		return nil, fmt.Errorf("rmdir failed: %w", err)
	}

	return map[string]interface{}{
		"removed": true,
		"path":    path,
	}, nil
}

func entryTypeToString(t ftp.EntryType) string {
	switch t {
	case ftp.EntryTypeFile:
		return "file"
	case ftp.EntryTypeFolder:
		return "directory"
	case ftp.EntryTypeLink:
		return "link"
	default:
		return "unknown"
	}
}

// SFTPNode handles SFTP operations (SSH File Transfer Protocol)
type SFTPNode struct{}

func (n *SFTPNode) Type() string {
	return "integrations.sftp"
}

func (n *SFTPNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	config := execCtx.Config

	operation := core.GetString(config, "operation", "list")
	host := core.GetString(config, "host", "")
	port := core.GetInt(config, "port", 22)
	username := core.GetString(config, "username", "")
	password := core.GetString(config, "password", "")

	if host == "" {
		return nil, fmt.Errorf("host is required")
	}
	if username == "" {
		return nil, fmt.Errorf("username is required")
	}

	// Note: Full SFTP implementation requires golang.org/x/crypto/ssh
	// This is a placeholder that shows the API
	// For production, you would use pkg/sftp or similar

	address := net.JoinHostPort(host, fmt.Sprintf("%d", port))

	// Placeholder response showing configuration accepted
	return map[string]interface{}{
		"operation": operation,
		"host":      host,
		"port":      port,
		"username":  username,
		"address":   address,
		"hasAuth":   password != "",
		"message":   "SFTP client configuration accepted. Full implementation requires golang.org/x/crypto/ssh and github.com/pkg/sftp packages.",
	}, nil
}

// Note: FTPNode and SFTPNode are registered in integrations/init.go
