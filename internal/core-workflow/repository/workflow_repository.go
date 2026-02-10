// Package repository 提供工作流数据访问层
package repository

import (
	"context"

	"github.com/kerbos/ticketdesk/internal/model"
	"gorm.io/gorm"
)

// WorkflowRepository 工作流数据访问接口
type WorkflowRepository interface {
	Create(ctx context.Context, workflow *model.Workflow) error
	GetByID(ctx context.Context, id uint64) (*model.Workflow, error)
	Update(ctx context.Context, workflow *model.Workflow) error
	Delete(ctx context.Context, id uint64) error
	List(ctx context.Context, filter *WorkflowFilter, offset, limit int) ([]*model.Workflow, int64, error)
	GetByProjectID(ctx context.Context, projectID uint64) ([]*model.Workflow, error)
	GetGlobalWorkflows(ctx context.Context) ([]*model.Workflow, error)
}

// WorkflowFilter 工作流过滤条件
type WorkflowFilter struct {
	ProjectID *uint64
	Status    *int8
	Keyword   string
}

// workflowRepository 工作流数据访问实现
type workflowRepository struct {
	db *gorm.DB
}

// NewWorkflowRepository 创建工作流数据访问实例
func NewWorkflowRepository(db *gorm.DB) WorkflowRepository {
	return &workflowRepository{db: db}
}

// Create 创建工作流
func (r *workflowRepository) Create(ctx context.Context, workflow *model.Workflow) error {
	return r.db.WithContext(ctx).Create(workflow).Error
}

// GetByID 根据 ID 获取工作流
func (r *workflowRepository) GetByID(ctx context.Context, id uint64) (*model.Workflow, error) {
	var workflow model.Workflow
	err := r.db.WithContext(ctx).First(&workflow, id).Error
	if err != nil {
		return nil, err
	}
	return &workflow, nil
}

// Update 更新工作流
func (r *workflowRepository) Update(ctx context.Context, workflow *model.Workflow) error {
	return r.db.WithContext(ctx).Save(workflow).Error
}

// Delete 软删除工作流
func (r *workflowRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.Workflow{}, id).Error
}

// List 分页查询工作流列表
func (r *workflowRepository) List(ctx context.Context, filter *WorkflowFilter, offset, limit int) ([]*model.Workflow, int64, error) {
	var workflows []*model.Workflow
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Workflow{})

	if filter != nil {
		if filter.ProjectID != nil {
			query = query.Where("project_id = ?", *filter.ProjectID)
		}
		if filter.Status != nil {
			query = query.Where("status = ?", *filter.Status)
		}
		if filter.Keyword != "" {
			keyword := "%" + filter.Keyword + "%"
			query = query.Where("name LIKE ? OR description LIKE ?", keyword, keyword)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(limit).Order("id DESC").Find(&workflows).Error; err != nil {
		return nil, 0, err
	}

	return workflows, total, nil
}

// GetByProjectID 获取项目的工作流
func (r *workflowRepository) GetByProjectID(ctx context.Context, projectID uint64) ([]*model.Workflow, error) {
	var workflows []*model.Workflow
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Find(&workflows).Error
	return workflows, err
}

// GetGlobalWorkflows 获取全局工作流
func (r *workflowRepository) GetGlobalWorkflows(ctx context.Context) ([]*model.Workflow, error) {
	var workflows []*model.Workflow
	err := r.db.WithContext(ctx).Where("project_id IS NULL").Find(&workflows).Error
	return workflows, err
}

// NodeRepository 节点数据访问接口
type NodeRepository interface {
	Create(ctx context.Context, node *model.WorkflowNode) error
	GetByID(ctx context.Context, id uint64) (*model.WorkflowNode, error)
	Update(ctx context.Context, node *model.WorkflowNode) error
	Delete(ctx context.Context, id uint64) error
	ListByWorkflow(ctx context.Context, workflowID uint64) ([]*model.WorkflowNode, error)
	GetStartNode(ctx context.Context, workflowID uint64) (*model.WorkflowNode, error)
	GetEndNodes(ctx context.Context, workflowID uint64) ([]*model.WorkflowNode, error)
	DeleteByWorkflow(ctx context.Context, workflowID uint64) error
}

// nodeRepository 节点数据访问实现
type nodeRepository struct {
	db *gorm.DB
}

// NewNodeRepository 创建节点数据访问实例
func NewNodeRepository(db *gorm.DB) NodeRepository {
	return &nodeRepository{db: db}
}

// Create 创建节点
func (r *nodeRepository) Create(ctx context.Context, node *model.WorkflowNode) error {
	return r.db.WithContext(ctx).Create(node).Error
}

// GetByID 根据 ID 获取节点
func (r *nodeRepository) GetByID(ctx context.Context, id uint64) (*model.WorkflowNode, error) {
	var node model.WorkflowNode
	err := r.db.WithContext(ctx).First(&node, id).Error
	if err != nil {
		return nil, err
	}
	return &node, nil
}

// Update 更新节点
func (r *nodeRepository) Update(ctx context.Context, node *model.WorkflowNode) error {
	return r.db.WithContext(ctx).Save(node).Error
}

// Delete 删除节点
func (r *nodeRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.WorkflowNode{}, id).Error
}

// ListByWorkflow 获取工作流的所有节点
func (r *nodeRepository) ListByWorkflow(ctx context.Context, workflowID uint64) ([]*model.WorkflowNode, error) {
	var nodes []*model.WorkflowNode
	err := r.db.WithContext(ctx).Where("workflow_id = ?", workflowID).Order("id ASC").Find(&nodes).Error
	return nodes, err
}

// GetStartNode 获取开始节点
func (r *nodeRepository) GetStartNode(ctx context.Context, workflowID uint64) (*model.WorkflowNode, error) {
	var node model.WorkflowNode
	err := r.db.WithContext(ctx).Where("workflow_id = ? AND node_type = ?", workflowID, "start").First(&node).Error
	if err != nil {
		return nil, err
	}
	return &node, nil
}

// GetEndNodes 获取结束节点
func (r *nodeRepository) GetEndNodes(ctx context.Context, workflowID uint64) ([]*model.WorkflowNode, error) {
	var nodes []*model.WorkflowNode
	err := r.db.WithContext(ctx).Where("workflow_id = ? AND node_type = ?", workflowID, "end").Find(&nodes).Error
	return nodes, err
}

// DeleteByWorkflow 删除工作流的所有节点
func (r *nodeRepository) DeleteByWorkflow(ctx context.Context, workflowID uint64) error {
	return r.db.WithContext(ctx).Where("workflow_id = ?", workflowID).Delete(&model.WorkflowNode{}).Error
}

// EdgeRepository 边数据访问接口
type EdgeRepository interface {
	Create(ctx context.Context, edge *model.WorkflowEdge) error
	GetByID(ctx context.Context, id uint64) (*model.WorkflowEdge, error)
	Update(ctx context.Context, edge *model.WorkflowEdge) error
	Delete(ctx context.Context, id uint64) error
	ListByWorkflow(ctx context.Context, workflowID uint64) ([]*model.WorkflowEdge, error)
	GetOutgoingEdges(ctx context.Context, nodeID uint64) ([]*model.WorkflowEdge, error)
	GetIncomingEdges(ctx context.Context, nodeID uint64) ([]*model.WorkflowEdge, error)
	DeleteByWorkflow(ctx context.Context, workflowID uint64) error
	DeleteByNode(ctx context.Context, nodeID uint64) error
}

// edgeRepository 边数据访问实现
type edgeRepository struct {
	db *gorm.DB
}

// NewEdgeRepository 创建边数据访问实例
func NewEdgeRepository(db *gorm.DB) EdgeRepository {
	return &edgeRepository{db: db}
}

// Create 创建边
func (r *edgeRepository) Create(ctx context.Context, edge *model.WorkflowEdge) error {
	return r.db.WithContext(ctx).Create(edge).Error
}

// GetByID 根据 ID 获取边
func (r *edgeRepository) GetByID(ctx context.Context, id uint64) (*model.WorkflowEdge, error) {
	var edge model.WorkflowEdge
	err := r.db.WithContext(ctx).First(&edge, id).Error
	if err != nil {
		return nil, err
	}
	return &edge, nil
}

// Update 更新边
func (r *edgeRepository) Update(ctx context.Context, edge *model.WorkflowEdge) error {
	return r.db.WithContext(ctx).Save(edge).Error
}

// Delete 删除边
func (r *edgeRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.WorkflowEdge{}, id).Error
}

// ListByWorkflow 获取工作流的所有边
func (r *edgeRepository) ListByWorkflow(ctx context.Context, workflowID uint64) ([]*model.WorkflowEdge, error) {
	var edges []*model.WorkflowEdge
	err := r.db.WithContext(ctx).Where("workflow_id = ?", workflowID).Order("id ASC").Find(&edges).Error
	return edges, err
}

// GetOutgoingEdges 获取节点的出边
func (r *edgeRepository) GetOutgoingEdges(ctx context.Context, nodeID uint64) ([]*model.WorkflowEdge, error) {
	var edges []*model.WorkflowEdge
	err := r.db.WithContext(ctx).Where("source_node_id = ?", nodeID).Find(&edges).Error
	return edges, err
}

// GetIncomingEdges 获取节点的入边
func (r *edgeRepository) GetIncomingEdges(ctx context.Context, nodeID uint64) ([]*model.WorkflowEdge, error) {
	var edges []*model.WorkflowEdge
	err := r.db.WithContext(ctx).Where("target_node_id = ?", nodeID).Find(&edges).Error
	return edges, err
}

// DeleteByWorkflow 删除工作流的所有边
func (r *edgeRepository) DeleteByWorkflow(ctx context.Context, workflowID uint64) error {
	return r.db.WithContext(ctx).Where("workflow_id = ?", workflowID).Delete(&model.WorkflowEdge{}).Error
}

// DeleteByNode 删除与节点相关的所有边
func (r *edgeRepository) DeleteByNode(ctx context.Context, nodeID uint64) error {
	return r.db.WithContext(ctx).Where("source_node_id = ? OR target_node_id = ?", nodeID, nodeID).Delete(&model.WorkflowEdge{}).Error
}

// WorkflowInstanceRepository 工作流实例数据访问接口
type WorkflowInstanceRepository interface {
	Create(ctx context.Context, instance *model.WorkflowInstance) error
	GetByID(ctx context.Context, id uint64) (*model.WorkflowInstance, error)
	GetByIssueID(ctx context.Context, issueID uint64) (*model.WorkflowInstance, error)
	Update(ctx context.Context, instance *model.WorkflowInstance) error
	ListByWorkflow(ctx context.Context, workflowID uint64) ([]*model.WorkflowInstance, error)
}

// workflowInstanceRepository 工作流实例数据访问实现
type workflowInstanceRepository struct {
	db *gorm.DB
}

// NewWorkflowInstanceRepository 创建工作流实例数据访问实例
func NewWorkflowInstanceRepository(db *gorm.DB) WorkflowInstanceRepository {
	return &workflowInstanceRepository{db: db}
}

// Create 创建工作流实例
func (r *workflowInstanceRepository) Create(ctx context.Context, instance *model.WorkflowInstance) error {
	return r.db.WithContext(ctx).Create(instance).Error
}

// GetByID 根据 ID 获取工作流实例
func (r *workflowInstanceRepository) GetByID(ctx context.Context, id uint64) (*model.WorkflowInstance, error) {
	var instance model.WorkflowInstance
	err := r.db.WithContext(ctx).First(&instance, id).Error
	if err != nil {
		return nil, err
	}
	return &instance, nil
}

// GetByIssueID 根据工单 ID 获取工作流实例
func (r *workflowInstanceRepository) GetByIssueID(ctx context.Context, issueID uint64) (*model.WorkflowInstance, error) {
	var instance model.WorkflowInstance
	err := r.db.WithContext(ctx).Where("issue_id = ?", issueID).First(&instance).Error
	if err != nil {
		return nil, err
	}
	return &instance, nil
}

// Update 更新工作流实例
func (r *workflowInstanceRepository) Update(ctx context.Context, instance *model.WorkflowInstance) error {
	return r.db.WithContext(ctx).Save(instance).Error
}

// ListByWorkflow 获取工作流的所有实例
func (r *workflowInstanceRepository) ListByWorkflow(ctx context.Context, workflowID uint64) ([]*model.WorkflowInstance, error) {
	var instances []*model.WorkflowInstance
	err := r.db.WithContext(ctx).Where("workflow_id = ?", workflowID).Order("id DESC").Find(&instances).Error
	return instances, err
}

// WorkflowHistoryRepository 流转历史数据访问接口
type WorkflowHistoryRepository interface {
	Create(ctx context.Context, history *model.WorkflowHistory) error
	GetByID(ctx context.Context, id uint64) (*model.WorkflowHistory, error)
	ListByInstance(ctx context.Context, instanceID uint64) ([]*model.WorkflowHistory, error)
}

// workflowHistoryRepository 流转历史数据访问实现
type workflowHistoryRepository struct {
	db *gorm.DB
}

// NewWorkflowHistoryRepository 创建流转历史数据访问实例
func NewWorkflowHistoryRepository(db *gorm.DB) WorkflowHistoryRepository {
	return &workflowHistoryRepository{db: db}
}

// Create 创建流转历史
func (r *workflowHistoryRepository) Create(ctx context.Context, history *model.WorkflowHistory) error {
	return r.db.WithContext(ctx).Create(history).Error
}

// GetByID 根据 ID 获取流转历史
func (r *workflowHistoryRepository) GetByID(ctx context.Context, id uint64) (*model.WorkflowHistory, error) {
	var history model.WorkflowHistory
	err := r.db.WithContext(ctx).First(&history, id).Error
	if err != nil {
		return nil, err
	}
	return &history, nil
}

// ListByInstance 获取实例的所有流转历史
func (r *workflowHistoryRepository) ListByInstance(ctx context.Context, instanceID uint64) ([]*model.WorkflowHistory, error) {
	var histories []*model.WorkflowHistory
	err := r.db.WithContext(ctx).Where("instance_id = ?", instanceID).Order("operated_at ASC").Find(&histories).Error
	return histories, err
}

// ApprovalRecordRepository 审批记录数据访问接口
type ApprovalRecordRepository interface {
	Create(ctx context.Context, record *model.ApprovalRecord) error
	GetByID(ctx context.Context, id uint64) (*model.ApprovalRecord, error)
	Update(ctx context.Context, record *model.ApprovalRecord) error
	ListByInstance(ctx context.Context, instanceID uint64) ([]*model.ApprovalRecord, error)
	ListByInstanceAndNode(ctx context.Context, instanceID, nodeID uint64) ([]*model.ApprovalRecord, error)
	GetByInstanceNodeAndApprover(ctx context.Context, instanceID, nodeID, approverID uint64) (*model.ApprovalRecord, error)
}

// approvalRecordRepository 审批记录数据访问实现
type approvalRecordRepository struct {
	db *gorm.DB
}

// NewApprovalRecordRepository 创建审批记录数据访问实例
func NewApprovalRecordRepository(db *gorm.DB) ApprovalRecordRepository {
	return &approvalRecordRepository{db: db}
}

// Create 创建审批记录
func (r *approvalRecordRepository) Create(ctx context.Context, record *model.ApprovalRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

// GetByID 根据 ID 获取审批记录
func (r *approvalRecordRepository) GetByID(ctx context.Context, id uint64) (*model.ApprovalRecord, error) {
	var record model.ApprovalRecord
	err := r.db.WithContext(ctx).First(&record, id).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// Update 更新审批记录
func (r *approvalRecordRepository) Update(ctx context.Context, record *model.ApprovalRecord) error {
	return r.db.WithContext(ctx).Save(record).Error
}

// ListByInstance 获取实例的所有审批记录
func (r *approvalRecordRepository) ListByInstance(ctx context.Context, instanceID uint64) ([]*model.ApprovalRecord, error) {
	var records []*model.ApprovalRecord
	err := r.db.WithContext(ctx).Where("instance_id = ?", instanceID).Order("id ASC").Find(&records).Error
	return records, err
}

// ListByInstanceAndNode 获取实例特定节点的所有审批记录
func (r *approvalRecordRepository) ListByInstanceAndNode(ctx context.Context, instanceID, nodeID uint64) ([]*model.ApprovalRecord, error) {
	var records []*model.ApprovalRecord
	err := r.db.WithContext(ctx).Where("instance_id = ? AND node_id = ?", instanceID, nodeID).Order("id ASC").Find(&records).Error
	return records, err
}

// GetByInstanceNodeAndApprover 根据实例、节点和审批人获取审批记录
func (r *approvalRecordRepository) GetByInstanceNodeAndApprover(ctx context.Context, instanceID, nodeID, approverID uint64) (*model.ApprovalRecord, error) {
	var record model.ApprovalRecord
	err := r.db.WithContext(ctx).Where("instance_id = ? AND node_id = ? AND approver_id = ?", instanceID, nodeID, approverID).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// WorkflowSchemeRepository 工作流方案数据访问接口
type WorkflowSchemeRepository interface {
	Create(ctx context.Context, scheme *model.WorkflowScheme) error
	GetByID(ctx context.Context, id uint64) (*model.WorkflowScheme, error)
	GetByProjectAndIssueType(ctx context.Context, projectID, issueTypeID uint64) (*model.WorkflowScheme, error)
	Update(ctx context.Context, scheme *model.WorkflowScheme) error
	Delete(ctx context.Context, id uint64) error
	DeleteByProjectAndIssueType(ctx context.Context, projectID, issueTypeID uint64) error
	HardDeleteByProjectAndIssueType(ctx context.Context, projectID, issueTypeID uint64) error
	ListByProject(ctx context.Context, projectID uint64) ([]*model.WorkflowScheme, error)
}

// workflowSchemeRepository 工作流方案数据访问实现
type workflowSchemeRepository struct {
	db *gorm.DB
}

// NewWorkflowSchemeRepository 创建工作流方案数据访问实例
func NewWorkflowSchemeRepository(db *gorm.DB) WorkflowSchemeRepository {
	return &workflowSchemeRepository{db: db}
}

// Create 创建工作流方案
func (r *workflowSchemeRepository) Create(ctx context.Context, scheme *model.WorkflowScheme) error {
	return r.db.WithContext(ctx).Create(scheme).Error
}

// GetByID 根据 ID 获取工作流方案
func (r *workflowSchemeRepository) GetByID(ctx context.Context, id uint64) (*model.WorkflowScheme, error) {
	var scheme model.WorkflowScheme
	err := r.db.WithContext(ctx).First(&scheme, id).Error
	if err != nil {
		return nil, err
	}
	return &scheme, nil
}

// GetByProjectAndIssueType 根据项目和工单类型获取工作流方案
func (r *workflowSchemeRepository) GetByProjectAndIssueType(ctx context.Context, projectID, issueTypeID uint64) (*model.WorkflowScheme, error) {
	var scheme model.WorkflowScheme
	err := r.db.WithContext(ctx).Where("project_id = ? AND issue_type_id = ?", projectID, issueTypeID).First(&scheme).Error
	if err != nil {
		return nil, err
	}
	return &scheme, nil
}

// Update 更新工作流方案
func (r *workflowSchemeRepository) Update(ctx context.Context, scheme *model.WorkflowScheme) error {
	return r.db.WithContext(ctx).Save(scheme).Error
}

// Delete 删除工作流方案
func (r *workflowSchemeRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.WorkflowScheme{}, id).Error
}

// DeleteByProjectAndIssueType 根据项目和工单类型删除工作流方案
func (r *workflowSchemeRepository) DeleteByProjectAndIssueType(ctx context.Context, projectID, issueTypeID uint64) error {
	return r.db.WithContext(ctx).Where("project_id = ? AND issue_type_id = ?", projectID, issueTypeID).Delete(&model.WorkflowScheme{}).Error
}

// HardDeleteByProjectAndIssueType 根据项目和工单类型硬删除工作流方案（包括软删除的记录）
func (r *workflowSchemeRepository) HardDeleteByProjectAndIssueType(ctx context.Context, projectID, issueTypeID uint64) error {
	return r.db.WithContext(ctx).Unscoped().Where("project_id = ? AND issue_type_id = ? AND deleted_at IS NOT NULL", projectID, issueTypeID).Delete(&model.WorkflowScheme{}).Error
}

// ListByProject 获取项目的所有工作流方案
func (r *workflowSchemeRepository) ListByProject(ctx context.Context, projectID uint64) ([]*model.WorkflowScheme, error) {
	var schemes []*model.WorkflowScheme
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Order("id ASC").Find(&schemes).Error
	return schemes, err
}

