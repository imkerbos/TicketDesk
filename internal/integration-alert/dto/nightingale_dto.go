// Package dto 定义夜莺告警模块的数据传输对象
package dto

import (
	"fmt"
	"strings"
	"time"
)

// N9eAlertEvent 夜莺告警事件
type N9eAlertEvent struct {
	ID           int64             `json:"id"`
	Hash         string            `json:"hash"`
	RuleID       int64             `json:"rule_id"`
	RuleName     string            `json:"rule_name"`
	RuleNote     string            `json:"rule_note"`
	Severity     int               `json:"severity"` // 1=紧急 2=警告 3=提醒
	TargetIdent  string            `json:"target_ident"`
	TargetNote   string            `json:"target_note"`
	TriggerTime  int64             `json:"trigger_time"`
	TriggerValue string            `json:"trigger_value"`
	Tags         []string          `json:"tags"`
	TagsMap      map[string]string `json:"tags_map"`
	Annotations  map[string]string `json:"annotations"`
	IsRecovered  bool              `json:"is_recovered"`
	RecoverTime  int64             `json:"recover_time"`
	GroupName    string            `json:"group_name"`
	PromQl       string            `json:"prom_ql"`
	RunbookURL   string            `json:"runbook_url"`
}

// N9eAlertListResponse 夜莺活跃告警列表 API 响应
type N9eAlertListResponse struct {
	Dat struct {
		List  []N9eAlertEvent `json:"list"`
		Total int64           `json:"total"`
	} `json:"dat"`
	Err string `json:"err"`
}

// ToAlertWebhookItem 将夜莺告警事件转换为 AlertWebhookAlertItem
func (e *N9eAlertEvent) ToAlertWebhookItem() *AlertWebhookAlertItem {
	// 构建 labels
	labels := make(map[string]string)

	// 合并 tags_map 到 labels
	for k, v := range e.TagsMap {
		labels[k] = v
	}

	// 解析 tags 数组（格式 "key=value"）合并到 labels
	for _, tag := range e.Tags {
		parts := strings.SplitN(tag, "=", 2)
		if len(parts) == 2 {
			if _, exists := labels[parts[0]]; !exists {
				labels[parts[0]] = parts[1]
			}
		}
	}

	// 设置 alertname
	labels["alertname"] = e.RuleName

	// severity 映射：1=紧急→critical, 2=警告→warning, 3=提醒→info
	switch e.Severity {
	case 1:
		labels["severity"] = "critical"
	case 2:
		labels["severity"] = "warning"
	case 3:
		labels["severity"] = "info"
	default:
		labels["severity"] = "warning"
	}

	// 设置 target_ident
	if e.TargetIdent != "" {
		labels["target_ident"] = e.TargetIdent
	}

	// 设置 group_name
	if e.GroupName != "" {
		labels["group_name"] = e.GroupName
	}

	// 构建 annotations
	annotations := make(map[string]string)
	for k, v := range e.Annotations {
		annotations[k] = v
	}

	// 将 rule_note 合并到 description
	if e.RuleNote != "" {
		if desc, ok := annotations["description"]; ok {
			annotations["description"] = desc + "\n" + e.RuleNote
		} else {
			annotations["description"] = e.RuleNote
		}
	}

	// 设置 summary（如果没有）
	if _, ok := annotations["summary"]; !ok {
		summary := e.RuleName
		if e.TriggerValue != "" {
			summary = fmt.Sprintf("%s (当前值: %s)", e.RuleName, e.TriggerValue)
		}
		annotations["summary"] = summary
	}

	// 设置 runbook_url
	if e.RunbookURL != "" {
		annotations["runbook_url"] = e.RunbookURL
	}

	// 状态映射
	status := "firing"
	if e.IsRecovered {
		status = "resolved"
	}

	// 时间转换
	startsAt := time.Unix(e.TriggerTime, 0)

	var endsAt *time.Time
	if e.IsRecovered && e.RecoverTime > 0 {
		t := time.Unix(e.RecoverTime, 0)
		endsAt = &t
	}

	// 指纹：使用夜莺的 hash
	fingerprint := e.Hash
	if fingerprint == "" {
		fingerprint = fmt.Sprintf("n9e-%d-%d", e.RuleID, e.ID)
	}

	return &AlertWebhookAlertItem{
		Status:      status,
		Labels:      labels,
		Annotations: annotations,
		StartsAt:    startsAt,
		EndsAt:      endsAt,
		Fingerprint: fingerprint,
	}
}
