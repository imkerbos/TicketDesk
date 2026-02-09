-- Epic 字段迁移脚本
-- 将 issue_field_values 表中的 epic_link 字段值迁移到 issues 表的 epic_id 列
-- 执行前请先备份数据库

-- 1. 确保 epic_id 列已存在（由 GORM AutoMigrate 自动创建）
-- ALTER TABLE issues ADD COLUMN epic_id BIGINT UNSIGNED NULL;
-- ALTER TABLE issues ADD INDEX idx_issues_epic_id (epic_id);

-- 2. 迁移数据：将 epic_link 字段值（存储的是 Epic 的 issue_id）迁移到 epic_id 列
UPDATE issues i
INNER JOIN issue_field_values fv ON i.id = fv.issue_id
INNER JOIN field_definitions fd ON fv.field_id = fd.id
SET i.epic_id = CAST(fv.value_number AS UNSIGNED)
WHERE fd.field_key = 'epic_link'
  AND fv.value_number IS NOT NULL
  AND fv.value_number > 0
  AND i.epic_id IS NULL;

-- 3. 验证迁移结果
SELECT
    COUNT(*) as total_migrated,
    (SELECT COUNT(*) FROM issue_field_values fv
     INNER JOIN field_definitions fd ON fv.field_id = fd.id
     WHERE fd.field_key = 'epic_link' AND fv.value_number IS NOT NULL) as total_epic_links,
    (SELECT COUNT(*) FROM issues WHERE epic_id IS NOT NULL) as issues_with_epic_id
FROM issues WHERE epic_id IS NOT NULL;

-- 4. 可选：清理旧数据（建议在确认迁移成功后执行）
-- DELETE fv FROM issue_field_values fv
-- INNER JOIN field_definitions fd ON fv.field_id = fd.id
-- WHERE fd.field_key = 'epic_link';

-- 5. 可选：删除 epic_link 字段定义（建议在确认迁移成功后执行）
-- DELETE FROM field_definitions WHERE field_key = 'epic_link';
