-- 修复缺失工单号的脚本
-- 执行前请先备份数据库

-- 1. 查看缺失工单号的工单
SELECT i.id, i.issue_key, i.title, i.project_id, p.project_key
FROM issues i
LEFT JOIN projects p ON i.project_id = p.id
WHERE i.issue_key IS NULL OR i.issue_key = ''
;

-- 2. 为缺失工单号的工单生成工单号
-- 这个脚本会为每个缺失工单号的工单生成一个新的工单号
-- 格式: 项目KEY-序号

-- 首先查看当前项目的最大工单号
SELECT p.project_key, MAX(CAST(SUBSTRING_INDEX(i.issue_key, '-', -1) AS UNSIGNED)) as max_num
FROM issues i
JOIN projects p ON i.project_id = p.id
WHERE i.issue_key IS NOT NULL AND i.issue_key != ''
GROUP BY p.project_key;

-- 3. 更新缺失工单号的工单 (针对 OPS 项目)
-- 假设 test2 是 OPS 项目的工单，需要生成 OPS-8 (因为已有 OPS-7)

-- 先查看 OPS 项目当前最大的工单号
SELECT MAX(CAST(SUBSTRING_INDEX(issue_key, '-', -1) AS UNSIGNED)) as max_num
FROM issues
WHERE issue_key LIKE 'OPS-%'
;

-- 更新缺失工单号的工单
-- 注意：这个脚本假设只有一个缺失工单号的工单，如果有多个需要逐个处理

-- 方法1: 直接更新 (如果你知道具体的工单ID和应该分配的工单号)
-- UPDATE issues SET issue_key = 'OPS-8' WHERE id = <工单ID> AND (issue_key IS NULL OR issue_key = '');

-- 方法2: 使用变量自动生成 (MySQL 8.0+)
SET @project_key = 'OPS';
SET @next_num = (
    SELECT COALESCE(MAX(CAST(SUBSTRING_INDEX(issue_key, '-', -1) AS UNSIGNED)), 0) + 1
    FROM issues
    WHERE issue_key LIKE CONCAT(@project_key, '-%')
);

-- 查看将要分配的工单号
SELECT CONCAT(@project_key, '-', @next_num) as new_issue_key;

-- 执行更新 (取消注释后执行)
-- UPDATE issues
-- SET issue_key = CONCAT(@project_key, '-', @next_num)
-- WHERE (issue_key IS NULL OR issue_key = '')
-- AND project_id = (SELECT id FROM projects WHERE project_key = @project_key)
-- LIMIT 1;

-- 4. 验证修复结果
-- SELECT id, issue_key, title, project_id FROM issues ORDER BY id DESC LIMIT 10;
