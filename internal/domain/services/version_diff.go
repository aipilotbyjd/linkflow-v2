package services

import (
	"context"
	"reflect"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
)

type VersionDiffService struct {
	workflowRepo *repositories.WorkflowRepository
	versionRepo  *repositories.WorkflowVersionRepository
}

func NewVersionDiffService(
	workflowRepo *repositories.WorkflowRepository,
	versionRepo *repositories.WorkflowVersionRepository,
) *VersionDiffService {
	return &VersionDiffService{
		workflowRepo: workflowRepo,
		versionRepo:  versionRepo,
	}
}

// DiffResult represents the difference between two versions
type DiffResult struct {
	FromVersion int             `json:"from_version"`
	ToVersion   int             `json:"to_version"`
	Nodes       NodeDiff        `json:"nodes"`
	Connections ConnectionDiff  `json:"connections"`
	Settings    SettingsDiff    `json:"settings"`
	Summary     DiffSummary     `json:"summary"`
}

type NodeDiff struct {
	Added    []NodeChange `json:"added"`
	Removed  []NodeChange `json:"removed"`
	Modified []NodeChange `json:"modified"`
}

type NodeChange struct {
	NodeID     string                 `json:"node_id"`
	NodeType   string                 `json:"node_type"`
	NodeName   string                 `json:"node_name"`
	OldValue   map[string]interface{} `json:"old_value,omitempty"`
	NewValue   map[string]interface{} `json:"new_value,omitempty"`
	Changes    []FieldChange          `json:"changes,omitempty"`
}

type ConnectionDiff struct {
	Added   []ConnectionChange `json:"added"`
	Removed []ConnectionChange `json:"removed"`
}

type ConnectionChange struct {
	FromNode   string `json:"from_node"`
	FromOutput string `json:"from_output"`
	ToNode     string `json:"to_node"`
	ToInput    string `json:"to_input"`
}

type SettingsDiff struct {
	Changes []FieldChange `json:"changes"`
}

type FieldChange struct {
	Field    string      `json:"field"`
	OldValue interface{} `json:"old_value"`
	NewValue interface{} `json:"new_value"`
}

type DiffSummary struct {
	NodesAdded       int `json:"nodes_added"`
	NodesRemoved     int `json:"nodes_removed"`
	NodesModified    int `json:"nodes_modified"`
	ConnectionsAdded int `json:"connections_added"`
	ConnectionsRemoved int `json:"connections_removed"`
	SettingsChanged  int `json:"settings_changed"`
	TotalChanges     int `json:"total_changes"`
}

// Compare compares two versions of a workflow
func (s *VersionDiffService) Compare(ctx context.Context, workflowID uuid.UUID, fromVersion, toVersion int) (*DiffResult, error) {
	var from, to *models.WorkflowVersion

	// Get the versions
	fromVer, err := s.versionRepo.FindByWorkflowAndVersion(ctx, workflowID, fromVersion)
	if err != nil {
		return nil, err
	}
	from = fromVer

	toVer, err := s.versionRepo.FindByWorkflowAndVersion(ctx, workflowID, toVersion)
	if err != nil {
		return nil, err
	}
	to = toVer

	result := &DiffResult{
		FromVersion: fromVersion,
		ToVersion:   toVersion,
		Nodes:       s.diffNodes(from.Nodes, to.Nodes),
		Connections: s.diffConnections(from.Connections, to.Connections),
		Settings:    s.diffSettings(from.Settings, to.Settings),
	}

	// Calculate summary
	result.Summary = DiffSummary{
		NodesAdded:       len(result.Nodes.Added),
		NodesRemoved:     len(result.Nodes.Removed),
		NodesModified:    len(result.Nodes.Modified),
		ConnectionsAdded: len(result.Connections.Added),
		ConnectionsRemoved: len(result.Connections.Removed),
		SettingsChanged:  len(result.Settings.Changes),
	}
	result.Summary.TotalChanges = result.Summary.NodesAdded + result.Summary.NodesRemoved +
		result.Summary.NodesModified + result.Summary.ConnectionsAdded +
		result.Summary.ConnectionsRemoved + result.Summary.SettingsChanged

	return result, nil
}

// CompareWithCurrent compares a version with the current workflow state
func (s *VersionDiffService) CompareWithCurrent(ctx context.Context, workflowID uuid.UUID, version int) (*DiffResult, error) {
	workflow, err := s.workflowRepo.FindByID(ctx, workflowID)
	if err != nil {
		return nil, err
	}

	fromVer, err := s.versionRepo.FindByWorkflowAndVersion(ctx, workflowID, version)
	if err != nil {
		return nil, err
	}

	result := &DiffResult{
		FromVersion: version,
		ToVersion:   workflow.Version,
		Nodes:       s.diffNodes(fromVer.Nodes, workflow.Nodes),
		Connections: s.diffConnections(fromVer.Connections, workflow.Connections),
		Settings:    s.diffSettings(fromVer.Settings, workflow.Settings),
	}

	result.Summary = DiffSummary{
		NodesAdded:       len(result.Nodes.Added),
		NodesRemoved:     len(result.Nodes.Removed),
		NodesModified:    len(result.Nodes.Modified),
		ConnectionsAdded: len(result.Connections.Added),
		ConnectionsRemoved: len(result.Connections.Removed),
		SettingsChanged:  len(result.Settings.Changes),
	}
	result.Summary.TotalChanges = result.Summary.NodesAdded + result.Summary.NodesRemoved +
		result.Summary.NodesModified + result.Summary.ConnectionsAdded +
		result.Summary.ConnectionsRemoved + result.Summary.SettingsChanged

	return result, nil
}

func (s *VersionDiffService) diffNodes(from, to models.JSONArray) NodeDiff {
	diff := NodeDiff{
		Added:    []NodeChange{},
		Removed:  []NodeChange{},
		Modified: []NodeChange{},
	}

	fromMap := s.nodesToMap(from)
	toMap := s.nodesToMap(to)

	// Find added and modified
	for id, toNode := range toMap {
		if fromNode, exists := fromMap[id]; exists {
			// Check if modified
			if changes := s.compareNodes(fromNode, toNode); len(changes) > 0 {
				diff.Modified = append(diff.Modified, NodeChange{
					NodeID:   id,
					NodeType: getStringField(toNode, "type"),
					NodeName: getStringField(toNode, "name"),
					OldValue: fromNode,
					NewValue: toNode,
					Changes:  changes,
				})
			}
		} else {
			// Added
			diff.Added = append(diff.Added, NodeChange{
				NodeID:   id,
				NodeType: getStringField(toNode, "type"),
				NodeName: getStringField(toNode, "name"),
				NewValue: toNode,
			})
		}
	}

	// Find removed
	for id, fromNode := range fromMap {
		if _, exists := toMap[id]; !exists {
			diff.Removed = append(diff.Removed, NodeChange{
				NodeID:   id,
				NodeType: getStringField(fromNode, "type"),
				NodeName: getStringField(fromNode, "name"),
				OldValue: fromNode,
			})
		}
	}

	return diff
}

func (s *VersionDiffService) diffConnections(from, to models.JSONArray) ConnectionDiff {
	diff := ConnectionDiff{
		Added:   []ConnectionChange{},
		Removed: []ConnectionChange{},
	}

	fromSet := s.connectionsToSet(from)
	toSet := s.connectionsToSet(to)

	// Find added
	for key, conn := range toSet {
		if _, exists := fromSet[key]; !exists {
			diff.Added = append(diff.Added, conn)
		}
	}

	// Find removed
	for key, conn := range fromSet {
		if _, exists := toSet[key]; !exists {
			diff.Removed = append(diff.Removed, conn)
		}
	}

	return diff
}

func (s *VersionDiffService) diffSettings(from, to models.JSON) SettingsDiff {
	diff := SettingsDiff{Changes: []FieldChange{}}

	fromMap := map[string]interface{}(from)
	toMap := map[string]interface{}(to)

	// Check all keys in both maps
	allKeys := make(map[string]bool)
	for k := range fromMap {
		allKeys[k] = true
	}
	for k := range toMap {
		allKeys[k] = true
	}

	for key := range allKeys {
		fromVal, fromExists := fromMap[key]
		toVal, toExists := toMap[key]

		if !fromExists {
			diff.Changes = append(diff.Changes, FieldChange{
				Field:    key,
				OldValue: nil,
				NewValue: toVal,
			})
		} else if !toExists {
			diff.Changes = append(diff.Changes, FieldChange{
				Field:    key,
				OldValue: fromVal,
				NewValue: nil,
			})
		} else if !reflect.DeepEqual(fromVal, toVal) {
			diff.Changes = append(diff.Changes, FieldChange{
				Field:    key,
				OldValue: fromVal,
				NewValue: toVal,
			})
		}
	}

	return diff
}

func (s *VersionDiffService) nodesToMap(nodes models.JSONArray) map[string]map[string]interface{} {
	result := make(map[string]map[string]interface{})
	for _, n := range nodes {
		if node, ok := n.(map[string]interface{}); ok {
			if id, ok := node["id"].(string); ok {
				result[id] = node
			}
		}
	}
	return result
}

func (s *VersionDiffService) connectionsToSet(connections models.JSONArray) map[string]ConnectionChange {
	result := make(map[string]ConnectionChange)
	for _, c := range connections {
		if conn, ok := c.(map[string]interface{}); ok {
			change := ConnectionChange{
				FromNode:   getStringField(conn, "from_node"),
				FromOutput: getStringField(conn, "from_output"),
				ToNode:     getStringField(conn, "to_node"),
				ToInput:    getStringField(conn, "to_input"),
			}
			key := change.FromNode + ":" + change.FromOutput + "->" + change.ToNode + ":" + change.ToInput
			result[key] = change
		}
	}
	return result
}

func (s *VersionDiffService) compareNodes(from, to map[string]interface{}) []FieldChange {
	var changes []FieldChange

	// Compare specific fields
	fields := []string{"name", "type", "position", "parameters", "disabled", "notes"}
	for _, field := range fields {
		fromVal, fromExists := from[field]
		toVal, toExists := to[field]

		if !fromExists && !toExists {
			continue
		}

		if !fromExists {
			changes = append(changes, FieldChange{Field: field, OldValue: nil, NewValue: toVal})
		} else if !toExists {
			changes = append(changes, FieldChange{Field: field, OldValue: fromVal, NewValue: nil})
		} else if !reflect.DeepEqual(fromVal, toVal) {
			changes = append(changes, FieldChange{Field: field, OldValue: fromVal, NewValue: toVal})
		}
	}

	return changes
}

func getStringField(m map[string]interface{}, field string) string {
	if v, ok := m[field].(string); ok {
		return v
	}
	return ""
}
