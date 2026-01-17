package workflow

import (
	"net/http"
	"reflect"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type CompareVersionsHandler struct {
	versionRepo workflow.VersionRepository
}

func NewCompareVersionsHandler(versionRepo workflow.VersionRepository) *CompareVersionsHandler {
	return &CompareVersionsHandler{versionRepo: versionRepo}
}

func (h *CompareVersionsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workflowIDStr := chi.URLParam(r, "workflowId")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		common.BadRequest(w, "invalid workflow ID")
		return
	}

	v1Str := r.URL.Query().Get("v1")
	v2Str := r.URL.Query().Get("v2")

	if v1Str == "" || v2Str == "" {
		common.BadRequest(w, "both v1 and v2 query parameters are required")
		return
	}

	v1, err := strconv.Atoi(v1Str)
	if err != nil {
		common.BadRequest(w, "invalid v1 version number")
		return
	}

	v2, err := strconv.Atoi(v2Str)
	if err != nil {
		common.BadRequest(w, "invalid v2 version number")
		return
	}

	version1, err := h.versionRepo.FindByWorkflowAndVersion(r.Context(), workflowID, v1)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	version2, err := h.versionRepo.FindByWorkflowAndVersion(r.Context(), workflowID, v2)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	differences, summary := compareVersions(version1, version2)

	common.Success(w, CompareVersionsResponse{
		Version1:    v1,
		Version2:    v2,
		Differences: differences,
		Summary:     summary,
	})
}

func compareVersions(v1, v2 *workflow.Version) ([]VersionDifference, DiffSummary) {
	var diffs []VersionDifference
	summary := DiffSummary{}

	nodesCompare := compareJSONArrays(v1.Nodes, v2.Nodes, "nodes")
	diffs = append(diffs, nodesCompare.diffs...)
	summary.NodesAdded = nodesCompare.added
	summary.NodesRemoved = nodesCompare.removed
	summary.NodesModified = nodesCompare.modified

	connCompare := compareJSONArrays(v1.Connections, v2.Connections, "connections")
	diffs = append(diffs, connCompare.diffs...)
	summary.ConnectionsAdded = connCompare.added
	summary.ConnectionsRemoved = connCompare.removed

	if !reflect.DeepEqual(v1.Settings, v2.Settings) {
		summary.SettingsChanged = true
		diffs = append(diffs, VersionDifference{
			Type:        "modified",
			Path:        "settings",
			OldValue:    v1.Settings,
			NewValue:    v2.Settings,
			Description: "Workflow settings changed",
		})
	}

	return diffs, summary
}

type arrayCompareResult struct {
	diffs    []VersionDifference
	added    int
	removed  int
	modified int
}

func compareJSONArrays(arr1, arr2 types.JSONArray, basePath string) arrayCompareResult {
	result := arrayCompareResult{}
	
	len1 := len(arr1)
	len2 := len(arr2)
	
	if len2 > len1 {
		result.added = len2 - len1
		result.diffs = append(result.diffs, VersionDifference{
			Type:        "added",
			Path:        basePath,
			Description: strconv.Itoa(result.added) + " items added",
		})
	} else if len1 > len2 {
		result.removed = len1 - len2
		result.diffs = append(result.diffs, VersionDifference{
			Type:        "removed",
			Path:        basePath,
			Description: strconv.Itoa(result.removed) + " items removed",
		})
	}

	minLen := len1
	if len2 < minLen {
		minLen = len2
	}

	for i := 0; i < minLen; i++ {
		if !reflect.DeepEqual(arr1[i], arr2[i]) {
			result.modified++
		}
	}

	if result.modified > 0 {
		result.diffs = append(result.diffs, VersionDifference{
			Type:        "modified",
			Path:        basePath,
			Description: strconv.Itoa(result.modified) + " items modified",
		})
	}

	return result
}
