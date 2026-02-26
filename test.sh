#!/bin/bash
#
# alert-webhook 测试脚本
# 适用于 CentOS 7
#

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="${SCRIPT_DIR}"
BINARY_NAME="alert-webhook"
BINARY_PATH="${PROJECT_DIR}/${BINARY_NAME}"
CONFIG_FILE="${PROJECT_DIR}/.env"

# 服务配置（从 .env 读取或使用默认值）
SERVER_PORT="${SERVER_PORT:-8080}"
SERVER_URL="http://localhost:${SERVER_PORT}"
MYSQL_HOST="${MYSQL_HOST:-localhost}"
MYSQL_PORT="${MYSQL_PORT:-3306}"
MYSQL_USER="${MYSQL_USER:-root}"
MYSQL_PASSWORD="${MYSQL_PASSWORD:-}"
MYSQL_DATABASE="${MYSQL_DATABASE:-alerts}"

# 测试统计
TESTS_PASSED=0
TESTS_FAILED=0

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((TESTS_PASSED++))
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((TESTS_FAILED++))
}

# 检查命令是否存在
check_command() {
    if ! command -v "$1" &> /dev/null; then
        log_error "命令 '$1' 未找到，请先安装"
        return 1
    fi
    return 0
}

# 加载配置文件
load_config() {
    if [ -f "${CONFIG_FILE}" ]; then
        log_info "从 .env 文件加载配置..."
        export $(grep -v '^#' "${CONFIG_FILE}" | xargs)
        SERVER_PORT="${SERVER_PORT:-8080}"
        SERVER_URL="http://localhost:${SERVER_PORT}"
    fi
}

# 测试1: 检查 Go 环境
test_go_environment() {
    log_info "测试 1/7: 检查 Go 编译环境..."
    
    if check_command "go"; then
        GO_VERSION=$(go version)
        log_success "Go 环境正常: ${GO_VERSION}"
        return 0
    else
        log_fail "Go 环境检查失败"
        return 1
    fi
}

# 测试2: 编译项目
test_build() {
    log_info "测试 2/7: 编译项目..."
    
    cd "${PROJECT_DIR}"
    
    # 清理旧构建
    if [ -f "${BINARY_PATH}" ]; then
        rm -f "${BINARY_PATH}"
    fi
    
    # 编译 Linux 版本
    if GOOS=linux GOARCH=amd64 go build -o "${BINARY_PATH}-linux" .; then
        log_success "Linux 版本编译成功: alert-webhook-linux"
    else
        log_fail "编译失败"
        return 1
    fi
    
    # 编译当前系统版本（用于测试）
    if go build -o "${BINARY_PATH}" .; then
        log_success "测试二进制编译成功: ${BINARY_NAME}"
        return 0
    else
        log_fail "测试二进制编译失败"
        return 1
    fi
}

# 测试3: 检查配置文件
test_config() {
    log_info "测试 3/7: 检查配置文件..."
    
    if [ ! -f "${CONFIG_FILE}" ]; then
        log_fail "配置文件 .env 不存在"
        return 1
    fi
    
    # 检查必要的配置项
    if [ -z "${MYSQL_HOST}" ]; then
        log_fail "MYSQL_HOST 未配置"
        return 1
    fi
    
    if [ -z "${MYSQL_USER}" ]; then
        log_fail "MYSQL_USER 未配置"
        return 1
    fi
    
    if [ -z "${MYSQL_DATABASE}" ]; then
        log_fail "MYSQL_DATABASE 未配置"
        return 1
    fi
    
    log_success "配置文件检查通过"
    log_info "  - 数据库: ${MYSQL_HOST}:${MYSQL_PORT}/${MYSQL_DATABASE}"
    log_info "  - 用户: ${MYSQL_USER}"
    log_info "  - 端口: ${SERVER_PORT}"
    return 0
}

# 测试4: 检查 MySQL 连接
test_mysql_connection() {
    log_info "测试 4/7: 检查 MySQL 连接..."
    
    check_command "mysql" || {
        log_warn "mysql 命令未安装，尝试使用其他方式检测"
        # 尝试通过服务启动后健康检查来验证
        return 0
    }
    
    # 尝试连接 MySQL
    if mysql -h"${MYSQL_HOST}" -P"${MYSQL_PORT}" -u"${MYSQL_USER}" -p"${MYSQL_PASSWORD}" -e "SELECT 1" &> /dev/null; then
        log_success "MySQL 连接成功"
        return 0
    else
        log_fail "MySQL 连接失败，请检查配置"
        return 1
    fi
}

# 启动服务（后台）
start_service() {
    log_info "启动 alert-webhook 服务..."
    
    cd "${PROJECT_DIR}"
    
    # 检查是否已运行
    if pgrep -f "./${BINARY_NAME}" > /dev/null; then
        log_warn "服务已在运行，先停止..."
        stop_service
    fi
    
    # 启动服务
    nohup ./${BINARY_NAME} > /tmp/alert-webhook.log 2>&1 &
    SERVICE_PID=$!
    
    # 等待服务启动
    log_info "等待服务启动 (PID: ${SERVICE_PID})..."
    sleep 3
    
    # 检查进程是否存活
    if ! kill -0 ${SERVICE_PID} 2>/dev/null; then
        log_error "服务启动失败"
        cat /tmp/alert-webhook.log
        return 1
    fi
    
    # 等待服务就绪
    for i in {1..10}; do
        if curl -s "${SERVER_URL}/health" &> /dev/null; then
            log_success "服务启动成功"
            return 0
        fi
        sleep 1
    done
    
    log_error "服务启动超时"
    cat /tmp/alert-webhook.log
    return 1
}

# 停止服务
stop_service() {
    log_info "停止 alert-webhook 服务..."
    
    # 尝试多种方式停止
    pkill -f "./${BINARY_NAME}" 2>/dev/null || true
    pkill -f "alert-webhook" 2>/dev/null || true
    
    # 等待进程退出
    sleep 1
    
    log_info "服务已停止"
}

# 测试5: 健康检查端点
test_health_endpoint() {
    log_info "测试 5/7: 测试健康检查端点 /health..."
    
    RESPONSE=$(curl -s -w "\n%{http_code}" "${SERVER_URL}/health")
    HTTP_CODE=$(echo "${RESPONSE}" | tail -n1)
    BODY=$(echo "${RESPONSE}" | head -n-1)
    
    if [ "${HTTP_CODE}" = "200" ]; then
        log_success "健康检查端点正常 (HTTP ${HTTP_CODE})"
        return 0
    else
        log_fail "健康检查失败 (HTTP ${HTTP_CODE}): ${BODY}"
        return 1
    fi
}

# 测试6: Webhook 端点
test_webhook_endpoint() {
    log_info "测试 6/7: 测试 Webhook 端点 /webhook..."
    
    # 构造测试告警数据（AlertManager 格式）
    TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    FINGERPRINT="test-fingerprint-$(date +%s)"
    
    WEBHOOK_JSON=$(cat <<EOF
{
  "version": "4",
  "groupKey": "test-group-key",
  "status": "firing",
  "receiver": "test-receiver",
  "groupLabels": {"alertname": "TestAlert"},
  "commonLabels": {"severity": "critical", "team": "ops"},
  "commonAnnotations": {"description": "Test alert"},
  "externalURL": "http://alertmanager:9093",
  "truncatedAlerts": 0,
  "alerts": [
    {
      "status": "firing",
      "labels": {
        "alertname": "TestAlert",
        "severity": "critical",
        "team": "ops",
        "instance": "server01"
      },
      "annotations": {
        "description": "Test alert for validation",
        "summary": "This is a test alert"
      },
      "startsAt": "${TIMESTAMP}",
      "endsAt": null,
      "generatorURL": "http://prometheus:9090/graph?g0.expr=up&g0.tab=1",
      "fingerprint": "${FINGERPRINT}"
    }
  ]
}
EOF
)
    
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST \
        -H "Content-Type: application/json" \
        -d "${WEBHOOK_JSON}" \
        "${SERVER_URL}/webhook")
    
    HTTP_CODE=$(echo "${RESPONSE}" | tail -n1)
    BODY=$(echo "${RESPONSE}" | head -n-1)
    
    if [ "${HTTP_CODE}" = "200" ]; then
        # 检查返回内容
        if echo "${BODY}" | grep -q "success"; then
            log_success "Webhook 端点正常 (HTTP ${HTTP_CODE})"
            log_info "响应: ${BODY}"
            return 0
        else
            log_fail "Webhook 返回格式异常: ${BODY}"
            return 1
        fi
    else
        log_fail "Webhook 请求失败 (HTTP ${HTTP_CODE}): ${BODY}"
        return 1
    fi
}

# 测试7: 验证数据存储
test_data_storage() {
    log_info "测试 7/7: 验证告警数据存储..."
    
    check_command "mysql" || {
        log_warn "mysql 命令不可用，跳过数据库验证"
        log_info "请手动验证: SELECT * FROM alerts;"
        return 0
    }
    
    # 查询数据库
    RESULT=$(mysql -h"${MYSQL_HOST}" -P"${MYSQL_PORT}" -u"${MYSQL_USER}" -p"${MYSQL_PASSWORD}" \
        -e "SELECT COUNT(*) as cnt FROM ${MYSQL_DATABASE}.alerts WHERE fingerprint LIKE 'test-fingerprint-%'" \
        -s -N 2>/dev/null)
    
    if [ "${RESULT}" -gt 0 ]; then
        log_success "告警数据已成功存储到数据库 (${RESULT} 条记录)"
        
        # 显示数据详情
        log_info "数据库中的测试告警记录:"
        mysql -h"${MYSQL_HOST}" -P"${MYSQL_PORT}" -u"${MYSQL_USER}" -p"${MYSQL_PASSWORD}" \
            -e "SELECT id, alertname, status, severity, fingerprint FROM ${MYSQL_DATABASE}.alerts WHERE fingerprint LIKE 'test-fingerprint-%'" \
            2>/dev/null || true
        return 0
    else
        log_fail "数据库中未找到测试告警记录"
        return 1
    fi
}

# 清理测试数据
cleanup_test_data() {
    log_info "清理测试数据..."
    
    check_command "mysql" || return 0
    
    mysql -h"${MYSQL_HOST}" -P"${MYSQL_PORT}" -u"${MYSQL_USER}" -p"${MYSQL_PASSWORD}" \
        -e "DELETE FROM ${MYSQL_DATABASE}.alerts WHERE fingerprint LIKE 'test-fingerprint-%'" \
        2>/dev/null && log_info "测试数据已清理" || log_warn "清理失败，可忽略"
}

# 打印测试摘要
print_summary() {
    echo ""
    echo "========================================"
    echo "           测试结果摘要"
    echo "========================================"
    echo -e "通过: ${GREEN}${TESTS_PASSED}${NC}"
    echo -e "失败: ${RED}${TESTS_FAILED}${NC}"
    echo "========================================"
    
    if [ ${TESTS_FAILED} -eq 0 ]; then
        echo -e "${GREEN}所有测试通过!${NC}"
        return 0
    else
        echo -e "${RED}部分测试失败，请检查上述错误${NC}"
        return 1
    fi
}

# 主函数
main() {
    echo "========================================"
    echo "     alert-webhook 测试脚本"
    echo "     适用于 CentOS 7"
    echo "========================================"
    echo ""
    
    # 加载配置
    load_config
    
    # 执行测试
    test_go_environment
    test_config
    test_build
    test_mysql_connection
    
    # 启动服务
    start_service
    
    # 测试服务
    test_health_endpoint
    test_webhook_endpoint
    test_data_storage
    
    # 清理
    cleanup_test_data
    
    # 停止服务
    stop_service
    
    # 打印摘要
    print_summary
}

# 显示帮助
show_help() {
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  --help     显示帮助信息"
    echo "  --build    仅编译项目"
    echo "  --run      仅运行服务测试"
    echo "  --quick    快速测试（跳过编译）"
    echo ""
}

# 根据参数执行
case "${1:-}" in
    --help)
        show_help
        exit 0
        ;;
    --build)
        load_config
        test_go_environment
        test_config
        test_build
        exit 0
        ;;
    --run)
        load_config
        start_service
        test_health_endpoint
        test_webhook_endpoint
        test_data_storage
        cleanup_test_data
        stop_service
        print_summary
        exit 0
        ;;
    --quick)
        load_config
        start_service
        test_health_endpoint
        test_webhook_endpoint
        stop_service
        print_summary
        exit 0
        ;;
    *)
        main
        ;;
esac
