// Package model 提供默认工作流种子数据
package model

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/kerbos/ticketdesk/pkg/logger"
)

// workflowDef 工作流定义（用于种子数据）
type workflowDef struct {
	Name        string
	Description string
	Nodes       []nodeDef
	ExtraEdges  []edgeDef // 额外的边（除顺序连接外的分支边）
}

// nodeDef 节点定义
type nodeDef struct {
	Name     string
	NodeType string
	Config   string
	PosX     int
	PosY     int
}

// edgeDef 额外边定义（用于分支，如审批拒绝）
type edgeDef struct {
	FromNode      string // 源节点名称
	ToNode        string // 目标节点名称
	ConditionExpr string // 条件表达式（如 "approved", "rejected"）
}

// approvalRejectBranch 返回审批节点的标准拒绝分支定义
// 拒绝后流转到"已取消"结束节点
func approvalRejectBranch() (nodeDef, edgeDef) {
	return nodeDef{
			Name: "已取消", NodeType: "end", Config: `{"target_status":"closed"}`, PosX: 300, PosY: 400,
		}, edgeDef{
			FromNode: "审批", ToNode: "已取消", ConditionExpr: "rejected",
		}
}

// getDefaultWorkflows 返回所有默认工作流定义
func getDefaultWorkflows() map[string]workflowDef {
	rejectNode, rejectEdge := approvalRejectBranch()

	return map[string]workflowDef{
		"Epic": {
			Name:        "Epic 工作流",
			Description: "Epic 默认工作流：开始 → 审批(通过/拒绝) → 进行中 → 验收 → 完成",
			Nodes: []nodeDef{
				{Name: "开始", NodeType: "start", Config: "{}", PosX: 100, PosY: 200},
				{Name: "审批", NodeType: "approval", Config: `{"approval_type":"single","approver_role":"administrators"}`, PosX: 300, PosY: 200},
				{Name: "进行中", NodeType: "work", Config: `{"target_status":"in_progress"}`, PosX: 500, PosY: 200},
				{Name: "验收", NodeType: "work", Config: `{}`, PosX: 700, PosY: 200},
				{Name: "完成", NodeType: "end", Config: `{"target_status":"resolved"}`, PosX: 900, PosY: 200},
				rejectNode,
			},
			ExtraEdges: []edgeDef{rejectEdge},
		},
		"Task": {
			Name:        "任务工作流",
			Description: "任务默认工作流：开始 → 审批(通过/拒绝) → 进行中 → 验收 → 完成",
			Nodes: []nodeDef{
				{Name: "开始", NodeType: "start", Config: "{}", PosX: 100, PosY: 200},
				{Name: "审批", NodeType: "approval", Config: `{"approval_type":"single","approver_role":"administrators"}`, PosX: 300, PosY: 200},
				{Name: "进行中", NodeType: "work", Config: `{"target_status":"in_progress"}`, PosX: 500, PosY: 200},
				{Name: "验收", NodeType: "work", Config: `{}`, PosX: 700, PosY: 200},
				{Name: "完成", NodeType: "end", Config: `{"target_status":"resolved"}`, PosX: 900, PosY: 200},
				rejectNode,
			},
			ExtraEdges: []edgeDef{rejectEdge},
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
			Description: "Bug 默认工作流：开始 → 审批(通过/拒绝) → 修复中 → 验证 → 完成",
			Nodes: []nodeDef{
				{Name: "开始", NodeType: "start", Config: "{}", PosX: 100, PosY: 200},
				{Name: "审批", NodeType: "approval", Config: `{"approval_type":"single","approver_role":"administrators"}`, PosX: 300, PosY: 200},
				{Name: "修复中", NodeType: "work", Config: `{"target_status":"in_progress"}`, PosX: 500, PosY: 200},
				{Name: "验证", NodeType: "work", Config: `{}`, PosX: 700, PosY: 200},
				{Name: "完成", NodeType: "end", Config: `{"target_status":"resolved"}`, PosX: 900, PosY: 200},
				rejectNode,
			},
			ExtraEdges: []edgeDef{rejectEdge},
		},
		"Fault": {
			Name:        "故障工作流",
			Description: "故障默认工作流：开始 → 审批(通过/拒绝) → 处理中 → 验收 → 完成",
			Nodes: []nodeDef{
				{Name: "开始", NodeType: "start", Config: "{}", PosX: 100, PosY: 200},
				{Name: "审批", NodeType: "approval", Config: `{"approval_type":"single","approver_role":"administrators"}`, PosX: 300, PosY: 200},
				{Name: "处理中", NodeType: "work", Config: `{"target_status":"in_progress"}`, PosX: 500, PosY: 200},
				{Name: "验收", NodeType: "work", Config: `{}`, PosX: 700, PosY: 200},
				{Name: "完成", NodeType: "end", Config: `{"target_status":"resolved"}`, PosX: 900, PosY: 200},
				rejectNode,
			},
			ExtraEdges: []edgeDef{rejectEdge},
		},
		"Change": {
			Name:        "变更工作流",
			Description: "变更默认工作流：开始 → 审批(通过/拒绝) → 实施 → 验收 → 完成",
			Nodes: []nodeDef{
				{Name: "开始", NodeType: "start", Config: "{}", PosX: 100, PosY: 200},
				{Name: "审批", NodeType: "approval", Config: `{"approval_type":"single","approver_role":"administrators"}`, PosX: 300, PosY: 200},
				{Name: "实施", NodeType: "work", Config: `{}`, PosX: 500, PosY: 200},
				{Name: "验收", NodeType: "work", Config: `{}`, PosX: 700, PosY: 200},
				{Name: "完成", NodeType: "end", Config: `{"target_status":"resolved"}`, PosX: 900, PosY: 200},
				rejectNode,
			},
			ExtraEdges: []edgeDef{rejectEdge},
		},
		"ServiceRequest": {
			Name:        "服务请求工作流",
			Description: "服务请求默认工作流：开始 → 审批(通过/拒绝) → 处理中 → 验收 → 完成",
			Nodes: []nodeDef{
				{Name: "开始", NodeType: "start", Config: "{}", PosX: 100, PosY: 200},
				{Name: "审批", NodeType: "approval", Config: `{"approval_type":"single","approver_role":"administrators"}`, PosX: 300, PosY: 200},
				{Name: "处理中", NodeType: "work", Config: `{}`, PosX: 500, PosY: 200},
				{Name: "验收", NodeType: "work", Config: `{}`, PosX: 700, PosY: 200},
				{Name: "完成", NodeType: "end", Config: `{"target_status":"resolved"}`, PosX: 900, PosY: 200},
				rejectNode,
			},
			ExtraEdges: []edgeDef{rejectEdge},
		},
		"Alert": {
			Name:        "告警工作流",
			Description: "告警默认工作流：开始 → 处理中 → 待确认(确认完成/继续处理) → 完成",
			Nodes: []nodeDef{
				{Name: "开始", NodeType: "start", Config: "{}", PosX: 100, PosY: 200},
				{Name: "处理中", NodeType: "work", Config: `{"target_status":"in_progress"}`, PosX: 300, PosY: 200},
				{Name: "待确认", NodeType: "work", Config: `{"target_status":"pending_review"}`, PosX: 500, PosY: 200},
				{Name: "完成", NodeType: "end", Config: `{"target_status":"resolved"}`, PosX: 700, PosY: 200},
			},
			ExtraEdges: []edgeDef{
				{FromNode: "待确认", ToNode: "完成", ConditionExpr: "confirmed"},
				{FromNode: "待确认", ToNode: "处理中", ConditionExpr: "continue"},
			},
		},
	}
}

// SeedDefaultWorkflows 初始化默认工作流
func SeedDefaultWorkflows(db *gorm.DB) error {
	logger.Info("seeding default workflows...")

	workflows := getDefaultWorkflows()

	// 按固定顺序处理，确保幂等
	typeOrder := []string{"Epic", "Task", "Subtask", "Bug", "Fault", "Change", "ServiceRequest", "Alert"}

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

		// 如果工作流已有节点，跳过（避免覆盖用户自定义的工作流）
		var nodeCount int64
		db.Model(&WorkflowNode{}).Where("workflow_id = ?", wf.ID).Count(&nodeCount)
		if nodeCount > 0 {
			logger.Debug("workflow already has nodes, skipping", zap.String("name", wf.Name))
			continue
		}

		// 创建节点
		nodeIDs := make([]uint64, len(def.Nodes))
		nodeNameToID := make(map[string]uint64, len(def.Nodes))
		for i, nd := range def.Nodes {
			node := WorkflowNode{
				WorkflowID: wf.ID,
				Name:       nd.Name,
				NodeType:   nd.NodeType,
				Config:     nd.Config,
				PositionX:  nd.PosX,
				PositionY:  nd.PosY,
			}
			if err := db.Create(&node).Error; err != nil {
				logger.Error("failed to create workflow node",
					zap.String("workflow", wf.Name),
					zap.String("node", nd.Name),
					zap.Error(err))
				return err
			}
			nodeIDs[i] = node.ID
			nodeNameToID[nd.Name] = node.ID
		}

		// 创建主线边（按顺序连接节点，审批节点出边标记 approved）
		// 构建 ExtraEdges 中已定义出边的源节点集合，避免主线边重复创建
		extraEdgeSources := make(map[string]bool)
		for _, ed := range def.ExtraEdges {
			extraEdgeSources[ed.FromNode] = true
		}

		edgeCount := 0
		for i := 0; i < len(nodeIDs)-1; i++ {
			// 跳过额外节点（如"已取消"），它们不在主线上
			if def.Nodes[i+1].Name == "已取消" {
				continue
			}
			// 跳过已在 ExtraEdges 中完全定义出边的节点（如"待确认"）
			if extraEdgeSources[def.Nodes[i].Name] && def.Nodes[i].NodeType != "approval" {
				continue
			}
			edge := WorkflowEdge{
				WorkflowID:   wf.ID,
				SourceNodeID: nodeIDs[i],
				TargetNodeID: nodeIDs[i+1],
			}
			// 审批节点的主线出边标记为 approved
			if def.Nodes[i].NodeType == "approval" {
				edge.ConditionExpr = "approved"
			}
			if err := db.Create(&edge).Error; err != nil {
				logger.Error("failed to create workflow edge",
					zap.String("workflow", wf.Name),
					zap.Error(err))
				return err
			}
			edgeCount++
		}

		// 创建额外分支边（如审批拒绝 → 已取消）
		for _, ed := range def.ExtraEdges {
			fromID, ok1 := nodeNameToID[ed.FromNode]
			toID, ok2 := nodeNameToID[ed.ToNode]
			if !ok1 || !ok2 {
				logger.Warn("extra edge node not found",
					zap.String("workflow", wf.Name),
					zap.String("from", ed.FromNode),
					zap.String("to", ed.ToNode))
				continue
			}
			edge := WorkflowEdge{
				WorkflowID:    wf.ID,
				SourceNodeID:  fromID,
				TargetNodeID:  toID,
				ConditionExpr: ed.ConditionExpr,
			}
			if err := db.Create(&edge).Error; err != nil {
				logger.Error("failed to create extra workflow edge",
					zap.String("workflow", wf.Name),
					zap.Error(err))
				return err
			}
			edgeCount++
		}

		logger.Info("workflow seeded",
			zap.String("name", wf.Name),
			zap.Int("nodes", len(def.Nodes)),
			zap.Int("edges", edgeCount))
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
	"Alert":          "告警工作流",
}

// MigrateWorkflowApprovalNodes 一次性迁移：为现有全局工作流添加审批节点
// 通过 system_configs 表标记只执行一次，不影响用户自定义工作流
func MigrateWorkflowApprovalNodes(db *gorm.DB) error {
	const migrationKey = "migration.workflow_approval_v1"

	// 检查是否已执行过
	var count int64
	db.Model(&SystemConfig{}).Where("config_key = ?", migrationKey).Count(&count)
	if count > 0 {
		return nil
	}

	logger.Info("starting one-time migration: add approval nodes to global workflows")

	// 只处理全局工作流（project_id IS NULL），不影响自定义工作流
	workflowNames := []string{
		"Epic 工作流", "任务工作流", "Bug 工作流",
		"故障工作流", "变更工作流", "服务请求工作流",
	}

	for _, wfName := range workflowNames {
		var wf Workflow
		if err := db.Where("name = ? AND project_id IS NULL", wfName).First(&wf).Error; err != nil {
			logger.Debug("workflow not found, skipping migration", zap.String("name", wfName))
			continue
		}

		// 检查是否已有审批节点（未被软删除的）
		var approvalCount int64
		db.Model(&WorkflowNode{}).Where("workflow_id = ? AND node_type = ?", wf.ID, "approval").Count(&approvalCount)

		if approvalCount > 0 {
			// 已有审批节点（可能之前自动重建过），只修复断裂的工作流实例
			fixBrokenWorkflowInstances(db, wf.ID)
			continue
		}

		// 没有审批节点，需要在 start 和第一个 work 节点之间插入审批节点
		if err := insertApprovalNode(db, &wf); err != nil {
			logger.Error("failed to insert approval node",
				zap.String("workflow", wfName), zap.Error(err))
			continue
		}
	}

	// 记录迁移完成
	db.Create(&SystemConfig{
		ConfigKey:   migrationKey,
		ConfigValue: "done",
		ConfigType:  "string",
		Category:    "migration",
		Description: "一次性迁移：为全局工作流添加审批节点",
		IsSecret:    false,
	})

	logger.Info("one-time migration completed: workflow approval nodes")
	return nil
}

// insertApprovalNode 在 start 和第一个 work 节点之间插入审批节点
func insertApprovalNode(db *gorm.DB, wf *Workflow) error {
	// 找到开始节点
	var startNode WorkflowNode
	if err := db.Where("workflow_id = ? AND node_type = ?", wf.ID, "start").First(&startNode).Error; err != nil {
		return fmt.Errorf("start node not found: %w", err)
	}

	// 找到开始节点的出边
	var startEdge WorkflowEdge
	if err := db.Where("workflow_id = ? AND source_node_id = ?", wf.ID, startNode.ID).First(&startEdge).Error; err != nil {
		return fmt.Errorf("start edge not found: %w", err)
	}
	firstWorkNodeID := startEdge.TargetNodeID

	// 创建审批节点
	approvalNode := WorkflowNode{
		WorkflowID: wf.ID,
		Name:       "审批",
		NodeType:   "approval",
		Config:     `{"approval_type":"single","approver_role":"administrators"}`,
		PositionX:  300,
		PositionY:  200,
	}
	if err := db.Create(&approvalNode).Error; err != nil {
		return fmt.Errorf("failed to create approval node: %w", err)
	}

	// 更新开始节点的出边：start → approval（原来是 start → firstWork）
	if err := db.Model(&WorkflowEdge{}).Where("id = ?", startEdge.ID).
		Update("target_node_id", approvalNode.ID).Error; err != nil {
		return fmt.Errorf("failed to update start edge: %w", err)
	}

	// 创建新边：approval → firstWork
	newEdge := WorkflowEdge{
		WorkflowID:   wf.ID,
		SourceNodeID: approvalNode.ID,
		TargetNodeID: firstWorkNodeID,
	}
	if err := db.Create(&newEdge).Error; err != nil {
		return fmt.Errorf("failed to create approval->work edge: %w", err)
	}

	// 调整后续节点的 X 坐标，为审批节点腾出位置
	db.Model(&WorkflowNode{}).
		Where("workflow_id = ? AND id != ? AND id != ? AND position_x >= 300",
			wf.ID, startNode.ID, approvalNode.ID).
		UpdateColumn("position_x", gorm.Expr("position_x + 200"))

	logger.Info("approval node inserted into workflow",
		zap.String("workflow", wf.Name),
		zap.Uint64("approval_node_id", approvalNode.ID))

	return nil
}

// fixBrokenWorkflowInstances 修复 current_node_id 指向已删除节点的工作流实例
// 同时修复 approval_records 中引用旧节点 ID 的审批记录
func fixBrokenWorkflowInstances(db *gorm.DB, workflowID uint64) {
	// 获取该工作流的所有实例
	var instances []WorkflowInstance
	db.Where("workflow_id = ?", workflowID).Find(&instances)
	if len(instances) == 0 {
		return
	}

	// 获取当前有效节点（未软删除）
	var validNodes []WorkflowNode
	db.Where("workflow_id = ?", workflowID).Find(&validNodes)

	validNodeIDs := make(map[uint64]bool, len(validNodes))
	nodeByName := make(map[string]*WorkflowNode, len(validNodes))
	for i := range validNodes {
		validNodeIDs[validNodes[i].ID] = true
		nodeByName[validNodes[i].Name] = &validNodes[i]
	}

	// 构建旧节点→新节点的映射（通过软删除的旧节点名称匹配）
	oldToNewNodeID := make(map[uint64]uint64)

	fixed := 0
	for _, inst := range instances {
		if validNodeIDs[inst.CurrentNodeID] {
			continue // 节点有效，无需修复
		}

		// 当前节点无效，尝试通过软删除的旧节点名称匹配新节点
		var oldNode WorkflowNode
		if err := db.Unscoped().Where("id = ?", inst.CurrentNodeID).First(&oldNode).Error; err != nil {
			logger.Warn("cannot find old node for broken instance",
				zap.Uint64("instance_id", inst.ID),
				zap.Uint64("node_id", inst.CurrentNodeID))
			continue
		}

		newNode, ok := nodeByName[oldNode.Name]
		if !ok {
			logger.Warn("no matching new node for broken instance",
				zap.Uint64("instance_id", inst.ID),
				zap.String("node_name", oldNode.Name))
			continue
		}

		// 修复工作流实例的 current_node_id
		db.Model(&WorkflowInstance{}).Where("id = ?", inst.ID).
			Update("current_node_id", newNode.ID)
		oldToNewNodeID[inst.CurrentNodeID] = newNode.ID
		fixed++

		// 修复该实例的审批记录 node_id（否则 Approve 按新 node_id 查不到旧记录）
		result := db.Model(&ApprovalRecord{}).
			Where("instance_id = ? AND node_id = ?", inst.ID, oldNode.ID).
			Update("node_id", newNode.ID)
		if result.RowsAffected > 0 {
			logger.Info("remapped approval records to new node",
				zap.Uint64("instance_id", inst.ID),
				zap.Int64("count", result.RowsAffected),
				zap.Uint64("old_node_id", oldNode.ID),
				zap.Uint64("new_node_id", newNode.ID))
		}

		logger.Info("remapped workflow instance to new node",
			zap.Uint64("instance_id", inst.ID),
			zap.String("node_name", oldNode.Name),
			zap.Uint64("old_node_id", inst.CurrentNodeID),
			zap.Uint64("new_node_id", newNode.ID))
	}

	if fixed > 0 {
		logger.Info("fixed broken workflow instances",
			zap.Uint64("workflow_id", workflowID),
			zap.Int("count", fixed))
	}
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

// MigrateAlertWorkflowSchemes 一次性迁移：为已有项目补充 Alert 工作流方案
// 并为缺失工作流实例的告警工单补建工作流实例
func MigrateAlertWorkflowSchemes(db *gorm.DB) error {
	const migrationKey = "migration.alert_workflow_scheme_v2"

	// 检查是否已执行过
	var count int64
	db.Model(&SystemConfig{}).Where("config_key = ?", migrationKey).Count(&count)
	if count > 0 {
		return nil
	}

	logger.Info("starting one-time migration: alert workflow schemes")

	// 获取所有名为 Alert 的工单类型（可能存在重复记录）
	var alertTypes []IssueType
	db.Where("name = ? AND project_id IS NULL", "Alert").Find(&alertTypes)
	if len(alertTypes) == 0 {
		logger.Warn("Alert issue type not found, skipping migration")
		db.Create(&SystemConfig{
			ConfigKey:   migrationKey,
			ConfigValue: "skipped_no_alert_type",
			ConfigType:  "string",
			Category:    "migration",
			Description: "一次性迁移：Alert 工单类型不存在，跳过",
			IsSecret:    false,
		})
		return nil
	}

	// 获取告警工作流
	var alertWorkflow Workflow
	if err := db.Where("name = ? AND project_id IS NULL", "告警工作流").First(&alertWorkflow).Error; err != nil {
		logger.Warn("Alert workflow not found, skipping migration")
		db.Create(&SystemConfig{
			ConfigKey:   migrationKey,
			ConfigValue: "skipped_no_alert_workflow",
			ConfigType:  "string",
			Category:    "migration",
			Description: "一次性迁移：告警工作流不存在，跳过",
			IsSecret:    false,
		})
		return nil
	}

	// 收集所有 Alert 工单类型 ID
	alertTypeIDs := make([]uint64, 0, len(alertTypes))
	for _, at := range alertTypes {
		alertTypeIDs = append(alertTypeIDs, at.ID)
	}

	// 获取所有项目
	var projects []Project
	db.Find(&projects)

	// 为每个项目、每个 Alert 类型 ID 补充工作流方案
	for _, proj := range projects {
		for _, typeID := range alertTypeIDs {
			scheme := WorkflowScheme{
				ProjectID:   proj.ID,
				IssueTypeID: typeID,
				WorkflowID:  alertWorkflow.ID,
			}
			result := db.Where("project_id = ? AND issue_type_id = ?", proj.ID, typeID).FirstOrCreate(&scheme)
			if result.Error != nil {
				logger.Error("failed to create alert workflow scheme",
					zap.Uint64("project_id", proj.ID),
					zap.Uint64("issue_type_id", typeID),
					zap.Error(result.Error))
				continue
			}
			if result.RowsAffected > 0 {
				logger.Info("alert workflow scheme created for project",
					zap.Uint64("project_id", proj.ID),
					zap.Uint64("issue_type_id", typeID),
					zap.String("project_name", proj.Name))
			}
		}
	}

	// 为所有 Alert 类型的工单补建缺失的工作流实例
	for _, typeID := range alertTypeIDs {
		fixAlertIssuesWithoutWorkflow(db, typeID, alertWorkflow.ID)
	}

	// 记录迁移完成
	db.Create(&SystemConfig{
		ConfigKey:   migrationKey,
		ConfigValue: "done",
		ConfigType:  "string",
		Category:    "migration",
		Description: "一次性迁移：为已有项目补充 Alert 工作流方案",
		IsSecret:    false,
	})

	logger.Info("one-time migration completed: alert workflow schemes")
	return nil
}

// MigrateApprovalRejectBranch 一次性迁移：为现有工作流的审批节点添加拒绝分支
// 添加"已取消"结束节点 + rejected 边，并给现有审批→下一节点的边加上 approved 条件
func MigrateApprovalRejectBranch(db *gorm.DB) error {
	const migrationKey = "migration.approval_reject_branch_v1"

	var count int64
	db.Model(&SystemConfig{}).Where("config_key = ?", migrationKey).Count(&count)
	if count > 0 {
		return nil
	}

	logger.Info("starting one-time migration: add approval reject branches to global workflows")

	// 获取所有全局工作流
	var workflows []Workflow
	db.Where("project_id IS NULL AND status = 1").Find(&workflows)

	for _, wf := range workflows {
		// 找到审批节点
		var approvalNodes []WorkflowNode
		db.Where("workflow_id = ? AND node_type = ?", wf.ID, "approval").Find(&approvalNodes)
		if len(approvalNodes) == 0 {
			continue // 无审批节点，跳过
		}

		for _, approvalNode := range approvalNodes {
			// 检查是否已有 rejected 边
			var rejectedCount int64
			db.Model(&WorkflowEdge{}).
				Where("workflow_id = ? AND source_node_id = ? AND condition_expr = ?",
					wf.ID, approvalNode.ID, "rejected").
				Count(&rejectedCount)
			if rejectedCount > 0 {
				continue // 已有拒绝分支，跳过
			}

			// 检查是否已有"已取消"结束节点
			var cancelNode WorkflowNode
			err := db.Where("workflow_id = ? AND name = ? AND node_type = ?",
				wf.ID, "已取消", "end").First(&cancelNode).Error
			if err != nil {
				// 创建"已取消"结束节点
				cancelNode = WorkflowNode{
					WorkflowID: wf.ID,
					Name:       "已取消",
					NodeType:   "end",
					Config:     `{"target_status":"closed"}`,
					PositionX:  300,
					PositionY:  400,
				}
				if err := db.Create(&cancelNode).Error; err != nil {
					logger.Error("failed to create cancel node",
						zap.String("workflow", wf.Name), zap.Error(err))
					continue
				}
			}

			// 给审批节点的现有出边加上 approved 条件
			db.Model(&WorkflowEdge{}).
				Where("workflow_id = ? AND source_node_id = ? AND (condition_expr IS NULL OR condition_expr = '')",
					wf.ID, approvalNode.ID).
				Update("condition_expr", "approved")

			// 创建 rejected 边：审批 → 已取消
			rejectedEdge := WorkflowEdge{
				WorkflowID:    wf.ID,
				SourceNodeID:  approvalNode.ID,
				TargetNodeID:  cancelNode.ID,
				ConditionExpr: "rejected",
			}
			if err := db.Create(&rejectedEdge).Error; err != nil {
				logger.Error("failed to create rejected edge",
					zap.String("workflow", wf.Name), zap.Error(err))
				continue
			}

			logger.Info("added approval reject branch",
				zap.String("workflow", wf.Name),
				zap.Uint64("approval_node_id", approvalNode.ID),
				zap.Uint64("cancel_node_id", cancelNode.ID))
		}
	}

	// 记录迁移完成
	db.Create(&SystemConfig{
		ConfigKey:   migrationKey,
		ConfigValue: "done",
		ConfigType:  "string",
		Category:    "migration",
		Description: "一次性迁移：为审批节点添加拒绝分支",
		IsSecret:    false,
	})

	logger.Info("one-time migration completed: approval reject branches")
	return nil
}

// MigrateAlertWorkflowReviewNode 一次性迁移：为告警工作流添加"待确认"节点和条件边
// 将 处理中→完成 改为 处理中→待确认，待确认有两条条件边：confirmed→完成、continue→处理中
func MigrateAlertWorkflowReviewNode(db *gorm.DB) error {
	const migrationKey = "migration.alert_workflow_review_node_v1"

	var count int64
	db.Model(&SystemConfig{}).Where("config_key = ?", migrationKey).Count(&count)
	if count > 0 {
		return nil
	}

	logger.Info("starting one-time migration: add review node to alert workflow")

	var wf Workflow
	if err := db.Where("name = ? AND project_id IS NULL", "告警工作流").First(&wf).Error; err != nil {
		logger.Warn("alert workflow not found, skipping migration")
		goto done
	}

	{
		// 检查是否已有"待确认"节点
		var existCount int64
		db.Model(&WorkflowNode{}).Where("workflow_id = ? AND name = ?", wf.ID, "待确认").Count(&existCount)
		if existCount > 0 {
			logger.Info("alert workflow already has review node, skipping")
			goto done
		}

		// 找到"处理中"和"完成"节点
		var workNode, endNode WorkflowNode
		if err := db.Where("workflow_id = ? AND name = ?", wf.ID, "处理中").First(&workNode).Error; err != nil {
			logger.Error("处理中 node not found", zap.Error(err))
			goto done
		}
		if err := db.Where("workflow_id = ? AND name = ? AND node_type = ?", wf.ID, "完成", "end").First(&endNode).Error; err != nil {
			logger.Error("完成 node not found", zap.Error(err))
			goto done
		}

		// 创建"待确认"节点
		reviewNode := WorkflowNode{
			WorkflowID: wf.ID,
			Name:       "待确认",
			NodeType:   "work",
			Config:     `{"target_status":"pending_review"}`,
			PositionX:  500,
			PositionY:  200,
		}
		if err := db.Create(&reviewNode).Error; err != nil {
			logger.Error("failed to create review node", zap.Error(err))
			goto done
		}

		// 调整"完成"节点位置
		db.Model(&WorkflowNode{}).Where("id = ?", endNode.ID).Update("position_x", 700)

		// 删除旧的 处理中→完成 边
		db.Where("workflow_id = ? AND source_node_id = ? AND target_node_id = ?",
			wf.ID, workNode.ID, endNode.ID).Delete(&WorkflowEdge{})

		// 创建新边：处理中→待确认（无条件，正常流转）
		db.Create(&WorkflowEdge{
			WorkflowID:   wf.ID,
			SourceNodeID: workNode.ID,
			TargetNodeID: reviewNode.ID,
		})

		// 创建条件边：待确认→完成（confirmed）
		db.Create(&WorkflowEdge{
			WorkflowID:    wf.ID,
			SourceNodeID:  reviewNode.ID,
			TargetNodeID:  endNode.ID,
			ConditionExpr: "confirmed",
		})

		// 创建条件边：待确认→处理中（continue）
		db.Create(&WorkflowEdge{
			WorkflowID:    wf.ID,
			SourceNodeID:  reviewNode.ID,
			TargetNodeID:  workNode.ID,
			ConditionExpr: "continue",
		})

		logger.Info("alert workflow review node added",
			zap.Uint64("workflow_id", wf.ID),
			zap.Uint64("review_node_id", reviewNode.ID))

		// 修复现有 pending_review 状态的告警工单：将 current_node_id 指向新的"待确认"节点
		var instances []WorkflowInstance
		db.Where("workflow_id = ? AND status IN ? AND current_node_id = ?",
			wf.ID, []string{"active", "reviewing"}, workNode.ID).Find(&instances)

		for _, inst := range instances {
			// 查工单状态
			var issue Issue
			if err := db.Where("id = ?", inst.IssueID).First(&issue).Error; err != nil {
				continue
			}
			if issue.Status == "pending_review" {
				// 工单是待确认状态，工作流应该在"待确认"节点
				db.Model(&WorkflowInstance{}).Where("id = ?", inst.ID).Updates(map[string]any{
					"current_node_id": reviewNode.ID,
					"status":          "active",
				})
				logger.Info("migrated alert workflow instance to review node",
					zap.Uint64("instance_id", inst.ID),
					zap.Uint64("issue_id", inst.IssueID))
			} else if inst.Status == "reviewing" {
				// reviewing 状态改为 active（不再使用 reviewing 状态）
				db.Model(&WorkflowInstance{}).Where("id = ?", inst.ID).Update("status", "active")
			}
		}
	}

done:
	db.Create(&SystemConfig{
		ConfigKey:   migrationKey,
		ConfigValue: "done",
		ConfigType:  "string",
		Category:    "migration",
		Description: "一次性迁移：为告警工作流添加待确认节点",
		IsSecret:    false,
	})

	logger.Info("one-time migration completed: alert workflow review node")
	return nil
}

// fixAlertIssuesWithoutWorkflow 为缺失工作流实例的告警工单补建工作流实例
func fixAlertIssuesWithoutWorkflow(db *gorm.DB, alertTypeID, workflowID uint64) {
	// 查找所有 Alert 类型且没有工作流实例的工单
	var issues []Issue
	db.Where("issue_type_id = ? AND (workflow_instance_id IS NULL OR workflow_instance_id = 0)", alertTypeID).
		Find(&issues)

	logger.Info("found alert issues without workflow instance",
		zap.Uint64("alert_type_id", alertTypeID),
		zap.Uint64("workflow_id", workflowID),
		zap.Int("count", len(issues)))

	if len(issues) == 0 {
		return
	}

	// 获取工作流的开始节点
	var startNode WorkflowNode
	if err := db.Where("workflow_id = ? AND node_type = ?", workflowID, "start").First(&startNode).Error; err != nil {
		logger.Error("failed to find start node for alert workflow", zap.Error(err))
		return
	}

	// 获取开始节点的下一个节点（处理中）
	var startEdge WorkflowEdge
	if err := db.Where("workflow_id = ? AND source_node_id = ?", workflowID, startNode.ID).First(&startEdge).Error; err != nil {
		logger.Error("failed to find start edge for alert workflow", zap.Error(err))
		return
	}
	workNodeID := startEdge.TargetNodeID

	fixed := 0
	for _, issue := range issues {
		// 根据工单状态决定工作流实例的状态和当前节点
		var instanceStatus string
		var currentNodeID uint64

		switch issue.Status {
		case "open":
			// 工单刚创建，工作流应在处理中节点
			instanceStatus = "active"
			currentNodeID = workNodeID
		case "in_progress":
			// 工单处理中
			instanceStatus = "active"
			currentNodeID = workNodeID
		case "pending_review":
			// 告警恢复后的待确认状态 → 工作流应为 reviewing
			instanceStatus = "reviewing"
			currentNodeID = workNodeID
		case "resolved", "closed":
			// 已完成的工单，跳过
			continue
		default:
			instanceStatus = "active"
			currentNodeID = workNodeID
		}

		// 创建工作流实例
		now := time.Now()
		instance := WorkflowInstance{
			WorkflowID:    workflowID,
			IssueID:       issue.ID,
			Status:        instanceStatus,
			CurrentNodeID: currentNodeID,
			StartedAt:     now,
		}
		if err := db.Create(&instance).Error; err != nil {
			logger.Error("failed to create workflow instance for alert issue",
				zap.Uint64("issue_id", issue.ID),
				zap.String("issue_key", issue.IssueKey),
				zap.Error(err))
			continue
		}

		// 更新工单的 workflow_instance_id
		db.Model(&Issue{}).Where("id = ?", issue.ID).Update("workflow_instance_id", instance.ID)

		fixed++
		logger.Info("created workflow instance for alert issue",
			zap.Uint64("issue_id", issue.ID),
			zap.String("issue_key", issue.IssueKey),
			zap.String("status", issue.Status),
			zap.String("instance_status", instanceStatus),
			zap.Uint64("instance_id", instance.ID))
	}

	if fixed > 0 {
		logger.Info("fixed alert issues without workflow instance", zap.Int("count", fixed))
	}
}
