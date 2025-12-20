package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
	"github.com/linkflow-ai/linkflow/internal/pkg/queue"
	"github.com/linkflow-ai/linkflow/internal/worker/nodes"
	"github.com/rs/zerolog/log"
)

type Executor struct {
	executionSvc  *services.ExecutionService
	credentialSvc *services.CredentialService
	workflowSvc   *services.WorkflowService
	registry      *nodes.Registry
}

func New(
	executionSvc *services.ExecutionService,
	credentialSvc *services.CredentialService,
	workflowSvc *services.WorkflowService,
) *Executor {
	return &Executor{
		executionSvc:  executionSvc,
		credentialSvc: credentialSvc,
		workflowSvc:   workflowSvc,
		registry:      nodes.NewRegistry(),
	}
}

func (e *Executor) Execute(ctx context.Context, payload queue.WorkflowExecutionPayload) error {
	// Create execution record
	execution, err := e.executionSvc.Create(ctx, services.CreateExecutionInput{
		WorkflowID:  payload.WorkflowID,
		WorkspaceID: payload.WorkspaceID,
		TriggeredBy: payload.TriggeredBy,
		TriggerType: payload.TriggerType,
		TriggerData: payload.TriggerData,
		InputData:   payload.InputData,
	})
	if err != nil {
		return fmt.Errorf("failed to create execution: %w", err)
	}

	log.Info().
		Str("execution_id", execution.ID.String()).
		Str("workflow_id", payload.WorkflowID.String()).
		Msg("Starting workflow execution")

	// Get workflow
	workflow, err := e.workflowSvc.GetByID(ctx, payload.WorkflowID)
	if err != nil {
		e.executionSvc.Fail(ctx, execution.ID, "Workflow not found", nil)
		return fmt.Errorf("workflow not found: %w", err)
	}

	// Start execution
	if err := e.executionSvc.Start(ctx, execution.ID); err != nil {
		return err
	}

	// Build DAG and execute
	result, err := e.executeWorkflow(ctx, execution.ID, workflow, payload.InputData)
	if err != nil {
		nodeID := extractErrorNodeID(err)
		e.executionSvc.Fail(ctx, execution.ID, err.Error(), nodeID)
		return err
	}

	// Complete execution
	if err := e.executionSvc.Complete(ctx, execution.ID, result); err != nil {
		return err
	}

	log.Info().
		Str("execution_id", execution.ID.String()).
		Msg("Workflow execution completed")

	return nil
}

func (e *Executor) executeWorkflow(ctx context.Context, executionID uuid.UUID, workflow *models.Workflow, input models.JSON) (models.JSON, error) {
	// Parse nodes from workflow
	nodesData, err := parseNodes(workflow.Nodes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse nodes: %w", err)
	}

	// Parse connections
	connections, err := parseConnections(workflow.Connections)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connections: %w", err)
	}

	// Build DAG
	dag := BuildDAG(nodesData, connections)

	// Get execution order
	order, err := dag.TopologicalSort()
	if err != nil {
		return nil, fmt.Errorf("failed to sort DAG: %w", err)
	}

	// Execution context
	execCtx := &ExecutionContext{
		ExecutionID: executionID,
		WorkflowID:  workflow.ID,
		Input:       input,
		Variables:   make(map[string]interface{}),
		NodeOutputs: make(map[string]interface{}),
	}

	// Execute nodes in order
	for _, nodeID := range order {
		node := dag.Nodes[nodeID]
		if node == nil {
			continue
		}

		if err := e.executeNode(ctx, execCtx, node); err != nil {
			return nil, &NodeExecutionError{NodeID: nodeID, Err: err}
		}
	}

	// Collect output from terminal nodes
	output := make(models.JSON)
	for nodeID, nodeOutput := range execCtx.NodeOutputs {
		output[nodeID] = nodeOutput
	}

	return output, nil
}

func (e *Executor) executeNode(ctx context.Context, execCtx *ExecutionContext, node *NodeData) error {
	log.Debug().
		Str("execution_id", execCtx.ExecutionID.String()).
		Str("node_id", node.ID).
		Str("node_type", node.Type).
		Msg("Executing node")

	// Create node execution record
	nodeExec, err := e.executionSvc.CreateNodeExecution(ctx, execCtx.ExecutionID, node.ID, node.Type, node.Name)
	if err != nil {
		return err
	}

	// Get node handler
	handler := e.registry.Get(node.Type)
	if handler == nil {
		e.executionSvc.FailNodeExecution(ctx, nodeExec.ID, "Unknown node type")
		return fmt.Errorf("unknown node type: %s", node.Type)
	}

	// Prepare input
	nodeInput := e.prepareNodeInput(execCtx, node)
	e.executionSvc.StartNodeExecution(ctx, nodeExec.ID, nodeInput)

	// Execute
	startTime := time.Now()
	result, err := handler.Execute(ctx, &nodes.ExecutionContext{
		ExecutionID: execCtx.ExecutionID,
		NodeID:      node.ID,
		Input:       nodeInput,
		Config:      node.Parameters,
		Variables:   execCtx.Variables,
		GetCredential: func(credID uuid.UUID) (*models.CredentialData, error) {
			_, data, err := e.credentialSvc.GetDecrypted(ctx, credID)
			return data, err
		},
	})
	durationMs := int(time.Since(startTime).Milliseconds())

	if err != nil {
		e.executionSvc.FailNodeExecution(ctx, nodeExec.ID, err.Error())
		return err
	}

	// Store output
	execCtx.NodeOutputs[node.ID] = result
	e.executionSvc.CompleteNodeExecution(ctx, nodeExec.ID, models.JSON(result), durationMs)

	return nil
}

func (e *Executor) prepareNodeInput(execCtx *ExecutionContext, node *NodeData) models.JSON {
	input := make(models.JSON)

	// Add workflow input
	input["$input"] = execCtx.Input

	// Add outputs from connected nodes
	for _, conn := range node.Inputs {
		if output, ok := execCtx.NodeOutputs[conn.SourceNodeID]; ok {
			input[conn.SourceNodeID] = output
		}
	}

	// Add variables
	input["$vars"] = execCtx.Variables

	return input
}

type ExecutionContext struct {
	ExecutionID uuid.UUID
	WorkflowID  uuid.UUID
	Input       models.JSON
	Variables   map[string]interface{}
	NodeOutputs map[string]interface{}
}

type NodeExecutionError struct {
	NodeID string
	Err    error
}

func (e *NodeExecutionError) Error() string {
	return fmt.Sprintf("node %s: %v", e.NodeID, e.Err)
}

func extractErrorNodeID(err error) *string {
	if nodeErr, ok := err.(*NodeExecutionError); ok {
		return &nodeErr.NodeID
	}
	return nil
}

func parseNodes(data models.JSON) ([]NodeData, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var nodes []NodeData
	if err := json.Unmarshal(jsonData, &nodes); err != nil {
		return nil, err
	}
	return nodes, nil
}

func parseConnections(data models.JSON) ([]Connection, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var connections []Connection
	if err := json.Unmarshal(jsonData, &connections); err != nil {
		return nil, err
	}
	return connections, nil
}
