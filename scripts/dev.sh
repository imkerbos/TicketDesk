#!/bin/bash

# TicketDesk 开发环境启动脚本
# 同时启动后端（Air 热更新）和前端（Vite 热更新）

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 配置文件路径
CONFIG_FILE="configs/config.yaml"

# 从配置文件读取端口
BACKEND_PORT=$(grep -A1 "^app:" "$CONFIG_FILE" | grep "port:" | awk '{print $2}' | tr -d '\r')
if [ -z "$BACKEND_PORT" ]; then
    BACKEND_PORT=10010
fi
FRONTEND_PORT=3100

# 日志函数
log_info() {
    echo -e "${CYAN}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 清理函数
cleanup() {
    log_info "正在停止所有服务..."

    # 停止后端
    if [ ! -z "$BACKEND_PID" ]; then
        kill $BACKEND_PID 2>/dev/null || true
        log_info "后端服务已停止"
    fi

    # 停止前端
    if [ ! -z "$FRONTEND_PID" ]; then
        kill $FRONTEND_PID 2>/dev/null || true
        log_info "前端服务已停止"
    fi

    # 清理 air 进程
    pkill -f "air -c .air.toml" 2>/dev/null || true

    # 清理 vite 进程
    pkill -f "vite" 2>/dev/null || true

    log_success "所有服务已停止"
    exit 0
}

# 捕获退出信号
trap cleanup SIGINT SIGTERM EXIT

# 打印 Banner
print_banner() {
    echo -e "${BLUE}"
    echo "╔════════════════════════════════════════════════════════════╗"
    echo "║                                                            ║"
    echo "║   ████████╗██╗ ██████╗██╗  ██╗███████╗████████╗            ║"
    echo "║   ╚══██╔══╝██║██╔════╝██║ ██╔╝██╔════╝╚══██╔══╝            ║"
    echo "║      ██║   ██║██║     █████╔╝ █████╗     ██║               ║"
    echo "║      ██║   ██║██║     ██╔═██╗ ██╔══╝     ██║               ║"
    echo "║      ██║   ██║╚██████╗██║  ██╗███████╗   ██║               ║"
    echo "║      ╚═╝   ╚═╝ ╚═════╝╚═╝  ╚═╝╚══════╝   ╚═╝               ║"
    echo "║                                                            ║"
    echo "║              TicketDesk Development Server                 ║"
    echo "║                                                            ║"
    echo "╚════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

# 检查依赖
check_dependencies() {
    log_info "检查依赖..."

    # 检查 Go
    if ! command -v go &> /dev/null; then
        log_error "Go 未安装，请先安装 Go"
        exit 1
    fi

    # 检查 Air
    if ! command -v air &> /dev/null; then
        log_warn "Air 未安装，正在安装..."
        go install github.com/air-verse/air@latest
    fi

    # 检查 Node.js
    if ! command -v node &> /dev/null; then
        log_error "Node.js 未安装，请先安装 Node.js"
        exit 1
    fi

    # 检查 npm
    if ! command -v npm &> /dev/null; then
        log_error "npm 未安装，请先安装 npm"
        exit 1
    fi

    log_success "依赖检查通过"
}

# 安装前端依赖
install_frontend_deps() {
    if [ ! -d "web/node_modules" ]; then
        log_info "安装前端依赖..."
        cd web
        npm install
        cd ..
        log_success "前端依赖安装完成"
    fi
}

# 启动后端
start_backend() {
    log_info "启动后端服务 (Air 热更新)..."
    air -c .air.toml &
    BACKEND_PID=$!
    log_success "后端服务已启动 (PID: $BACKEND_PID)"
    log_info "后端地址: ${GREEN}http://localhost:${BACKEND_PORT}${NC}"
}

# 启动前端
start_frontend() {
    log_info "启动前端服务 (Vite 热更新)..."
    cd web
    npm run dev &
    FRONTEND_PID=$!
    cd ..
    log_success "前端服务已启动 (PID: $FRONTEND_PID)"
    log_info "前端地址: ${GREEN}http://localhost:${FRONTEND_PORT}${NC}"
}

# 主函数
main() {
    print_banner

    check_dependencies
    install_frontend_deps

    echo ""
    log_info "=========================================="
    log_info "启动开发服务器..."
    log_info "=========================================="
    echo ""

    start_backend
    sleep 2  # 等待后端启动
    start_frontend

    echo ""
    log_success "=========================================="
    log_success "开发服务器启动完成！"
    log_success "=========================================="
    echo ""
    log_info "后端 API:  ${GREEN}http://localhost:${BACKEND_PORT}${NC}"
    log_info "前端页面:  ${GREEN}http://localhost:${FRONTEND_PORT}${NC}"
    log_info "API 文档:  ${GREEN}http://localhost:${BACKEND_PORT}/swagger/index.html${NC}"
    echo ""
    log_info "按 ${YELLOW}Ctrl+C${NC} 停止所有服务"
    echo ""

    # 等待进程
    wait
}

# 运行主函数
main
