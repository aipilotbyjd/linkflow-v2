package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dop251/goja"
)

// Sandbox provides a secure JavaScript execution environment
type Sandbox struct {
	memoryLimit int64
	timeLimit   time.Duration
	vmPool      *VMPool
}

// SandboxConfig configures the sandbox
type SandboxConfig struct {
	MemoryLimit    int64         // Max memory in bytes (default: 50MB)
	TimeLimit      time.Duration // Max execution time (default: 30s)
	MaxVMs         int           // Max VMs in pool (default: 10)
	EnableConsole  bool          // Enable console.log
	AllowedGlobals []string      // Additional allowed globals
}

// DefaultSandboxConfig returns default sandbox configuration
func DefaultSandboxConfig() SandboxConfig {
	return SandboxConfig{
		MemoryLimit:   50 * 1024 * 1024, // 50MB
		TimeLimit:     30 * time.Second,
		MaxVMs:        10,
		EnableConsole: true,
	}
}

// NewSandbox creates a new sandbox
func NewSandbox(cfg SandboxConfig) *Sandbox {
	if cfg.MemoryLimit == 0 {
		cfg.MemoryLimit = 50 * 1024 * 1024
	}
	if cfg.TimeLimit == 0 {
		cfg.TimeLimit = 30 * time.Second
	}
	if cfg.MaxVMs == 0 {
		cfg.MaxVMs = 10
	}

	return &Sandbox{
		memoryLimit: cfg.MemoryLimit,
		timeLimit:   cfg.TimeLimit,
		vmPool:      NewVMPool(cfg.MaxVMs, cfg.EnableConsole),
	}
}

// Execute runs JavaScript code in the sandbox
func (s *Sandbox) Execute(ctx context.Context, code string, input map[string]interface{}) (interface{}, error) {
	vm := s.vmPool.Get()
	defer s.vmPool.Put(vm)

	// Set up timeout
	timer := time.AfterFunc(s.timeLimit, func() {
		vm.Interrupt("execution timeout exceeded")
	})
	defer timer.Stop()

	// Inject input
	if err := vm.Set("input", input); err != nil {
		return nil, fmt.Errorf("failed to set input: %w", err)
	}

	// Inject $json for convenience
	if json, ok := input["$json"]; ok {
		_ = vm.Set("$json", json)
	}

	// Inject helper functions
	injectHelpers(vm)

	// Execute with panic recovery
	var result interface{}
	var execErr error

	done := make(chan struct{})
	go func() {
		defer close(done)
		defer func() {
			if r := recover(); r != nil {
				execErr = fmt.Errorf("sandbox panic: %v", r)
			}
		}()

		val, err := vm.RunString(code)
		if err != nil {
			execErr = err
			return
		}

		result = exportValue(val)
	}()

	select {
	case <-ctx.Done():
		vm.Interrupt("context cancelled")
		return nil, ctx.Err()
	case <-done:
		return result, execErr
	}
}

// ExecuteFunction runs a JavaScript function with arguments
func (s *Sandbox) ExecuteFunction(ctx context.Context, code string, funcName string, args ...interface{}) (interface{}, error) {
	vm := s.vmPool.Get()
	defer s.vmPool.Put(vm)

	timer := time.AfterFunc(s.timeLimit, func() {
		vm.Interrupt("execution timeout exceeded")
	})
	defer timer.Stop()

	// Run the code to define functions
	_, err := vm.RunString(code)
	if err != nil {
		return nil, fmt.Errorf("failed to run code: %w", err)
	}

	// Get the function
	fn, ok := goja.AssertFunction(vm.Get(funcName))
	if !ok {
		return nil, fmt.Errorf("function '%s' not found", funcName)
	}

	// Convert arguments
	gojaArgs := make([]goja.Value, len(args))
	for i, arg := range args {
		gojaArgs[i] = vm.ToValue(arg)
	}

	// Call the function
	result, err := fn(goja.Undefined(), gojaArgs...)
	if err != nil {
		return nil, err
	}

	return exportValue(result), nil
}

// VMPool manages a pool of JavaScript VMs
type VMPool struct {
	pool          chan *goja.Runtime
	enableConsole bool
}

// NewVMPool creates a new VM pool
func NewVMPool(size int, enableConsole bool) *VMPool {
	pool := &VMPool{
		pool:          make(chan *goja.Runtime, size),
		enableConsole: enableConsole,
	}

	// Pre-create VMs
	for i := 0; i < size; i++ {
		pool.pool <- pool.createVM()
	}

	return pool
}

func (p *VMPool) createVM() *goja.Runtime {
	vm := goja.New()
	vm.SetFieldNameMapper(goja.UncapFieldNameMapper())

	// Remove dangerous globals
	_ = vm.Set("eval", goja.Undefined())
	_ = vm.Set("Function", goja.Undefined())

	// Add console if enabled
	if p.enableConsole {
		console := vm.NewObject()
		_ = console.Set("log", func(call goja.FunctionCall) goja.Value {
			// In production, you might want to capture these logs
			return goja.Undefined()
		})
		_ = console.Set("error", func(call goja.FunctionCall) goja.Value {
			return goja.Undefined()
		})
		_ = console.Set("warn", func(call goja.FunctionCall) goja.Value {
			return goja.Undefined()
		})
		_ = console.Set("info", func(call goja.FunctionCall) goja.Value {
			return goja.Undefined()
		})
		_ = vm.Set("console", console)
	}

	return vm
}

// Get retrieves a VM from the pool
func (p *VMPool) Get() *goja.Runtime {
	select {
	case vm := <-p.pool:
		return vm
	default:
		return p.createVM()
	}
}

// Put returns a VM to the pool
func (p *VMPool) Put(vm *goja.Runtime) {
	// Clear interrupts
	vm.ClearInterrupt()

	// Try to return to pool, or discard if full
	select {
	case p.pool <- vm:
	default:
		// Pool is full, discard VM
	}
}

func injectHelpers(vm *goja.Runtime) {
	// JSON helpers
	_ = vm.Set("JSON", map[string]interface{}{
		"parse": func(s string) interface{} {
			var v interface{}
			_ = json.Unmarshal([]byte(s), &v)
			return v
		},
		"stringify": func(v interface{}) string {
			b, _ := json.Marshal(v)
			return string(b)
		},
	})

	// Array helpers
	_ = vm.Set("Array", map[string]interface{}{
		"isArray": func(v interface{}) bool {
			_, ok := v.([]interface{})
			return ok
		},
	})

	// Object helpers
	_ = vm.Set("Object", map[string]interface{}{
		"keys": func(obj map[string]interface{}) []string {
			keys := make([]string, 0, len(obj))
			for k := range obj {
				keys = append(keys, k)
			}
			return keys
		},
		"values": func(obj map[string]interface{}) []interface{} {
			values := make([]interface{}, 0, len(obj))
			for _, v := range obj {
				values = append(values, v)
			}
			return values
		},
		"entries": func(obj map[string]interface{}) [][]interface{} {
			entries := make([][]interface{}, 0, len(obj))
			for k, v := range obj {
				entries = append(entries, []interface{}{k, v})
			}
			return entries
		},
		"assign": func(target map[string]interface{}, sources ...map[string]interface{}) map[string]interface{} {
			for _, src := range sources {
				for k, v := range src {
					target[k] = v
				}
			}
			return target
		},
	})

	// Math helpers (subset)
	_ = vm.Set("Math", map[string]interface{}{
		"round": func(x float64) float64 { return float64(int(x + 0.5)) },
		"floor": func(x float64) float64 { return float64(int(x)) },
		"ceil": func(x float64) float64 {
			i := int(x)
			if float64(i) < x {
				return float64(i + 1)
			}
			return float64(i)
		},
		"abs": func(x float64) float64 {
			if x < 0 {
				return -x
			}
			return x
		},
		"max": func(args ...float64) float64 {
			if len(args) == 0 {
				return 0
			}
			max := args[0]
			for _, v := range args[1:] {
				if v > max {
					max = v
				}
			}
			return max
		},
		"min": func(args ...float64) float64 {
			if len(args) == 0 {
				return 0
			}
			min := args[0]
			for _, v := range args[1:] {
				if v < min {
					min = v
				}
			}
			return min
		},
		"random": func() float64 {
			return float64(time.Now().UnixNano()%1000) / 1000
		},
	})
}

func exportValue(val goja.Value) interface{} {
	if val == nil || goja.IsUndefined(val) || goja.IsNull(val) {
		return nil
	}
	return val.Export()
}

// CodeExecutor is a higher-level code execution interface
type CodeExecutor struct {
	sandbox *Sandbox
}

// NewCodeExecutor creates a new code executor
func NewCodeExecutor(cfg SandboxConfig) *CodeExecutor {
	return &CodeExecutor{
		sandbox: NewSandbox(cfg),
	}
}

// Execute executes code with the given input
func (ce *CodeExecutor) Execute(ctx context.Context, code string, input map[string]interface{}) (map[string]interface{}, error) {
	result, err := ce.sandbox.Execute(ctx, code, input)
	if err != nil {
		return nil, err
	}

	// Ensure result is a map
	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}

	// Wrap non-map results
	return map[string]interface{}{
		"result": result,
	}, nil
}

// ExecuteTransform executes a transformation function on items
func (ce *CodeExecutor) ExecuteTransform(ctx context.Context, code string, items []interface{}) ([]interface{}, error) {
	results := make([]interface{}, len(items))

	for i, item := range items {
		input := map[string]interface{}{
			"$json": item,
			"$item": item,
			"$index": i,
		}

		result, err := ce.sandbox.Execute(ctx, code, input)
		if err != nil {
			return nil, fmt.Errorf("transform failed at index %d: %w", i, err)
		}

		results[i] = result
	}

	return results, nil
}

// ExecuteFilter executes a filter function on items
func (ce *CodeExecutor) ExecuteFilter(ctx context.Context, code string, items []interface{}) ([]interface{}, error) {
	var results []interface{}

	for i, item := range items {
		input := map[string]interface{}{
			"$json": item,
			"$item": item,
			"$index": i,
		}

		result, err := ce.sandbox.Execute(ctx, code, input)
		if err != nil {
			return nil, fmt.Errorf("filter failed at index %d: %w", i, err)
		}

		// Check if item passes filter
		if pass, ok := result.(bool); ok && pass {
			results = append(results, item)
		}
	}

	return results, nil
}
