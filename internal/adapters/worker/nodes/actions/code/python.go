package code

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// PythonNode executes Python code
type PythonNode struct{}

// NewPythonNode creates a new Python code node
func NewPythonNode() *PythonNode {
	return &PythonNode{}
}

// Execute runs Python code and returns the result
func (n *PythonNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})
	inputData := runtime.GetInputData()

	code, _ := params["code"].(string)
	if code == "" {
		return nil, fmt.Errorf("code is required")
	}

	// Get timeout (default 30 seconds)
	timeoutSec := 30
	if t, ok := params["timeout"].(float64); ok && t > 0 {
		timeoutSec = int(t)
	}

	// Prepare input data as JSON
	inputJSON, err := json.Marshal(inputData)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize input: %w", err)
	}

	// Create wrapper script
	wrapperCode := fmt.Sprintf(`
import sys
import json

# Input data from workflow
input_data = json.loads('''%s''')

# User code
def main(data):
%s

# Execute and output result
try:
    result = main(input_data)
    print(json.dumps({"success": True, "output": result}))
except Exception as e:
    print(json.dumps({"success": False, "error": str(e), "type": type(e).__name__}))
`, string(inputJSON), indentCode(code, "    "))

	// Create temp file
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, fmt.Sprintf("workflow_python_%d.py", time.Now().UnixNano()))
	if err := os.WriteFile(tmpFile, []byte(wrapperCode), 0600); err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile)

	// Execute with timeout
	execCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSec)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(execCtx, "python3", tmpFile)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	startTime := time.Now()
	err = cmd.Run()
	duration := time.Since(startTime)

	if execCtx.Err() == context.DeadlineExceeded {
		return nil, fmt.Errorf("execution timed out after %d seconds", timeoutSec)
	}

	if err != nil {
		return types.JSON{
			"success":     false,
			"error":       stderr.String(),
			"exit_code":   cmd.ProcessState.ExitCode(),
			"duration_ms": duration.Milliseconds(),
		}, nil
	}

	// Parse output
	var result map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return types.JSON{
			"success":     true,
			"output":      stdout.String(),
			"duration_ms": duration.Milliseconds(),
		}, nil
	}

	result["duration_ms"] = duration.Milliseconds()
	return result, nil
}

// Metadata returns node metadata
func (n *PythonNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "action.python",
		Name:        "Python Code",
		Description: "Execute Python code with full standard library access",
		Category:    "action",
		Version:     "1.0.0",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "any"}, {Name: "error", Type: "error"}},
		Parameters: []wtypes.NodeParameter{
			{Name: "code", Type: "code", Description: "Python code to execute. Receives 'data' parameter with input data.", Required: true},
			{Name: "timeout", Type: "number", Description: "Execution timeout in seconds", Default: 30},
		},
	}
}

// TypeScriptNode executes TypeScript code
type TypeScriptNode struct{}

// NewTypeScriptNode creates a new TypeScript code node
func NewTypeScriptNode() *TypeScriptNode {
	return &TypeScriptNode{}
}

// Execute runs TypeScript code using Deno or ts-node
func (n *TypeScriptNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})
	inputData := runtime.GetInputData()

	code, _ := params["code"].(string)
	if code == "" {
		return nil, fmt.Errorf("code is required")
	}

	timeoutSec := 30
	if t, ok := params["timeout"].(float64); ok && t > 0 {
		timeoutSec = int(t)
	}

	inputJSON, err := json.Marshal(inputData)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize input: %w", err)
	}

	// Create TypeScript wrapper
	wrapperCode := fmt.Sprintf(`
const inputData = %s;

async function main(data: any): Promise<any> {
%s
}

(async () => {
  try {
    const result = await main(inputData);
    console.log(JSON.stringify({ success: true, output: result }));
  } catch (e: any) {
    console.log(JSON.stringify({ success: false, error: e.message, type: e.name }));
  }
})();
`, string(inputJSON), indentCode(code, "  "))

	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, fmt.Sprintf("workflow_ts_%d.ts", time.Now().UnixNano()))
	if err := os.WriteFile(tmpFile, []byte(wrapperCode), 0600); err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile)

	execCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSec)*time.Second)
	defer cancel()

	// Try Deno first, fall back to ts-node
	var cmd *exec.Cmd
	if _, err := exec.LookPath("deno"); err == nil {
		cmd = exec.CommandContext(execCtx, "deno", "run", "--allow-net", "--allow-env", tmpFile)
	} else {
		cmd = exec.CommandContext(execCtx, "npx", "ts-node", tmpFile)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	startTime := time.Now()
	err = cmd.Run()
	duration := time.Since(startTime)

	if execCtx.Err() == context.DeadlineExceeded {
		return nil, fmt.Errorf("execution timed out after %d seconds", timeoutSec)
	}

	if err != nil {
		return types.JSON{
			"success":     false,
			"error":       stderr.String(),
			"duration_ms": duration.Milliseconds(),
		}, nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return types.JSON{
			"success":     true,
			"output":      stdout.String(),
			"duration_ms": duration.Milliseconds(),
		}, nil
	}

	result["duration_ms"] = duration.Milliseconds()
	return result, nil
}

// Metadata returns node metadata
func (n *TypeScriptNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "action.typescript",
		Name:        "TypeScript Code",
		Description: "Execute TypeScript code with type safety",
		Category:    "action",
		Version:     "1.0.0",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "any"}, {Name: "error", Type: "error"}},
		Parameters: []wtypes.NodeParameter{
			{Name: "code", Type: "code", Description: "TypeScript code to execute. Receives 'data' parameter with input data.", Required: true},
			{Name: "timeout", Type: "number", Description: "Execution timeout in seconds", Default: 30},
		},
	}
}

func indentCode(code, indent string) string {
	var result bytes.Buffer
	for _, line := range bytes.Split([]byte(code), []byte("\n")) {
		result.WriteString(indent)
		result.Write(line)
		result.WriteString("\n")
	}
	return result.String()
}
