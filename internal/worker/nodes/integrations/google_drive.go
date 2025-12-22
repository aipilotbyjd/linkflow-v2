package integrations

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/linkflow-ai/linkflow/internal/worker/core"
)

// GoogleDriveNode handles Google Drive operations
type GoogleDriveNode struct{}

func (n *GoogleDriveNode) Type() string {
	return "integrations.google_drive"
}

func (n *GoogleDriveNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	config := execCtx.Config

	accessToken := core.GetString(config, "accessToken", "")
	if accessToken == "" {
		return nil, fmt.Errorf("accessToken is required")
	}

	operation := core.GetString(config, "operation", "list")

	switch operation {
	case "list":
		return n.listFiles(ctx, config, accessToken)
	case "get":
		return n.getFile(ctx, config, accessToken)
	case "download":
		return n.downloadFile(ctx, config, accessToken)
	case "upload":
		return n.uploadFile(ctx, config, accessToken, execCtx.Input)
	case "create":
		return n.createFile(ctx, config, accessToken, execCtx.Input)
	case "update":
		return n.updateFile(ctx, config, accessToken, execCtx.Input)
	case "delete":
		return n.deleteFile(ctx, config, accessToken)
	case "copy":
		return n.copyFile(ctx, config, accessToken)
	case "move":
		return n.moveFile(ctx, config, accessToken)
	case "createFolder":
		return n.createFolder(ctx, config, accessToken)
	case "search":
		return n.searchFiles(ctx, config, accessToken)
	case "share":
		return n.shareFile(ctx, config, accessToken)
	default:
		return n.listFiles(ctx, config, accessToken)
	}
}

func (n *GoogleDriveNode) listFiles(ctx context.Context, config map[string]interface{}, accessToken string) (map[string]interface{}, error) {
	endpoint := "https://www.googleapis.com/drive/v3/files"

	params := url.Values{}
	params.Set("fields", "files(id,name,mimeType,size,createdTime,modifiedTime,parents,webViewLink)")

	if folderId := core.GetString(config, "folderId", ""); folderId != "" {
		params.Set("q", fmt.Sprintf("'%s' in parents", folderId))
	}
	if pageSize := core.GetInt(config, "pageSize", 100); pageSize > 0 {
		params.Set("pageSize", fmt.Sprintf("%d", pageSize))
	}
	if pageToken := core.GetString(config, "pageToken", ""); pageToken != "" {
		params.Set("pageToken", pageToken)
	}

	result, err := n.makeRequest(ctx, "GET", endpoint+"?"+params.Encode(), nil, accessToken)
	if err != nil {
		return nil, err
	}

	files := result["files"]
	if files == nil {
		files = []interface{}{}
	}

	return map[string]interface{}{
		"files":         files,
		"nextPageToken": result["nextPageToken"],
	}, nil
}

func (n *GoogleDriveNode) getFile(ctx context.Context, config map[string]interface{}, accessToken string) (map[string]interface{}, error) {
	fileId := core.GetString(config, "fileId", "")
	if fileId == "" {
		return nil, fmt.Errorf("fileId is required")
	}

	endpoint := fmt.Sprintf("https://www.googleapis.com/drive/v3/files/%s", fileId)
	params := url.Values{}
	params.Set("fields", "id,name,mimeType,size,createdTime,modifiedTime,parents,webViewLink,webContentLink,description")

	result, err := n.makeRequest(ctx, "GET", endpoint+"?"+params.Encode(), nil, accessToken)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (n *GoogleDriveNode) downloadFile(ctx context.Context, config map[string]interface{}, accessToken string) (map[string]interface{}, error) {
	fileId := core.GetString(config, "fileId", "")
	if fileId == "" {
		return nil, fmt.Errorf("fileId is required")
	}

	// First get file metadata
	metaEndpoint := fmt.Sprintf("https://www.googleapis.com/drive/v3/files/%s?fields=name,mimeType,size", fileId)
	meta, err := n.makeRequest(ctx, "GET", metaEndpoint, nil, accessToken)
	if err != nil {
		return nil, err
	}

	// Download content
	endpoint := fmt.Sprintf("https://www.googleapis.com/drive/v3/files/%s?alt=media", fileId)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read content: %w", err)
	}

	// Return as string or base64
	content := string(data)
	isBase64 := false
	mimeType := ""
	if mt, ok := meta["mimeType"].(string); ok {
		mimeType = mt
		if !strings.HasPrefix(mimeType, "text/") && mimeType != "application/json" {
			content = base64.StdEncoding.EncodeToString(data)
			isBase64 = true
		}
	}

	return map[string]interface{}{
		"content":  content,
		"isBase64": isBase64,
		"name":     meta["name"],
		"mimeType": mimeType,
		"size":     len(data),
	}, nil
}

func (n *GoogleDriveNode) uploadFile(ctx context.Context, config map[string]interface{}, accessToken string, input map[string]interface{}) (map[string]interface{}, error) {
	name := core.GetString(config, "name", "")
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	content := core.GetString(config, "content", "")
	if content == "" {
		if c, ok := input["content"].(string); ok {
			content = c
		}
	}

	mimeType := core.GetString(config, "mimeType", "text/plain")
	folderId := core.GetString(config, "folderId", "")

	// Decode base64 if needed
	var data []byte
	if isBase64 := core.GetBool(config, "isBase64", false); isBase64 {
		var err error
		data, err = base64.StdEncoding.DecodeString(content)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64: %w", err)
		}
	} else {
		data = []byte(content)
	}

	// Create multipart upload
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Metadata part
	metadata := map[string]interface{}{
		"name":     name,
		"mimeType": mimeType,
	}
	if folderId != "" {
		metadata["parents"] = []string{folderId}
	}

	metadataJSON, _ := json.Marshal(metadata)
	metaPart, _ := writer.CreatePart(map[string][]string{
		"Content-Type": {"application/json; charset=UTF-8"},
	})
	_, _ = metaPart.Write(metadataJSON)

	// Content part
	contentPart, _ := writer.CreatePart(map[string][]string{
		"Content-Type": {mimeType},
	})
	_, _ = contentPart.Write(data)
	_ = writer.Close()

	endpoint := "https://www.googleapis.com/upload/drive/v3/files?uploadType=multipart&fields=id,name,mimeType,size,webViewLink"

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, &body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("upload failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	_ = json.Unmarshal(respBody, &result)

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("upload failed: %v", result)
	}

	return map[string]interface{}{
		"uploaded": true,
		"id":       result["id"],
		"name":     result["name"],
		"mimeType": result["mimeType"],
		"size":     result["size"],
		"link":     result["webViewLink"],
	}, nil
}

func (n *GoogleDriveNode) createFile(ctx context.Context, config map[string]interface{}, accessToken string, input map[string]interface{}) (map[string]interface{}, error) {
	return n.uploadFile(ctx, config, accessToken, input)
}

func (n *GoogleDriveNode) updateFile(ctx context.Context, config map[string]interface{}, accessToken string, input map[string]interface{}) (map[string]interface{}, error) {
	fileId := core.GetString(config, "fileId", "")
	if fileId == "" {
		return nil, fmt.Errorf("fileId is required")
	}

	content := core.GetString(config, "content", "")
	if content == "" {
		if c, ok := input["content"].(string); ok {
			content = c
		}
	}

	mimeType := core.GetString(config, "mimeType", "text/plain")

	var data []byte
	if isBase64 := core.GetBool(config, "isBase64", false); isBase64 {
		var err error
		data, err = base64.StdEncoding.DecodeString(content)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64: %w", err)
		}
	} else {
		data = []byte(content)
	}

	endpoint := fmt.Sprintf("https://www.googleapis.com/upload/drive/v3/files/%s?uploadType=media&fields=id,name,mimeType,size,modifiedTime", fileId)

	req, err := http.NewRequestWithContext(ctx, "PATCH", endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", mimeType)

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("update failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	_ = json.Unmarshal(respBody, &result)

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("update failed: %v", result)
	}

	return map[string]interface{}{
		"updated":      true,
		"id":           result["id"],
		"name":         result["name"],
		"modifiedTime": result["modifiedTime"],
	}, nil
}

func (n *GoogleDriveNode) deleteFile(ctx context.Context, config map[string]interface{}, accessToken string) (map[string]interface{}, error) {
	fileId := core.GetString(config, "fileId", "")
	if fileId == "" {
		return nil, fmt.Errorf("fileId is required")
	}

	endpoint := fmt.Sprintf("https://www.googleapis.com/drive/v3/files/%s", fileId)

	req, err := http.NewRequestWithContext(ctx, "DELETE", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("delete failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("delete failed: %s", string(body))
	}

	return map[string]interface{}{
		"deleted": true,
		"fileId":  fileId,
	}, nil
}

func (n *GoogleDriveNode) copyFile(ctx context.Context, config map[string]interface{}, accessToken string) (map[string]interface{}, error) {
	fileId := core.GetString(config, "fileId", "")
	if fileId == "" {
		return nil, fmt.Errorf("fileId is required")
	}

	endpoint := fmt.Sprintf("https://www.googleapis.com/drive/v3/files/%s/copy?fields=id,name,mimeType,webViewLink", fileId)

	body := map[string]interface{}{}
	if name := core.GetString(config, "name", ""); name != "" {
		body["name"] = name
	}
	if folderId := core.GetString(config, "folderId", ""); folderId != "" {
		body["parents"] = []string{folderId}
	}

	bodyJSON, _ := json.Marshal(body)
	result, err := n.makeRequestWithBody(ctx, "POST", endpoint, bodyJSON, accessToken)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"copied": true,
		"id":     result["id"],
		"name":   result["name"],
		"link":   result["webViewLink"],
	}, nil
}

func (n *GoogleDriveNode) moveFile(ctx context.Context, config map[string]interface{}, accessToken string) (map[string]interface{}, error) {
	fileId := core.GetString(config, "fileId", "")
	folderId := core.GetString(config, "folderId", "")

	if fileId == "" || folderId == "" {
		return nil, fmt.Errorf("fileId and folderId are required")
	}

	// Get current parents
	metaEndpoint := fmt.Sprintf("https://www.googleapis.com/drive/v3/files/%s?fields=parents", fileId)
	meta, err := n.makeRequest(ctx, "GET", metaEndpoint, nil, accessToken)
	if err != nil {
		return nil, err
	}

	var removeParents string
	if parents, ok := meta["parents"].([]interface{}); ok && len(parents) > 0 {
		parentIds := make([]string, len(parents))
		for i, p := range parents {
			parentIds[i] = fmt.Sprintf("%v", p)
		}
		removeParents = strings.Join(parentIds, ",")
	}

	endpoint := fmt.Sprintf("https://www.googleapis.com/drive/v3/files/%s?addParents=%s&removeParents=%s&fields=id,name,parents",
		fileId, folderId, removeParents)

	result, err := n.makeRequestWithBody(ctx, "PATCH", endpoint, []byte("{}"), accessToken)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"moved":  true,
		"id":     result["id"],
		"name":   result["name"],
		"folder": folderId,
	}, nil
}

func (n *GoogleDriveNode) createFolder(ctx context.Context, config map[string]interface{}, accessToken string) (map[string]interface{}, error) {
	name := core.GetString(config, "name", "")
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	endpoint := "https://www.googleapis.com/drive/v3/files?fields=id,name,mimeType,webViewLink"

	body := map[string]interface{}{
		"name":     name,
		"mimeType": "application/vnd.google-apps.folder",
	}
	if parentId := core.GetString(config, "parentId", ""); parentId != "" {
		body["parents"] = []string{parentId}
	}

	bodyJSON, _ := json.Marshal(body)
	result, err := n.makeRequestWithBody(ctx, "POST", endpoint, bodyJSON, accessToken)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"created": true,
		"id":      result["id"],
		"name":    result["name"],
		"link":    result["webViewLink"],
	}, nil
}

func (n *GoogleDriveNode) searchFiles(ctx context.Context, config map[string]interface{}, accessToken string) (map[string]interface{}, error) {
	query := core.GetString(config, "query", "")
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}

	endpoint := "https://www.googleapis.com/drive/v3/files"
	params := url.Values{}
	params.Set("q", query)
	params.Set("fields", "files(id,name,mimeType,size,createdTime,modifiedTime,webViewLink)")

	if pageSize := core.GetInt(config, "pageSize", 100); pageSize > 0 {
		params.Set("pageSize", fmt.Sprintf("%d", pageSize))
	}

	result, err := n.makeRequest(ctx, "GET", endpoint+"?"+params.Encode(), nil, accessToken)
	if err != nil {
		return nil, err
	}

	files := result["files"]
	if files == nil {
		files = []interface{}{}
	}

	return map[string]interface{}{
		"files": files,
		"query": query,
	}, nil
}

func (n *GoogleDriveNode) shareFile(ctx context.Context, config map[string]interface{}, accessToken string) (map[string]interface{}, error) {
	fileId := core.GetString(config, "fileId", "")
	if fileId == "" {
		return nil, fmt.Errorf("fileId is required")
	}

	role := core.GetString(config, "role", "reader") // reader, writer, commenter
	shareType := core.GetString(config, "type", "anyone") // user, group, domain, anyone
	email := core.GetString(config, "email", "")

	endpoint := fmt.Sprintf("https://www.googleapis.com/drive/v3/files/%s/permissions", fileId)

	body := map[string]interface{}{
		"role": role,
		"type": shareType,
	}
	if shareType == "user" || shareType == "group" {
		if email == "" {
			return nil, fmt.Errorf("email is required for user/group sharing")
		}
		body["emailAddress"] = email
	}

	bodyJSON, _ := json.Marshal(body)
	result, err := n.makeRequestWithBody(ctx, "POST", endpoint, bodyJSON, accessToken)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"shared":       true,
		"permissionId": result["id"],
		"role":         role,
		"type":         shareType,
	}, nil
}

func (n *GoogleDriveNode) makeRequest(ctx context.Context, method, endpoint string, body []byte, accessToken string) (map[string]interface{}, error) {
	var req *http.Request
	var err error

	if body != nil {
		req, err = http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(body))
	} else {
		req, err = http.NewRequestWithContext(ctx, method, endpoint, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error: %v", result)
	}

	return result, nil
}

func (n *GoogleDriveNode) makeRequestWithBody(ctx context.Context, method, endpoint string, body []byte, accessToken string) (map[string]interface{}, error) {
	return n.makeRequest(ctx, method, endpoint, body, accessToken)
}

// Note: GoogleDriveNode is registered in integrations/init.go
