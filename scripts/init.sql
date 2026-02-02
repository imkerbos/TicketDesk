-- TicketDesk 数据库初始化脚本

-- 创建数据库（如果不存在）
CREATE DATABASE IF NOT EXISTS ticketdesk
    CHARACTER SET utf8mb4
    COLLATE utf8mb4_unicode_ci;

USE ticketdesk;

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    display_name VARCHAR(100),
    avatar_url VARCHAR(255),
    status TINYINT DEFAULT 1 COMMENT '状态: 0-禁用, 1-启用',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    INDEX idx_users_username (username),
    INDEX idx_users_email (email),
    INDEX idx_users_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 角色表
CREATE TABLE IF NOT EXISTS roles (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    display_name VARCHAR(100),
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    INDEX idx_roles_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色表';

-- 用户角色关联表
CREATE TABLE IF NOT EXISTS user_roles (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    role_id BIGINT UNSIGNED NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_user_role (user_id, role_id),
    INDEX idx_user_roles_user_id (user_id),
    INDEX idx_user_roles_role_id (role_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户角色关联表';

-- 项目表
CREATE TABLE IF NOT EXISTS projects (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    project_key VARCHAR(20) NOT NULL UNIQUE COMMENT '项目标识，如 OPS-INFRA',
    name VARCHAR(100) NOT NULL,
    description TEXT,
    lead_user_id BIGINT UNSIGNED COMMENT '项目负责人',
    status TINYINT DEFAULT 1 COMMENT '状态: 0-归档, 1-活跃',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    INDEX idx_projects_key (project_key),
    INDEX idx_projects_status (status),
    INDEX idx_projects_lead (lead_user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='项目表';

-- 工单类型表
CREATE TABLE IF NOT EXISTS issue_types (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    project_id BIGINT UNSIGNED COMMENT '所属项目，NULL 表示全局类型',
    name VARCHAR(50) NOT NULL COMMENT '类型名称: Epic, Task, Bug, Fault, Change, ServiceRequest',
    display_name VARCHAR(100),
    description TEXT,
    icon VARCHAR(50),
    color VARCHAR(20),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    INDEX idx_issue_types_project (project_id),
    INDEX idx_issue_types_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工单类型表';

-- 工单表
CREATE TABLE IF NOT EXISTS issues (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    issue_key VARCHAR(30) NOT NULL UNIQUE COMMENT '工单编号，如 OPS-123',
    project_id BIGINT UNSIGNED NOT NULL,
    issue_type_id BIGINT UNSIGNED NOT NULL,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    priority VARCHAR(10) DEFAULT 'P2' COMMENT '优先级: P0, P1, P2, P3',
    status VARCHAR(30) DEFAULT 'open' COMMENT '状态',
    reporter_id BIGINT UNSIGNED NOT NULL COMMENT '报告人',
    assignee_id BIGINT UNSIGNED COMMENT '指派人',
    parent_id BIGINT UNSIGNED COMMENT '父工单 ID',
    due_date DATETIME COMMENT '截止日期',
    resolved_at DATETIME COMMENT '解决时间',
    closed_at DATETIME COMMENT '关闭时间',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    INDEX idx_issues_key (issue_key),
    INDEX idx_issues_project (project_id),
    INDEX idx_issues_type (issue_type_id),
    INDEX idx_issues_status (status),
    INDEX idx_issues_priority (priority),
    INDEX idx_issues_reporter (reporter_id),
    INDEX idx_issues_assignee (assignee_id),
    INDEX idx_issues_parent (parent_id),
    INDEX idx_issues_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工单表';

-- 工单评论表
CREATE TABLE IF NOT EXISTS issue_comments (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    issue_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    INDEX idx_issue_comments_issue (issue_id),
    INDEX idx_issue_comments_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工单评论表';

-- 工单关注人表
CREATE TABLE IF NOT EXISTS issue_watchers (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    issue_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_issue_watcher (issue_id, user_id),
    INDEX idx_issue_watchers_issue (issue_id),
    INDEX idx_issue_watchers_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工单关注人表';

-- 工作流定义表
CREATE TABLE IF NOT EXISTS workflows (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    project_id BIGINT UNSIGNED COMMENT '所属项目',
    name VARCHAR(100) NOT NULL,
    description TEXT,
    status TINYINT DEFAULT 1 COMMENT '状态: 0-禁用, 1-启用',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    INDEX idx_workflows_project (project_id),
    INDEX idx_workflows_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工作流定义表';

-- 工作流节点表
CREATE TABLE IF NOT EXISTS workflow_nodes (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    workflow_id BIGINT UNSIGNED NOT NULL,
    name VARCHAR(100) NOT NULL,
    node_type VARCHAR(20) NOT NULL COMMENT '节点类型: start, end, approval, work, system',
    config JSON COMMENT '节点配置',
    position_x INT DEFAULT 0,
    position_y INT DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    INDEX idx_workflow_nodes_workflow (workflow_id),
    INDEX idx_workflow_nodes_type (node_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工作流节点表';

-- 工作流边（流转）表
CREATE TABLE IF NOT EXISTS workflow_edges (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    workflow_id BIGINT UNSIGNED NOT NULL,
    source_node_id BIGINT UNSIGNED NOT NULL,
    target_node_id BIGINT UNSIGNED NOT NULL,
    condition_expr VARCHAR(500) COMMENT '条件表达式',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    INDEX idx_workflow_edges_workflow (workflow_id),
    INDEX idx_workflow_edges_source (source_node_id),
    INDEX idx_workflow_edges_target (target_node_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工作流边表';

-- 告警表
CREATE TABLE IF NOT EXISTS alerts (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    fingerprint VARCHAR(64) NOT NULL COMMENT '告警指纹',
    source VARCHAR(50) NOT NULL COMMENT '告警来源: prometheus, cloudwatch, etc.',
    alert_name VARCHAR(200) NOT NULL,
    severity VARCHAR(20) DEFAULT 'warning' COMMENT '严重级别: critical, warning, info',
    status VARCHAR(20) DEFAULT 'firing' COMMENT '状态: firing, resolved',
    labels JSON COMMENT '标签',
    annotations JSON COMMENT '注解',
    starts_at DATETIME NOT NULL,
    ends_at DATETIME,
    issue_id BIGINT UNSIGNED COMMENT '关联的工单 ID',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_alerts_fingerprint (fingerprint),
    INDEX idx_alerts_source (source),
    INDEX idx_alerts_status (status),
    INDEX idx_alerts_severity (severity),
    INDEX idx_alerts_issue (issue_id),
    INDEX idx_alerts_starts (starts_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='告警表';

-- 插入默认角色
INSERT INTO roles (name, display_name, description) VALUES
    ('admin', '管理员', '系统管理员，拥有所有权限'),
    ('user', '普通用户', '普通用户，基本操作权限'),
    ('project_admin', '项目管理员', '项目管理员，管理项目相关配置')
ON DUPLICATE KEY UPDATE display_name = VALUES(display_name);

-- 插入默认工单类型
INSERT INTO issue_types (name, display_name, description, icon, color) VALUES
    ('Epic', 'Epic', '阶段性目标/大型需求', 'lightning', '#6554C0'),
    ('Task', '任务', '普通任务', 'check-square', '#4FADE6'),
    ('Bug', '缺陷', '研发缺陷', 'bug', '#E5493A'),
    ('Fault', '故障', '生产故障/告警工单', 'alert-triangle', '#FF5630'),
    ('Change', '变更', '变更工单', 'git-branch', '#36B37E'),
    ('ServiceRequest', '服务请求', '服务申请', 'help-circle', '#00B8D9')
ON DUPLICATE KEY UPDATE display_name = VALUES(display_name);
