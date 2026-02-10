// Package model 提供默认工作流种子数据
package model

import (
	"github.com/kerbos/ticketdesk/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// workflowDef 工作流定义（用于种子数据）
type workflowDef struct {
	Name        string
	Description string
	Nodes       []nodeDef
}

// nodeDef 节点定义
type nodeDef struct {
	Name     string
	NodeType string
	Config   string
	PosX     int
	PosY     int
}

// getDefaultWorkflows 返回所有默认工作流定义
func getDefaultWorkflows() map[string]workflowDef {
	return map[string]workflowDef{
		"Epic": {
			Name:        "Epic 工作流",
			Description: "Epic 默认工作流：开始 → 进行中 → 验收 → 完成",
			Nodes: []nodeDef{
				{Name: "开始", NodeType: "start", Config: "{}", PosX: 100, PosY: 200},
				{Name: "进行中", NodeType: "work", Config: `{"target_status":"in_progress"}`, PosX: 300, PosY: 200},
				{Name: "验收", NodeType: "work", Config: `{}`, PosX: 500, PosY: 200},
				{Name: "完成", NodeType: "end", Config: `{"target_status":"resolved"}`, PosX: 700, PosY: 200},
			},
		},
		"Task": {
			Name:        "任务工作流",
			Description: "任务默认工作流：开始 → 进行中 → 验收 → 完成",
			Nodes: []nodeDef{
				{Name: "开始", NodeType: "start", Config: "{}", PosX: 100, PosY: 200},
				{Name: "进行中", NodeType: "work", Config: `{"target_status":"in_progress"}`, PosX: 300, PosY: 200},
				{Name: "验收", NodeType: "work", Config: `{}`, PosX: 500, PosY: 200},
				{Name: "完成", NodeType: "end", Config: `{"target_status":"resolved"}`, PosX: 700, PosY: 200},
			},
		},
		"Subtask": {
			Name:        "子任务工作流",
			Description: "子任务默认工作流：开始 → 进行中 → 完成",
			Nodes: []nodeDef{
				{Name: "开始", NodeType: "start", Config: "{}", PosX: 100, PosY: 200},
				{Name: "进行中", NodeType: "work", Config: `{"target_status":"in_progress"}`, PosX: 300, PosY: 200},
				{Name: "完成", NodeType: "end", Config: `{"target_status":"resolved"}`, PosX: 500, PosY: 200},
			},
		},
		"Bug": {
			Name:        "Bug 工作流",
			Description: "Bug 默认工作流：开始 → 修复中 → 验证 → 完成",
			Nodes: []nodeDef{
				{Name: "开始", NodeType: "start", Config: "{}", PosX: 100, PosY: 200},
				{Name: "修复中", NodeType: "work", Config: `{"target_status":"in_progress"}`, PosX: 300, PosY: 200},
				{Name: "验证", NodeType: "work", Config: `{}`, PosX: 500, PosY: 200},
				{Name: "完成", NodeType: "end", Config: `{"target_status":"resolved"}`, PosX: 700, PosY: 200},
			},
		},
		"Fault": {
			Name:        "故障工作流",
			Description: "故障默认工作流：开始 → 处理中 → 验收 → 完成",
			Nodes: []nodeDef{
				{Name: "开始", NodeType: "start", Config: "{}", PosX: 100, PosY: 200},
				{Name: "处理中", NodeType: "work", Config: `{"target_status":"in_progress"}`, PosX: 300, PosY: 200},
				{Name: "验收", NodeType: "work", Config: `{}`, PosX: 500, PosY: 200},
				{Name: "完成", NodeType: "end", Config: `{"target_status":"resolved"}`, PosX: 700, PosY: 200},
			},
		},
		"Change": {
			Name:        "变更工作流",
			Description: "变更默认工作流：开始 → 审批 → 实施 → 验收 → 完成",
			Nodes: []nodeDef{
				{Name: "开始", NodeType: "start", Config: "{}", PosX: 100, PosY: 200},
				{Name: "审批", NodeType: "approval", Config: `{"approval_type":"single","approver_role":"administrators","target_status":"in_progress"}`, PosX: 300, PosY: 200},
				{Name: "实施", NodeType: "work", Config: `{}`, PosX: 500, PosY: 200},
				{Name: "验收", NodeType: "work", Config: `{}`, PosX: 700, PosY: 200},
				{Name: "完成", NodeType: "end", Config: `{"target_status":"resolved"}`, PosX: 900, PosY: 200},
			},
		},
		"ServiceRequest": {
			Name:        "服务请求工作流",
			Description: "服务请求默认工作流：开始 → 审批 → 处理中 → 验收 → 完成",
			Nodes: []nodeDef{
				{Name: "开始", NodeType: "start", Config: "{}", PosX: 100, PosY: 200},
				{Name: "审批", NodeType: "approval", Config: `{"approval_type":"single","approver_role":"administrators","target_status":"in_progress"}`, PosX: 300, PosY: 200},
				{Name: "处理中", NodeType: "work", Config: `{}`, PosX: 500, PosY: 200},
				{Name: "验收", NodeType: "work", Config: `{}`, PosX: 700, PosY: 200},
				{Name: "完成", NodeType: "end", Config: `{"target_status":"resolved"}`, PosX: 900, PosY: 200},
			},
		},
	}
}

// SeedDefaultWorkflows 初始化默认工作流
func SeedDefaultWorkflows(db *gorm.DB) error {
	logger.Info("seeding default workflows...")

	workflows := getDefaultWorkflows()

	// 按固定顺序处理，确保幂等
	typeOrder := []string{"Epic", "Task", "Subtask", "Bug", "Fault", "Change", "ServiceRequest"}

	for _, typeName := range typeOrder {
		def := workflows[typeName]

		// 创建或获取工作流
		wf := Workflow{
			Name:        def.Name,
			Description: def.Description,
			Status:      1,
		}
		result := db.Where("name = ? AND project_id IS NULL", wf.Name).FirstOrCreate(&wf)
		if result.Error != nil {
			logger.Error("failed to seed workflow", zap.String("name", wf.Name), zap.Error(result.Error))
			return result.Error
		}

		// 如果工作流已有节点，跳过（避免重复创建）
		var nodeCount int64
		db.Model(&WorkflowNode{}).Where("workflow_id = ?", wf.ID).Count(&nodeCount)
		if nodeCount > 0 {
			logger.Debug("workflow already has nodes, skipping", zap.String("name", wf.Name))
			continue
		}

		// 创建节点
		nodeIDs := make([]uint64, len(def.Nodes))
		for i, nodeDef := range def.Nodes {
			node := WorkflowNode{
				WorkflowID: wf.ID,
				Name:       nodeDef.Name,
				NodeType:   nodeDef.NodeType,
				Config:     nodeDef.Config,
				PositionX:  nodeDef.PosX,
				PositionY:  nodeDef.PosY,
			}
			if err := db.Create(&node).Error; err != nil {
				logger.Error("failed to create workflow node",
					zap.String("workflow", wf.Name),
					zap.String("node", nodeDef.Name),
					zap.Error(err))
				return err
			}
			nodeIDs[i] = node.ID
		}

		// 创建边（按顺序连接节点）
		for i := 0; i < len(nodeIDs)-1; i++ {
			edge := WorkflowEdge{
				WorkflowID:   wf.ID,
				SourceNodeID: nodeIDs[i],
				TargetNodeID: nodeIDs[i+1],
			}
			if err := db.Create(&edge).Error; err != nil {
				logger.Error("failed to create workflow edge",
					zap.String("workflow", wf.Name),
					zap.Error(err))
				return err
			}
		}

		logger.Info("workflow seeded",
			zap.String("name", wf.Name),
			zap.Int("nodes", len(def.Nodes)),
			zap.Int("edges", len(def.Nodes)-1))
	}

	logger.Info("default workflows seeded")
	return nil
}

// issueTypeWorkflowMap 工单类型到工作流名称的映射
var issueTypeWorkflowMap = map[string]string{
	"Epic":           "Epic 工作流",
	"Task":           "任务工作流",
	"Subtask":        "子任务工作流",
	"Bug":            "Bug 工作流",
	"Fault":          "故障工作流",
	"Change":         "变更工作流",
	"ServiceRequest": "服务请求工作流",
}

// InitProjectWorkflowSchemes 初始化项目的工作流方案（为每个工单类型绑定默认工作流）
func InitProjectWorkflowSchemes(db *gorm.DB, projectID uint64) error {
	logger.Info("initializing project workflow schemes", zap.Uint64("project_id", projectID))

	// 获取全局工单类型
	var issueTypes []IssueType
	if err := db.Where("project_id IS NULL").Find(&issueTypes).Error; err != nil {
		return err
	}

	// 获取全局工作流
	var workflows []Workflow
	if err := db.Where("project_id IS NULL AND status = 1").Find(&workflows).Error; err != nil {
		return err
	}

	// 创建工作流名称到 ID 的映射
	wfMap := make(map[string]uint64)
	for _, wf := range workflows {
		wfMap[wf.Name] = wf.ID
	}

	// 为每个工单类型绑定工作流
	for _, it := range issueTypes {
		wfName, ok := issueTypeWorkflowMap[it.Name]
		if !ok {
			continue
		}
		wfID, ok := wfMap[wfName]
		if !ok {
			logger.Warn("workflow not found for issue type",
				zap.String("issue_type", it.Name),
				zap.String("workflow", wfName))
			continue
		}

		scheme := WorkflowScheme{
			ProjectID:   projectID,
			IssueTypeID: it.ID,
			WorkflowID:  wfID,
		}
		result := db.Where("project_id = ? AND issue_type_id = ?", projectID, it.ID).FirstOrCreate(&scheme)
		if result.Error != nil {
			logger.Error("failed to create workflow scheme",
				zap.Uint64("project_id", projectID),
				zap.String("issue_type", it.Name),
				zap.Error(result.Error))
			return result.Error
		}
	}

	logger.Info("project workflow schemes initialized", zap.Uint64("project_id", projectID))
	return nil
}
