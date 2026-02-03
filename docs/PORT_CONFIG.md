# TicketDesk 端口配置说明

## 端口配置

### 后端端口

后端端口在 `configs/config.yaml` 中配置：

```yaml
app:
  name: ticketdesk
  env: development
  port: 10010  # 后端服务端口
  debug: true
```

**默认端口：10010**

### 前端端口

前端端口在 `web/vite.config.ts` 中配置：

```typescript
server: {
  port: 3100,  // 前端开发服务器端口
  strictPort: false,  // 端口被占用时自动尝试下一个
  proxy: {
    '/api': {
      target: 'http://localhost:10010',  // 代理到后端
      changeOrigin: true,
    },
  },
}
```

**默认端口：3100**（如果被占用会自动使用 3101、3102...）

## 修改端口

### 修改后端端口

1. 编辑 `configs/config.yaml`：
```yaml
app:
  port: 你的端口号
```

2. 编辑 `web/vite.config.ts`，更新代理目标：
```typescript
proxy: {
  '/api': {
    target: 'http://localhost:你的端口号',
    changeOrigin: true,
  },
}
```

### 修改前端端口

编辑 `web/vite.config.ts`：
```typescript
server: {
  port: 你的端口号,
  // ...
}
```

## 访问地址

启动开发服务器后：

- **后端 API**: http://localhost:10010
- **前端页面**: http://localhost:3100
- **API 文档**: http://localhost:10010/swagger/index.html

## 注意事项

1. **端口冲突**：前端配置了 `strictPort: false`，端口被占用时会自动尝试下一个可用端口
2. **代理配置**：前端修改端口后，确保 Vite 代理配置指向正确的后端端口
3. **防火墙**：确保防火墙允许访问这些端口
4. **生产环境**：生产环境建议使用标准端口（80/443）或通过 Nginx 反向代理

## 端口检查

检查端口是否被占用：

```bash
# macOS/Linux
lsof -i :10010
lsof -i :3000

# Windows
netstat -ano | findstr :10010
netstat -ano | findstr :3000
```

## 开发脚本

开发脚本 `scripts/dev.sh` 会自动从配置文件读取后端端口：

```bash
# 从 config.yaml 读取端口
BACKEND_PORT=$(grep -A1 "^app:" "$CONFIG_FILE" | grep "port:" | awk '{print $2}')
```

因此修改 `configs/config.yaml` 中的端口后，开发脚本会自动使用新端口。
