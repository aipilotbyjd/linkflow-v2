package integrations

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/jlaffaye/ftp"
	"github.com/linkflow-ai/linkflow/internal/worker/core"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
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

	client, sftpClient, err := n.connect(config)
	if err != nil {
		return nil, fmt.Errorf("SFTP connection failed: %w", err)
	}
	defer sftpClient.Close()
	defer client.Close()

	switch operation {
	case "list":
		return n.list(sftpClient, config)
	case "download":
		return n.download(sftpClient, config)
	case "upload":
		return n.upload(sftpClient, config, execCtx.Input)
	case "delete":
		return n.delete(sftpClient, config)
	case "rename":
		return n.rename(sftpClient, config)
	case "mkdir":
		return n.mkdir(sftpClient, config)
	case "rmdir":
		return n.rmdir(sftpClient, config)
	case "stat":
		return n.stat(sftpClient, config)
	case "chmod":
		return n.chmod(sftpClient, config)
	default:
		return n.list(sftpClient, config)
	}
}

func (n *SFTPNode) connect(config map[string]interface{}) (*ssh.Client, *sftp.Client, error) {
	host := core.GetString(config, "host", "")
	port := core.GetInt(config, "port", 22)
	username := core.GetString(config, "username", "")
	password := core.GetString(config, "password", "")
	privateKey := core.GetString(config, "privateKey", "")
	timeout := core.GetInt(config, "timeout", 30)

	if host == "" {
		return nil, nil, fmt.Errorf("host is required")
	}
	if username == "" {
		return nil, nil, fmt.Errorf("username is required")
	}

	var authMethods []ssh.AuthMethod

	if privateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(privateKey))
		if err != nil {
			return nil, nil, fmt.Errorf("invalid private key: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	if password != "" {
		authMethods = append(authMethods, ssh.Password(password))
	}

	if len(authMethods) == 0 {
		return nil, nil, fmt.Errorf("password or privateKey is required")
	}

	sshConfig := &ssh.ClientConfig{
		User:            username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Duration(timeout) * time.Second,
	}

	address := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	client, err := ssh.Dial("tcp", address, sshConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("SSH connection failed: %w", err)
	}

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		client.Close()
		return nil, nil, fmt.Errorf("SFTP client creation failed: %w", err)
	}

	return client, sftpClient, nil
}

func (n *SFTPNode) list(client *sftp.Client, config map[string]interface{}) (map[string]interface{}, error) {
	path := core.GetString(config, "path", "/")

	entries, err := client.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("list failed: %w", err)
	}

	var files []map[string]interface{}
	for _, entry := range entries {
		files = append(files, map[string]interface{}{
			"name":        entry.Name(),
			"size":        entry.Size(),
			"mode":        entry.Mode().String(),
			"modTime":     entry.ModTime().Format(time.RFC3339),
			"path":        filepath.Join(path, entry.Name()),
			"isDir":       entry.IsDir(),
			"permissions": fmt.Sprintf("%04o", entry.Mode().Perm()),
		})
	}

	return map[string]interface{}{
		"files": files,
		"count": len(files),
		"path":  path,
	}, nil
}

func (n *SFTPNode) download(client *sftp.Client, config map[string]interface{}) (map[string]interface{}, error) {
	path := core.GetString(config, "path", "")
	if path == "" {
		return nil, fmt.Errorf("path is required")
	}

	file, err := client.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open failed: %w", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat failed: %w", err)
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("read failed: %w", err)
	}

	return map[string]interface{}{
		"content":  string(data),
		"size":     len(data),
		"path":     path,
		"filename": filepath.Base(path),
		"mode":     stat.Mode().String(),
		"modTime":  stat.ModTime().Format(time.RFC3339),
	}, nil
}

func (n *SFTPNode) upload(client *sftp.Client, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
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

	file, err := client.Create(path)
	if err != nil {
		return nil, fmt.Errorf("create failed: %w", err)
	}
	defer file.Close()

	written, err := file.Write([]byte(content))
	if err != nil {
		return nil, fmt.Errorf("write failed: %w", err)
	}

	return map[string]interface{}{
		"uploaded": true,
		"path":     path,
		"size":     written,
		"filename": filepath.Base(path),
	}, nil
}

func (n *SFTPNode) delete(client *sftp.Client, config map[string]interface{}) (map[string]interface{}, error) {
	path := core.GetString(config, "path", "")
	if path == "" {
		return nil, fmt.Errorf("path is required")
	}

	err := client.Remove(path)
	if err != nil {
		return nil, fmt.Errorf("delete failed: %w", err)
	}

	return map[string]interface{}{
		"deleted":  true,
		"path":     path,
		"filename": filepath.Base(path),
	}, nil
}

func (n *SFTPNode) rename(client *sftp.Client, config map[string]interface{}) (map[string]interface{}, error) {
	oldPath := core.GetString(config, "oldPath", "")
	newPath := core.GetString(config, "newPath", "")

	if oldPath == "" || newPath == "" {
		return nil, fmt.Errorf("oldPath and newPath are required")
	}

	err := client.Rename(oldPath, newPath)
	if err != nil {
		return nil, fmt.Errorf("rename failed: %w", err)
	}

	return map[string]interface{}{
		"renamed": true,
		"oldPath": oldPath,
		"newPath": newPath,
	}, nil
}

func (n *SFTPNode) mkdir(client *sftp.Client, config map[string]interface{}) (map[string]interface{}, error) {
	path := core.GetString(config, "path", "")
	if path == "" {
		return nil, fmt.Errorf("path is required")
	}

	err := client.MkdirAll(path)
	if err != nil {
		return nil, fmt.Errorf("mkdir failed: %w", err)
	}

	return map[string]interface{}{
		"created": true,
		"path":    path,
	}, nil
}

func (n *SFTPNode) rmdir(client *sftp.Client, config map[string]interface{}) (map[string]interface{}, error) {
	path := core.GetString(config, "path", "")
	if path == "" {
		return nil, fmt.Errorf("path is required")
	}

	err := client.RemoveDirectory(path)
	if err != nil {
		return nil, fmt.Errorf("rmdir failed: %w", err)
	}

	return map[string]interface{}{
		"removed": true,
		"path":    path,
	}, nil
}

func (n *SFTPNode) stat(client *sftp.Client, config map[string]interface{}) (map[string]interface{}, error) {
	path := core.GetString(config, "path", "")
	if path == "" {
		return nil, fmt.Errorf("path is required")
	}

	info, err := client.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat failed: %w", err)
	}

	return map[string]interface{}{
		"name":        info.Name(),
		"size":        info.Size(),
		"mode":        info.Mode().String(),
		"modTime":     info.ModTime().Format(time.RFC3339),
		"isDir":       info.IsDir(),
		"permissions": fmt.Sprintf("%04o", info.Mode().Perm()),
		"path":        path,
	}, nil
}

func (n *SFTPNode) chmod(client *sftp.Client, config map[string]interface{}) (map[string]interface{}, error) {
	path := core.GetString(config, "path", "")
	mode := core.GetInt(config, "mode", 0)

	if path == "" {
		return nil, fmt.Errorf("path is required")
	}
	if mode == 0 {
		return nil, fmt.Errorf("mode is required (e.g., 755, 644)")
	}

	err := client.Chmod(path, os.FileMode(mode))
	if err != nil {
		return nil, fmt.Errorf("chmod failed: %w", err)
	}

	return map[string]interface{}{
		"changed": true,
		"path":    path,
		"mode":    fmt.Sprintf("%04o", mode),
	}, nil
}
