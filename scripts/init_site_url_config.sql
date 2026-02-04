-- 初始化站点域名配置
-- 这个脚本会在系统配置表中插入站点域名配置项

INSERT INTO system_configs (config_key, config_value, config_type, category, description, is_secret, created_at, updated_at)
VALUES
    ('general.site_url', 'http://localhost:5173', 'string', 'general', '站点域名（用于生成邮件中的链接）', false, NOW(), NOW())
ON DUPLICATE KEY UPDATE
    description = '站点域名（用于生成邮件中的链接）',
    category = 'general',
    config_type = 'string';
