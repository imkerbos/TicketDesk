# 忘记密码功能配置说明

## 功能概述

忘记密码功能允许用户通过邮箱重置密码。系统会发送包含重置链接的邮件到用户注册邮箱。

## 配置步骤

### 1. 配置邮件服务

在系统设置 -> 邮件配置中配置 SMTP 服务器信息：

- SMTP 服务器地址
- SMTP 端口
- 用户名和密码
- 发件人地址和名称
- 是否使用 TLS

### 2. 配置站点域名

在系统设置 -> 通用配置中配置站点域名：

- 站点域名：填写完整的域名（包含协议），例如：`https://ticketdesk.example.com`
- 不要以斜杠结尾

**重要说明：**
- 站点域名用于生成邮件中的重置密码链接
- 如果未配置，系统将使用默认值 `http://localhost:5173`
- 生产环境必须配置正确的域名

### 3. 初始化配置

**自动初始化（推荐）：**

系统在首次启动时会自动初始化站点域名配置，默认值为 `http://localhost:5173`。

如果需要修改，只需在系统设置 -> 通用配置中更新即可。

**手动初始化（可选）：**

如果需要通过 SQL 手动初始化站点域名配置，可以执行：

```bash
mysql -u root -p ticketdesk < scripts/init_site_url_config.sql
```

或者手动插入配置：

```sql
INSERT INTO system_configs (config_key, config_value, config_type, category, description, is_secret, created_at, updated_at)
VALUES ('general.site_url', 'https://your-domain.com', 'string', 'general', '站点域名（用于生成邮件中的链接）', false, NOW(), NOW());
```

## 使用流程

1. 用户点击登录页的"忘记密码？"链接
2. 进入忘记密码页面，输入注册邮箱
3. 系统生成重置密码令牌（有效期 30 分钟）
4. 系统发送包含重置链接的邮件到用户邮箱
5. 用户点击邮件中的链接，跳转到重置密码页面
6. 系统验证令牌有效性
7. 用户输入新密码并确认
8. 密码重置成功，跳转到登录页

## 安全特性

1. **令牌安全**
   - 使用 32 字节随机令牌（64 位十六进制字符串）
   - 令牌有效期 30 分钟
   - 令牌使用后自动失效

2. **隐私保护**
   - 即使邮箱不存在也返回成功消息，不泄露用户信息
   - 禁用的用户也不会泄露状态

3. **邮件内容**
   - 精美的 HTML 格式邮件
   - 包含安全提示
   - 明确标注有效期

## API 接口

### 1. 请求重置密码

```
POST /api/v1/auth/forgot-password
Content-Type: application/json

{
  "email": "user@example.com"
}
```

### 2. 验证重置令牌

```
GET /api/v1/auth/verify-reset-token?token=<token>
```

### 3. 重置密码

```
POST /api/v1/auth/reset-password
Content-Type: application/json

{
  "token": "<token>",
  "new_password": "newpassword123"
}
```

## 前端路由

- `/forgot-password` - 忘记密码页面
- `/reset-password?token=<token>` - 重置密码页面

## 数据库变更

在 `users` 表中添加了两个字段：

- `reset_password_token` - 重置密码令牌
- `reset_password_expires` - 令牌过期时间

这些字段会在应用启动时自动通过 `AutoMigrate` 创建。

## 故障排查

### 邮件发送失败

1. 检查邮件服务配置是否正确
2. 检查 SMTP 服务器是否可访问
3. 检查用户名和密码是否正确
4. 查看后端日志获取详细错误信息

### 重置链接无效

1. 检查站点域名配置是否正确
2. 检查令牌是否已过期（30 分钟）
3. 检查令牌是否已被使用

### 前端无法访问重置页面

1. 检查前端路由配置
2. 检查前端构建是否成功
3. 检查浏览器控制台是否有错误
