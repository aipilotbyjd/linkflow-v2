package transform

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// StringOperationNode performs string operations
type StringOperationNode struct{}

func NewStringOperationNode() *StringOperationNode {
	return &StringOperationNode{}
}

func (n *StringOperationNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})
	inputData := runtime.GetInputData()

	str, _ := params["string"].(string)
	if str == "" {
		if s, ok := inputData["data"].(string); ok {
			str = s
		}
	}

	operation, _ := params["operation"].(string)

	var result string
	switch operation {
	case "uppercase":
		result = strings.ToUpper(str)
	case "lowercase":
		result = strings.ToLower(str)
	case "trim":
		result = strings.TrimSpace(str)
	case "trim_left":
		result = strings.TrimLeft(str, " \t\n\r")
	case "trim_right":
		result = strings.TrimRight(str, " \t\n\r")
	case "reverse":
		runes := []rune(str)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		result = string(runes)
	case "base64_encode":
		result = base64.StdEncoding.EncodeToString([]byte(str))
	case "base64_decode":
		decoded, err := base64.StdEncoding.DecodeString(str)
		if err != nil {
			return types.JSON{"error": err.Error(), "success": false}, nil
		}
		result = string(decoded)
	case "sha256":
		hash := sha256.Sum256([]byte(str))
		result = hex.EncodeToString(hash[:])
	case "length":
		return types.JSON{"result": len(str), "success": true}, nil
	case "split":
		delimiter, _ := params["delimiter"].(string)
		if delimiter == "" {
			delimiter = ","
		}
		parts := strings.Split(str, delimiter)
		return types.JSON{"result": parts, "success": true}, nil
	case "replace":
		find, _ := params["find"].(string)
		replace, _ := params["replace"].(string)
		result = strings.ReplaceAll(str, find, replace)
	case "regex_replace":
		pattern, _ := params["pattern"].(string)
		replace, _ := params["replace"].(string)
		re, err := regexp.Compile(pattern)
		if err != nil {
			return types.JSON{"error": err.Error(), "success": false}, nil
		}
		result = re.ReplaceAllString(str, replace)
	case "regex_match":
		pattern, _ := params["pattern"].(string)
		re, err := regexp.Compile(pattern)
		if err != nil {
			return types.JSON{"error": err.Error(), "success": false}, nil
		}
		matches := re.FindAllString(str, -1)
		return types.JSON{"matches": matches, "success": true}, nil
	case "substring":
		start, _ := params["start"].(float64)
		end, _ := params["end"].(float64)
		if int(end) > len(str) || end == 0 {
			end = float64(len(str))
		}
		result = str[int(start):int(end)]
	case "pad_left":
		length, _ := params["length"].(float64)
		padChar, _ := params["pad_char"].(string)
		if padChar == "" {
			padChar = " "
		}
		for len(result) < int(length) {
			result = padChar + str
		}
	case "pad_right":
		length, _ := params["length"].(float64)
		padChar, _ := params["pad_char"].(string)
		if padChar == "" {
			padChar = " "
		}
		result = str
		for len(result) < int(length) {
			result = result + padChar
		}
	default:
		result = str
	}

	return types.JSON{"result": result, "success": true}, nil
}

func (n *StringOperationNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "transform.string",
		Name:        "String Operations",
		Description: "Perform string manipulation operations",
		Category:    "transform",
		Version:     "1.0.0",
		Icon:        "TextFont",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "string"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "any"}},
		Parameters: []wtypes.NodeParameter{
			{Name: "string", Type: "string", Description: "Input string"},
			{Name: "operation", Type: "options", Description: "Operation to perform", Required: true, Options: []wtypes.ParamOption{
				{Value: "uppercase", Name: "Uppercase"},
				{Value: "lowercase", Name: "Lowercase"},
				{Value: "trim", Name: "Trim"},
				{Value: "reverse", Name: "Reverse"},
				{Value: "base64_encode", Name: "Base64 Encode"},
				{Value: "base64_decode", Name: "Base64 Decode"},
				{Value: "md5", Name: "MD5 Hash"},
				{Value: "sha256", Name: "SHA256 Hash"},
				{Value: "length", Name: "Length"},
				{Value: "split", Name: "Split"},
				{Value: "replace", Name: "Replace"},
				{Value: "regex_replace", Name: "Regex Replace"},
				{Value: "regex_match", Name: "Regex Match"},
				{Value: "substring", Name: "Substring"},
			}},
			{Name: "delimiter", Type: "string", Description: "Delimiter for split"},
			{Name: "find", Type: "string", Description: "String to find"},
			{Name: "replace", Type: "string", Description: "Replacement string"},
			{Name: "pattern", Type: "string", Description: "Regex pattern"},
			{Name: "start", Type: "number", Description: "Start index"},
			{Name: "end", Type: "number", Description: "End index"},
		},
	}
}

// JoinNode joins array elements into a string
type JoinNode struct{}

func NewJoinNode() *JoinNode {
	return &JoinNode{}
}

func (n *JoinNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})
	inputData := runtime.GetInputData()

	items, _ := inputData["items"].([]interface{})
	if items == nil {
		if arr, ok := inputData["data"].([]interface{}); ok {
			items = arr
		}
	}

	delimiter, _ := params["delimiter"].(string)
	if delimiter == "" {
		delimiter = ","
	}

	var parts []string
	for _, item := range items {
		parts = append(parts, fmt.Sprintf("%v", item))
	}

	return types.JSON{"result": strings.Join(parts, delimiter), "success": true}, nil
}

func (n *JoinNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "transform.join",
		Name:        "Join",
		Description: "Join array elements into a string",
		Category:    "transform",
		Version:     "1.0.0",
		Icon:        "TextFont",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "array"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "string"}},
		Parameters: []wtypes.NodeParameter{
			{Name: "delimiter", Type: "string", Description: "Delimiter between elements", Default: ","},
		},
	}
}
