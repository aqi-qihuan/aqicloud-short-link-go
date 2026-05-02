#!/usr/bin/env bash
#
# run_bench.sh — AqiCloud Short-Link 端到端性能压测编排脚本
#
# 完整流程:
#   1. 环境检查 (k6, docker, curl)
#   2. Docker 启动 & 中间件就绪等待
#   3. 应用服务健康检查
#   4. 测试数据填充 (seed_data.py)
#   5. 冒烟测试 (k6_smoke.js)
#   6. 压测执行 (可选单场景或全部)
#   7. 结果汇总 & Markdown 报告生成
#
# 使用方法:
#   ./test/stress/run_bench.sh                        # 全部场景
#   ./test/stress/run_bench.sh redirect               # 仅短链重定向
#   ./test/stress/run_bench.sh mixed                  # 仅混合负载
#   ./test/stress/run_bench.sh gateway                # 仅 Gateway 限流
#   ./test/stress/run_bench.sh smoke                  # 仅冒烟测试
#   ./test/stress/run_bench.sh --skip-docker          # 跳过 Docker 启动
#   ./test/stress/run_bench.sh --count 50000          # 自定义数据量
#   ./test/stress/run_bench.sh --no-seed              # 跳过数据填充
#   ./test/stress/run_bench.sh --no-smoke             # 跳过冒烟测试
#   ./test/stress/run_bench.sh --help                 # 显示帮助
#
# 环境变量:
#   BASE_URL       Gateway 地址 (默认 http://localhost:8888)
#   LINK_DIRECT    Link 服务直连地址 (默认 http://localhost:8003)
#   AUTH_TOKEN     JWT token (可选，不提供则自动登录)
#   LOGIN_PHONE    自动登录手机号
#   LOGIN_PWD      自动登录密码
#   RESULT_DIR     结果输出目录 (默认 test/stress/results)
#   MYSQL_HOST     MySQL 地址 (默认 127.0.0.1)
#   MYSQL_PORT     MySQL 端口 (默认 3306)
#   MYSQL_USER     MySQL 用户 (默认 root)
#   MYSQL_PWD      MySQL 密码 (默认 root)

set -uo pipefail
# 注意: 不用 -e，因为允许单个场景失败后继续执行

# ──────────────────────────────────────
# 颜色 & 日志
# ──────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

info()  { echo -e "${BLUE}[INFO]${NC}  $(date +%H:%M:%S) $*"; }
ok()    { echo -e "${GREEN}[OK]${NC}    $(date +%H:%M:%S) $*"; }
warn()  { echo -e "${YELLOW}[WARN]${NC}  $(date +%H:%M:%S) $*"; }
err()   { echo -e "${RED}[ERROR]${NC} $(date +%H:%M:%S) $*" >&2; }
step()  { echo -e "\n${CYAN}━━━ [$1] $2 ━━━${NC}"; }
fatal() { err "$*"; exit 1; }

# ──────────────────────────────────────
# 默认参数
# ──────────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
RESULT_DIR="${RESULT_DIR:-$SCRIPT_DIR/results}"
BASE_URL="${BASE_URL:-http://localhost:8888}"
LINK_DIRECT="${LINK_DIRECT:-http://localhost:8003}"
MYSQL_HOST="${MYSQL_HOST:-127.0.0.1}"
MYSQL_PORT="${MYSQL_PORT:-3306}"
MYSQL_USER="${MYSQL_USER:-root}"
MYSQL_PWD="${MYSQL_PWD:-root}"
AUTH_TOKEN="${AUTH_TOKEN:-}"
LOGIN_PHONE="${LOGIN_PHONE:-}"
LOGIN_PWD="${LOGIN_PWD:-}"

SKIP_DOCKER=false
SKIP_SEED=false
SKIP_SMOKE=false
DATA_COUNT=10000
SCENARIO="all"
SCRIPT_START_TIME=$(date +%s)

# ──────────────────────────────────────
# 帮助信息
# ──────────────────────────────────────
show_help() {
  cat <<'HELP'
AqiCloud Short-Link 性能压测编排脚本

用法: ./test/stress/run_bench.sh [场景] [选项]

场景 (第一个位置参数):
  all              全部场景顺序执行 (默认)
  smoke            仅冒烟测试 (10秒)
  redirect         仅短链重定向压测
  mixed            仅混合负载压测
  gateway          仅 Gateway 限流压测

选项:
  --skip-docker    跳过 Docker 启动 (假设环境已就绪)
  --no-seed        跳过测试数据填充
  --no-smoke       跳过冒烟测试 (压测前预检)
  --count N        测试数据条数 (默认 10000)
  --base-url URL   Gateway 地址 (默认 http://localhost:8888)
  --token TOKEN    JWT 认证 token
  --phone PHONE    自动登录手机号
  --pwd PASSWORD   自动登录密码
  --help           显示帮助

环境变量:
  BASE_URL, LINK_DIRECT, AUTH_TOKEN, LOGIN_PHONE, LOGIN_PWD
  MYSQL_HOST, MYSQL_PORT, MYSQL_USER, MYSQL_PWD, RESULT_DIR

示例:
  ./test/stress/run_bench.sh                                    # 全流程
  ./test/stress/run_bench.sh redirect --count 50000             # 仅重定向，5万条数据
  ./test/stress/run_bench.sh mixed --skip-docker --no-seed      # 跳过环境和数据
  ./test/stress/run_bench.sh smoke                              # 快速预检
  ./test/stress/run_bench.sh all --phone 13800000001 --pwd 123  # 带自动登录
HELP
}

# ──────────────────────────────────────
# 参数解析 (正确处理位置参数和选项混合)
# ──────────────────────────────────────
while [[ $# -gt 0 ]]; do
  case "$1" in
    --help|-h)      show_help; exit 0 ;;
    --skip-docker)  SKIP_DOCKER=true; shift ;;
    --no-seed)      SKIP_SEED=true; shift ;;
    --no-smoke)     SKIP_SMOKE=true; shift ;;
    --count)        DATA_COUNT="$2"; shift 2 ;;
    --base-url)     BASE_URL="$2"; shift 2 ;;
    --token)        AUTH_TOKEN="$2"; shift 2 ;;
    --phone)        LOGIN_PHONE="$2"; shift 2 ;;
    --pwd)          LOGIN_PWD="$2"; shift 2 ;;
    all|smoke|redirect|mixed|gateway)
      SCENARIO="$1"; shift ;;
    *)
      warn "未知参数: $1 (忽略)"
      shift ;;
  esac
done

# ──────────────────────────────────────
# 计数器 & 结果追踪
# ──────────────────────────────────────
TOTAL_STEPS=0
PASSED_STEPS=0
FAILED_STEPS=0
declare -A SCENARIO_STATUS=()

record_result() {
  local name=$1 status=$2
  SCENARIO_STATUS["$name"]="$status"
  if [[ "$status" == "PASS" ]]; then
    ((PASSED_STEPS++)) || true
    ok "$name → ✅ PASS"
  else
    ((FAILED_STEPS++)) || true
    err "$name → ❌ FAIL"
  fi
  ((TOTAL_STEPS++)) || true
}

# ──────────────────────────────────────
# 启动横幅
# ──────────────────────────────────────
echo ""
echo -e "${BOLD}${CYAN}╔══════════════════════════════════════════════════╗${NC}"
echo -e "${BOLD}${CYAN}║  AqiCloud Short-Link 性能压测                    ║${NC}"
echo -e "${BOLD}${CYAN}║  $(date '+%Y-%m-%d %H:%M:%S')                                    ║${NC}"
echo -e "${BOLD}${CYAN}╚══════════════════════════════════════════════════╝${NC}"
echo ""
info "场景:     $SCENARIO"
info "目标:     $BASE_URL"
info "数据量:   $DATA_COUNT"
info "Docker:   $([ "$SKIP_DOCKER" = true ] && echo '跳过' || echo '自动启动')"
info "数据填充: $([ "$SKIP_SEED" = true ] && echo '跳过' || echo '执行')"
info "冒烟测试: $([ "$SKIP_SMOKE" = true ] && echo '跳过' || echo '执行')"

# ──────────────────────────────────────
# Step 1: 环境检查
# ──────────────────────────────────────
step "1/7" "环境检查"

# k6
if command -v k6 &>/dev/null; then
  K6_VERSION=$(k6 version 2>/dev/null | head -1 || echo "unknown")
  ok "k6: $K6_VERSION"
else
  warn "k6 未安装，尝试自动安装..."
  if [[ "$(uname)" == "Darwin" ]]; then
    if command -v brew &>/dev/null; then
      brew install k6 || fatal "k6 安装失败"
    else
      fatal "请先安装 Homebrew: https://brew.sh"
    fi
  elif [[ "$(uname)" == "Linux" ]]; then
    if command -v apt-get &>/dev/null; then
      sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg \
        --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D68 2>/dev/null || true
      echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" \
        | sudo tee /etc/apt/sources.list.d/k6.list >/dev/null
      sudo apt-get update -qq && sudo apt-get install -y -qq k6 || fatal "k6 安装失败"
    else
      fatal "请手动安装 k6: https://k6.io/docs/getting-started/installation/"
    fi
  else
    fatal "不支持的操作系统，请手动安装 k6"
  fi
  ok "k6 安装完成"
fi

# curl
command -v curl &>/dev/null || fatal "curl 未安装"

# python3
if command -v python3 &>/dev/null; then
  ok "python3: $(python3 --version 2>&1 | awk '{print $2}')"
else
  fatal "python3 未安装 (数据填充需要)"
fi

# docker (如果需要)
if [[ "$SKIP_DOCKER" != true ]]; then
  if command -v docker &>/dev/null; then
    ok "docker: $(docker --version 2>/dev/null | awk '{print $3}' | tr -d ',')"
  else
    fatal "docker 未安装 (使用 --skip-docker 跳过)"
  fi
fi

# ──────────────────────────────────────
# Step 2: Docker 环境
# ──────────────────────────────────────
step "2/7" "Docker 环境"

if [[ "$SKIP_DOCKER" == true ]]; then
  info "跳过 Docker (--skip-docker)"
else
  cd "$PROJECT_ROOT"

  # 检测 compose 命令
  if docker compose version &>/dev/null 2>&1; then
    COMPOSE="docker compose"
  elif command -v docker-compose &>/dev/null; then
    COMPOSE="docker-compose"
  else
    fatal "docker-compose 未安装"
  fi

  info "启动中间件..."
  $COMPOSE up -d mysql redis rabbitmq kafka clickhouse minio 2>/dev/null || \
  $COMPOSE up -d 2>/dev/null

  # 等待 MySQL
  info "等待 MySQL..."
  local_mysql_wait=0
  while [[ $local_mysql_wait -lt 60 ]]; do
    if docker exec aqicloud-mysql mysqladmin ping -h localhost -uroot -proot --silent 2>/dev/null; then
      ok "MySQL 就绪 (${local_mysql_wait}s)"
      break
    fi
    sleep 2
    local_mysql_wait=$((local_mysql_wait + 2))
  done
  [[ $local_mysql_wait -ge 60 ]] && fatal "MySQL 启动超时 (60s)"

  # 等待 Redis
  info "等待 Redis..."
  local_redis_wait=0
  while [[ $local_redis_wait -lt 30 ]]; do
    if docker exec aqicloud-redis redis-cli ping 2>/dev/null | grep -q PONG; then
      ok "Redis 就绪 (${local_redis_wait}s)"
      break
    fi
    sleep 2
    local_redis_wait=$((local_redis_wait + 2))
  done

  # 等待 Kafka
  info "等待 Kafka..."
  sleep 5

  # 启动应用服务
  info "启动应用服务..."
  $COMPOSE up -d gateway account link data shop ai 2>/dev/null || true
  info "等待应用服务启动 (15s)..."
  sleep 15
  ok "Docker 环境就绪"
fi

# ──────────────────────────────────────
# Step 3: 服务健康检查
# ──────────────────────────────────────
step "3/7" "服务健康检查"

check_service() {
  local name=$1 url=$2 required=${3:-true}
  local http_code
  http_code=$(curl -s -o /dev/null -w "%{http_code}" "$url" \
    --connect-timeout 5 --max-time 10 2>/dev/null || echo "000")

  if [[ "$http_code" == "000" ]]; then
    if [[ "$required" == true ]]; then
      err "$name 不可达 ($url)"
      return 1
    else
      warn "$name 不可达 ($url) [非必须]"
      return 0
    fi
  elif [[ "$http_code" =~ ^[2345] ]]; then
    ok "$name → HTTP $http_code"
    return 0
  else
    warn "$name → HTTP $http_code"
    return 0
  fi
}

HEALTH_OK=true
check_service "Gateway"  "$BASE_URL"              true  || HEALTH_OK=false
check_service "Link"     "$LINK_DIRECT"            true  || HEALTH_OK=false
check_service "Account"  "http://localhost:8001"   true  || HEALTH_OK=false
check_service "Data"     "http://localhost:8002"   false || true
check_service "Shop"     "http://localhost:8005"   false || true

if [[ "$HEALTH_OK" != true ]]; then
  fatal "核心服务不可用，请检查 docker-compose logs"
fi

# ──────────────────────────────────────
# Step 4: 测试数据填充
# ──────────────────────────────────────
step "4/7" "测试数据填充"

SAMPLE_CODES=""

if [[ "$SKIP_SEED" == true ]]; then
  info "跳过数据填充 (--no-seed)"
else
  # 检查 pymysql
  if ! python3 -c "import pymysql" 2>/dev/null; then
    info "安装 pymysql..."
    pip3 install pymysql --quiet --break-system-packages 2>/dev/null || pip3 install pymysql --quiet
  fi

  info "填充 $DATA_COUNT 条测试数据..."
  SEED_OUTPUT=$(python3 "$SCRIPT_DIR/seed_data.py" \
    --host "$MYSQL_HOST" --port "$MYSQL_PORT" \
    --user "$MYSQL_USER" --password "$MYSQL_PWD" \
    --count "$DATA_COUNT" 2>&1) || {
    err "数据填充失败:"
    echo "$SEED_OUTPUT"
    warn "继续执行 (k6 将使用动态生成的 codes)"
  }

  if [[ -n "${SEED_OUTPUT:-}" ]]; then
    echo "$SEED_OUTPUT"
    SAMPLE_CODES=$(echo "$SEED_OUTPUT" | grep "样本 codes" | sed "s/.*: //" || echo "")
    if [[ -n "$SAMPLE_CODES" ]]; then
      ok "获取到样本 codes (${#SAMPLE_CODES} chars)"
    else
      warn "未能提取样本 codes"
    fi
  fi
fi

# ──────────────────────────────────────
# Step 5: 冒烟测试
# ──────────────────────────────────────
step "5/7" "冒烟测试"

if [[ "$SCENARIO" == "smoke" ]]; then
  # 仅冒烟测试模式
  info "执行冒烟测试..."
  K6_ENV_ARGS=(-e "BASE_URL=$BASE_URL")
  [[ -n "$SAMPLE_CODES" ]] && K6_ENV_ARGS+=(-e "SHORT_CODES=$SAMPLE_CODES")

  if k6 run "${K6_ENV_ARGS[@]}" "$SCRIPT_DIR/k6_smoke.js" 2>&1; then
    record_result "冒烟测试" "PASS"
  else
    record_result "冒烟测试" "FAIL"
  fi
  # 冒烟模式到此结束
  step "完成" "结果汇总"
  echo ""
  for name in "${!SCENARIO_STATUS[@]}"; do
    status="${SCENARIO_STATUS[$name]}"
    if [[ "$status" == "PASS" ]]; then
      echo -e "  ${GREEN}✅ $name${NC}"
    else
      echo -e "  ${RED}❌ $name${NC}"
    fi
  done
  echo ""
  exit 0

elif [[ "$SKIP_SMOKE" == true ]]; then
  info "跳过冒烟测试 (--no-smoke)"

else
  info "执行冒烟测试..."
  K6_ENV_ARGS=(-e "BASE_URL=$BASE_URL")
  [[ -n "$SAMPLE_CODES" ]] && K6_ENV_ARGS+=(-e "SHORT_CODES=$SAMPLE_CODES")

  if k6 run "${K6_ENV_ARGS[@]}" "$SCRIPT_DIR/k6_smoke.js" 2>&1; then
    ok "冒烟测试通过"
  else
    warn "冒烟测试有问题 (继续执行压测)"
  fi
fi

# ──────────────────────────────────────
# Step 6: 压测执行
# ──────────────────────────────────────
step "6/7" "压测执行"

mkdir -p "$RESULT_DIR"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# 构建通用 k6 环境变量
build_k6_env() {
  local -a args=()
  args+=(-e "BASE_URL=$BASE_URL")
  args+=(-e "LINK_DIRECT=$LINK_DIRECT")
  [[ -n "$SAMPLE_CODES" ]] && args+=(-e "SHORT_CODES=$SAMPLE_CODES")
  [[ -n "$AUTH_TOKEN" ]]    && args+=(-e "AUTH_TOKEN=$AUTH_TOKEN")
  [[ -n "$LOGIN_PHONE" ]]   && args+=(-e "LOGIN_PHONE=$LOGIN_PHONE")
  [[ -n "$LOGIN_PWD" ]]     && args+=(-e "LOGIN_PWD=$LOGIN_PWD")
  echo "${args[@]}"
}

run_scenario() {
  local name=$1 script=$2 desc=$3
  local out_file="$RESULT_DIR/${name}_${TIMESTAMP}.json"
  local summary_file="$RESULT_DIR/${name}_${TIMESTAMP}_summary.txt"
  local log_file="$RESULT_DIR/${name}_${TIMESTAMP}.log"

  echo ""
  info "━━━ $desc ━━━"
  local scenario_start
  scenario_start=$(date +%s)

  # 构建 k6 参数
  local -a K6_ARGS=()
  while IFS= read -r arg; do
    K6_ARGS+=("$arg")
  done < <(build_k6_env | tr ' ' '\n')

  # 执行 k6
  local exit_code=0
  k6 run \
    "${K6_ARGS[@]}" \
    --out json="$out_file" \
    --summary-export="$summary_file" \
    "$SCRIPT_DIR/$script" 2>&1 | tee "$log_file" || exit_code=$?

  local scenario_end
  scenario_end=$(date +%s)
  local scenario_duration=$((scenario_end - scenario_start))

  if [[ $exit_code -eq 0 ]]; then
    record_result "$desc (${scenario_duration}s)" "PASS"
  else
    record_result "$desc (${scenario_duration}s, exit=$exit_code)" "FAIL"
  fi
}

# ── 场景执行逻辑 ──
case "$SCENARIO" in
  all)
    # 全部场景顺序执行，场景间冷却 30s
    run_scenario "redirect" "k6_redirect.js" "场景 1/3: 短链重定向压测"
    echo ""; info "冷却 30s..."; sleep 30

    run_scenario "mixed" "k6_mixed.js" "场景 2/3: 混合负载压测"
    echo ""; info "冷却 30s..."; sleep 30

    run_scenario "gateway" "k6_gateway.js" "场景 3/3: Gateway 限流压测"
    ;;
  redirect)
    run_scenario "redirect" "k6_redirect.js" "短链重定向压测"
    ;;
  mixed)
    run_scenario "mixed" "k6_mixed.js" "混合负载压测"
    ;;
  gateway)
    run_scenario "gateway" "k6_gateway.js" "Gateway 限流压测"
    ;;
  smoke)
    # 已在 Step 5 处理
    ;;
  *)
    fatal "未知场景: $SCENARIO"
    ;;
esac

# ──────────────────────────────────────
# Step 7: 结果汇总 & 报告生成
# ──────────────────────────────────────
step "7/7" "结果汇总"

SCRIPT_END_TIME=$(date +%s)
TOTAL_DURATION=$((SCRIPT_END_TIME - SCRIPT_START_TIME))

# ── 终端汇总 ──
echo ""
echo -e "${BOLD}${CYAN}╔══════════════════════════════════════════════════╗${NC}"
echo -e "${BOLD}${CYAN}║  压测结果汇总                                     ║${NC}"
echo -e "${BOLD}${CYAN}╚══════════════════════════════════════════════════╝${NC}"
echo ""

for name in "${!SCENARIO_STATUS[@]}"; do
  status="${SCENARIO_STATUS[$name]}"
  if [[ "$status" == "PASS" ]]; then
    echo -e "  ${GREEN}✅ $name${NC}"
  else
    echo -e "  ${RED}❌ $name${NC}"
  fi
done

echo ""
echo -e "  通过: ${GREEN}${PASSED_STEPS}${NC}  失败: ${RED}${FAILED_STEPS}${NC}  总计: ${TOTAL_STEPS}  耗时: ${TOTAL_DURATION}s"
echo ""

# 提取各场景 summary 中的关键指标
for summary_file in "$RESULT_DIR"/*_${TIMESTAMP}_summary.txt; do
  [[ -f "$summary_file" ]] || continue
  scene=$(basename "$summary_file" | sed "s/_${TIMESTAMP}_summary.txt//")
  echo -e "${BOLD}[$scene]${NC}"
  grep -E "(http_req_duration|http_req_failed|iterations|vus_max)" "$summary_file" 2>/dev/null | head -10 || true
  echo ""
done

# ── 生成 Markdown 报告 ──
REPORT_FILE="$RESULT_DIR/report_${TIMESTAMP}.md"

cat > "$REPORT_FILE" <<REPORT_EOF
# AqiCloud 性能压测报告

> 生成时间: $(date '+%Y-%m-%d %H:%M:%S')
> 场景: $SCENARIO
> 目标: $BASE_URL
> 数据量: $DATA_COUNT
> 总耗时: ${TOTAL_DURATION}s

## 执行结果

| 场景 | 状态 |
|------|------|
REPORT_EOF

for name in "${!SCENARIO_STATUS[@]}"; do
  status="${SCENARIO_STATUS[$name]}"
  if [[ "$status" == "PASS" ]]; then
    echo "| $name | ✅ PASS |" >> "$REPORT_FILE"
  else
    echo "| $name | ❌ FAIL |" >> "$REPORT_FILE"
  fi
done

cat >> "$REPORT_FILE" <<'REPORT_EOF'

## 关键指标

REPORT_EOF

for summary_file in "$RESULT_DIR"/*_${TIMESTAMP}_summary.txt; do
  [[ -f "$summary_file" ]] || continue
  scene=$(basename "$summary_file" | sed "s/_${TIMESTAMP}_summary.txt//")
  echo "### $scene" >> "$REPORT_FILE"
  echo "" >> "$REPORT_FILE"
  echo '```' >> "$REPORT_FILE"
  # 提取关键指标
  grep -E "(http_req_duration|http_req_failed|iterations|data_received|data_sent|vus_max|checks)" \
    "$summary_file" 2>/dev/null | head -20 >> "$REPORT_FILE" || echo "(无 summary 数据)" >> "$REPORT_FILE"
  echo '```' >> "$REPORT_FILE"
  echo "" >> "$REPORT_FILE"
done

cat >> "$REPORT_FILE" <<REPORT_EOF

## 输出文件

| 文件 | 说明 |
|------|------|
REPORT_EOF

for f in "$RESULT_DIR"/*_${TIMESTAMP}*; do
  [[ -f "$f" ]] || continue
  fname=$(basename "$f")
  size=$(ls -lh "$f" | awk '{print $5}')
  echo "| \`$fname\` | $size |" >> "$REPORT_FILE"
done

cat >> "$REPORT_FILE" <<'REPORT_EOF'

## 优化建议

基于压测结果，参考 `performance-report.md` 中的优化方案:
1. **P0**: 热路径添加 Redis 缓存 (预计 QPS 提升 50x)
2. **P0**: Gateway 复用 Proxy 实例 (消除 GC 风暴)
3. **P1**: 配置 GORM 连接池 (防止连接耗尽)
4. **P1**: 修复 Kafka BatchTimeout (10ns → 100ms)
REPORT_EOF

ok "报告已生成: $REPORT_FILE"

# ── 最终输出 ──
echo ""
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "结果目录: ${RESULT_DIR}"
echo -e "报告文件: ${REPORT_FILE}"
echo -e "时间戳:   ${TIMESTAMP}"
echo ""
echo -e "输出文件:"
ls -lh "$RESULT_DIR"/*_${TIMESTAMP}* 2>/dev/null
echo ""

if [[ $FAILED_STEPS -eq 0 ]]; then
  echo -e "${GREEN}${BOLD}全部通过！${NC}"
  exit 0
else
  echo -e "${YELLOW}${BOLD}有 $FAILED_STEPS 个场景失败${NC}"
  exit 1
fi
