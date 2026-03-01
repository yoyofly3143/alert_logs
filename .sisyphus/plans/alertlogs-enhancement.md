# alertlogs 功能增强计划

## TL;DR

> **快速摘要**：为 alert-webhook 项目增加 JWT 认证、按 labels.quality 级别筛选、默认显示最近1天告警、统计支持筛选刷新、前端改名为 alertlogs 并实现无感刷新。
> 
> **交付物**：
> - JWT 认证中间件 + 登录 API
> - 后端 quality 筛选功能
> - 统计 API 支持 quality + 时间范围筛选
> - 前端登录页面 + 无感刷新 + 默认时间修改
> 
> **预估工作量**：中等 (Medium)
> **并行执行**：是 - 2个波次
> **关键路径**：认证中间件 → API 筛选 → 前端集成

---

## 上下文

### 原始需求
1. 增加登录功能 - JWT 认证
2. 提取 labels.quality 作为前端级别筛选
3. 前端默认只显示最近一天内的告警
4. 保留时间范围筛选功能
5. 统计显示：总告警数、告警中、已恢复、最近24小时 quality 匹配的告警数
6. 统计可根据选择的时间范围刷新
7. 前端改名为 alertlogs
8. 前端无感刷新

### 技术栈
- 后端：Go + Gin + MySQL
- 前端：Vanilla JS + HTML/CSS
- 认证：JWT

### Metis 审查发现的关键点

**必须保持公开的端点**（防止认证破坏告警接收）：
- `/webhook` - AlertManager 推送接口
- `/health` - 健康检查
- `/` - 登录页面

**必须设置的 Guardrails**：
- 不添加用户管理 UI（单管理员通过环境变量配置）
- 不修改告警表结构
- 不增加每页显示上限（保持100）

---

## 工作目标

### 核心目标
为 alert-webhook 增加完整的认证和筛选功能，实现安全可控的告警查看体验。

### 具体交付物

| 交付物 | 说明 |
|--------|------|
| JWT 认证系统 | 登录 API + 中间件 + Token 验证 |
| Quality 筛选 | 后端按 labels.quality 筛选，前端下拉选择 |
| 统计筛选 | 统计 API 支持 quality + 时间范围参数 |
| 前端改造 | 登录页、默认1天、无感刷新、名称修改 |

### 定义完成
- [ ] 登录功能正常工作（正确/错误密码返回对应状态码）
- [ ] 前端未登录时跳转到登录页
- [ ] Quality 筛选生效（选择 critical 只显示 critical 告警）
- [ ] 默认显示最近1天（不是7天）
- [ ] 统计响应 quality 和时间范围筛选
- [ ] 前端标题显示为 alertlogs
- [ ] 自动刷新不阻塞用户操作

### 必须有
- JWT_SECRET 环境变量配置
- /webhook 和 /health 保持公开
- Token 过期处理

### 必须没有
- 用户管理 UI（通过环境变量配置管理员）
- 密码重置功能
- 告警表结构变更
- 导出功能

---

## 验证策略

### 测试决策
- **测试框架**：无现有测试框架
- **自动化测试**：无（用户未要求）
- **验证方式**：Agent-Executed QA - 通过 curl 直接测试 API

### QA 策略
每个任务完成后，执行 agent-executed QA 验证：
- **API 测试**：使用 curl 发送请求，验证响应状态码和内容
- **前端测试**：使用 Playwright 验证页面行为

---

## 执行策略

### 波次1：后端核心（并行度高）

| 任务 | 描述 | 并行 |
|------|------|------|
| T1 | JWT 配置 + 中间件 | - |
| T2 | 登录 API | T1 完成后 |
| T3 | Quality 筛选（GetAlerts） | T1 完成后 |
| T4 | 修复 GetSeverities（从 DB 查询） | T1 完成后 |
| T5 | 统计 API 筛选参数 | T3,T4 完成后 |

### 波次2：前端集成

| 任务 | 描述 | 并行 |
|------|------|------|
| T6 | 登录页面 + Token 处理 | T2 完成后 |
| T7 | Quality 筛选 + 默认时间修改 | T5 完成后 |
| T8 | 无感刷新 + 标题修改 | T6,T7 完成后 |

### 波次3：最终验证

| 任务 | 描述 |
|------|------|
| T9 | 端到端功能验证 |

---

## TODOs

- [x] 1. 后端：JWT 配置与认证中间件

  **What to do**:
  - 在 `config/config.go` 添加 JWT_SECRET 配置项
  - 创建 `middleware/auth.go` - JWT 验证中间件
  - 修改 `main.go` - 路由应用中间件（保护 /api/*，排除 /api/health）
  
  **Must NOT do**:
  - 不要在 /webhook 应用中间件
  - 不要在 /health 应用中间件
  - 不要在 / (首页) 应用中间件

  **Recommended Agent Profile**:
  > **Category**: `deep` - 需要理解 JWT 流程和 Gin 中间件
  >   Reason: 涉及认证流程和安全逻辑
  > **Skills**: `[]`
  >   - 不需要额外技能

  **Parallelization**:
  - **Can Run In Parallel**: NO (基础组件)
  - **Parallel Group**: Wave 1
  - **Blocks**: T2, T5
  - **Blocked By**: None

  **References**:
  - `config/config.go` - 配置结构参考
  - `handlers/webhook.go` - 现有中间件模式参考
  - 官方库：`github.com/golang-jwt/jwt/v5` - JWT 实现

  **Acceptance Criteria**:
  - [ ] JWT_SECRET 未配置时服务启动失败
  - [ ] 中间件正确拦截未授权请求

  **QA Scenarios**:
  ```
  Scenario: 服务启动时没有 JWT_SECRET
    Tool: Bash
    Steps:
      1. 临时移除 .env 中 JWT_SECRET
      2. 启动服务
    Expected Result: 服务启动失败，日志提示 JWT_SECRET 未配置
  
  Scenario: 保护 API 端点
    Tool: Bash (curl)
    Steps:
      1. curl http://localhost:8080/api/alerts/stats
    Expected Result: 返回 401 Unauthorized
  ```

  **Commit**: YES
  - Message: `feat(auth): add JWT middleware and config`
  - Files: `config/config.go`, `middleware/auth.go`
  - Pre-commit: `go build ./...`

---

- [x] 2. 后端：登录 API

  **What to do**:
  - 创建 `handlers/auth.go` - 登录/登出处理
  - 添加 `/api/auth/login` 端点
  - 用户名密码从环境变量配置（ADMIN_USER, ADMIN_PASSWORD）
  
  **Must NOT do**:
  - 不实现用户注册功能
  - 不实现密码修改功能

  **Recommended Agent Profile**:
  > **Category**: `deep` - 登录逻辑涉及安全处理
  >   Reason: 需要正确处理密码验证和 Token 生成
  > **Skills**: `[]`

  **Parallelization**:
  - **Can Run In Parallel**: NO (依赖 T1)
  - **Parallel Group**: Wave 1
  - **Blocks**: T6
  - **Blocked By**: T1

  **References**:
  - `handlers/webhook.go` - 现有 handler 模式
  - `config/config.go` - 配置读取

  **Acceptance Criteria**:
  - [ ] 正确用户名密码返回 200 + JWT token
  - [ ] 错误用户名密码返回 401

  **QA Scenarios**:
  ```
  Scenario: 登录成功
    Tool: Bash
    Steps:
      1. curl -X POST http://localhost:8080/api/auth/login -H "Content-Type: application/json" -d '{"username":"admin","password":"admin123"}'
    Expected Result: 返回 {"success":true,"token":"eyJ..."}

  Scenario: 登录失败
    Tool: Bash
    Steps:
      1. curl -X POST http://localhost:8080/api/auth/login -H "Content-Type: application/json" -d '{"username":"admin","password":"wrong"}'
    Expected Result: 返回 401 Unauthorized
  ```

  **Commit**: YES
  - Message: `feat(auth): add login endpoint`
  - Files: `handlers/auth.go`

---

- [ ] 3. 后端：Quality 筛选功能

  **What to do**:
  - 修改 `handlers/alert.go` 的 GetAlerts 函数
  - 添加 `quality` 查询参数支持
  - 从 `labels->>'$.quality'` 或 JSON 字段查询
  
  **Must NOT do**:
  - 不修改数据库表结构

  **Recommended Agent Profile**:
  > **Category**: `deep` - 需要理解 JSON 查询
  >   Reason: labels 是 JSON 字段，需要正确查询
  > **Skills**: `[]`

  **Parallelization**:
  - **Can Run In Parallel**: NO (依赖 T1)
  - **Parallel Group**: Wave 1
  - **Blocks**: T7
  - **Blocked By**: T1

  **References**:
  - `handlers/alert.go:GetAlerts` - 现有筛选逻辑参考
  - `models/alert.go` - Alert 模型定义

  **Acceptance Criteria**:
  - [ ] /api/alerts?quality=critical 只返回 critical 告警
  - [ ] /api/alerts?quality= 空时返回所有

  **QA Scenarios**:
  ```
  Scenario: Quality 筛选正常
    Tool: Bash
    Steps:
      1. 先创建一些测试数据
      2. curl "http://localhost:8080/api/alerts?quality=critical"
    Expected Result: 返回的告警 labels.quality 均为 critical

  Scenario: 无 quality 参数
    Tool: Bash
    Steps:
      1. curl "http://localhost:8080/api/alerts"
    Expected Result: 返回所有告警
  ```

  **Commit**: YES
  - Message: `feat(filter): add quality filter to alerts`
  - Files: `handlers/alert.go`

---

- [ ] 4. 后端：修复 GetSeverities 返回真实数据

  **What to do**:
  - 修改 `handlers/alert.go` 的 GetSeverities 函数
  - 从数据库查询 DISTINCT quality 值
  - 不再返回硬编码数组
  
  **Must NOT do**:
  - 不返回硬编码值

  **Recommended Agent Profile**:
  > **Category**: `unspecified-low` - 小改动
  >   Reason: 简单的数据库查询修改
  > **Skills**: `[]`

  **Parallelization**:
  - **Can Run In Parallel**: NO (依赖 T1)
  - **Parallel Group**: Wave 1
  - **Blocks**: T7
  - **Blocked By**: T1

  **References**:
  - `handlers/alert.go:GetSeverities` - 当前实现

  **Acceptance Criteria**:
  - [ ] /api/filters/severities 返回数据库中实际存在的 quality 值

  **QA Scenarios**:
  ```
  Scenario: 获取真实 severity 列表
    Tool: Bash
    Steps:
      1. curl http://localhost:8080/api/filters/severities
    Expected Result: 返回数组包含数据库中实际存在的 quality 值
  ```

  **Commit**: YES
  - Message: `fix(alert): return actual severities from DB`
  - Files: `handlers/alert.go`

---

- [ ] 5. 后端：统计 API 筛选参数

  **What to do**:
  - 修改 GetStats 函数
  - 添加 quality、start_date、end_date 参数支持
  - 统计时应用筛选条件
  
  **Must NOT do**:
  - 不破坏现有统计逻辑

  **Recommended Agent Profile**:
  > **Category**: `deep` - 统计逻辑修改
  >   Reason: 需要理解多个筛选条件的组合查询
  > **Skills**: `[]`

  **Parallelization**:
  - **Can Run In Parallel**: NO (依赖 T3,T4)
  - **Parallel Group**: Wave 1
  - **Blocks**: T7
  - **Blocked By**: T3, T4

  **References**:
  - `handlers/alert.go:GetStats` - 当前实现
  - `handlers/alert.go:GetAlerts` - 筛选逻辑参考

  **Acceptance Criteria**:
  - [ ] /api/alerts/stats?quality=critical 只统计 critical 告警
  - [ ] /api/alerts/stats?start_date=2026-02-01&end_date=2026-02-15 统计指定范围

  **QA Scenarios**:
  ```
  Scenario: 统计按 quality 筛选
    Tool: Bash
    Steps:
      1. curl "http://localhost:8080/api/alerts/stats?quality=critical"
    Expected Result: 返回的 total 等于 critical 告警数

  Scenario: 统计按时间范围筛选
    Tool: Bash
    Steps:
      1. curl "http://localhost:8080/api/alerts/stats?start_date=2026-02-27&end_date=2026-02-28"
    Expected Result: 返回指定日期范围内的统计
  ```

  **Commit**: YES
  - Message: `feat(stats): add quality and date filters`
  - Files: `handlers/alert.go`

---

- [ ] 6. 前端：登录页面 + Token 处理

  **What to do**:
  - 修改 `templates/index.html` - 添加登录表单
  - 修改 `static/js/app.js` - 添加登录/登出逻辑
  - Token 存储在 localStorage
  - 401 响应时跳转到登录页
  
  **Must NOT do**:
  - 不添加用户注册页面

  **Recommended Agent Profile**:
  > **Category**: `visual-engineering` - 前端 UI 修改
  >   Reason: 需要修改 HTML 和 JS
  > **Skills**: `[]`

  **Parallelization**:
  - **Can Run In Parallel**: NO (依赖 T2)
  - **Parallel Group**: Wave 2
  - **Blocks**: T8
  - **Blocked By**: T2

  **References**:
  - `templates/index.html` - 现有结构
  - `static/js/app.js` - 现有 JS 模式
  - `static/css/style.css` - 样式参考

  **Acceptance Criteria**:
  - [ ] 访问首页显示登录表单（未登录时）
  - [ ] 登录成功后跳转到主界面
  - [ ] Token 保存在 localStorage
  - [ ] API 返回 401 时自动跳转到登录页

  **QA Scenarios**:
  ```
  Scenario: 未登录访问首页
    Tool: Playwright
    Steps:
      1. 清除 localStorage
      2. 访问 http://localhost:8080/
    Expected Result: 显示登录表单

  Scenario: 登录成功
    Tool: Playwright
    Steps:
      1. 填写用户名 admin，密码 admin123
      2. 点击登录
    Expected Result: 跳转到主界面，显示告警列表
  ```

  **Commit**: YES
  - Message: `feat(ui): add login page and auth handling`
  - Files: `templates/index.html`, `static/js/app.js`

---

- [ ] 7. 前端：Quality 筛选 + 默认时间修改

  **What to do**:
  - 修改筛选栏的 severity 下拉，从 API 获取真实值
  - 修改 `initDateDefaults()` - 默认改为最近1天
  - 筛选变更时刷新统计
  
  **Must NOT do**:
  - 不改变筛选栏的整体布局

  **Recommended Agent Profile**:
  > **Category**: `visual-engineering` - 前端筛选逻辑
  >   Reason: 需要修改 JS 筛选逻辑
  > **Skills**: `[]`

  **Parallelization**:
  - **Can Run In Parallel**: NO (依赖 T5,T6)
  - **Parallel Group**: Wave 2
  - **Blocks**: T8
  - **Blocked By**: T5, T6

  **References**:
  - `static/js/app.js` - 现有筛选逻辑

  **Acceptance Criteria**:
  - [ ] Severity 下拉显示数据库中真实的 quality 值
  - [ ] 默认选择"全部"
  - [ ] 默认时间范围是今天（不是7天前）

  **QA Scenarios**:
  ```
  Scenario: Severity 下拉选项
    Tool: Playwright
    Steps:
      1. 登录后查看筛选栏
    Expected Result: Severity 下拉包含数据库中实际存在的值

  Scenario: 默认时间
    Tool: Playwright
    Steps:
      1. 登录后查看日期输入框
    Expected Result: 开始日期是今天（不是7天前）
  ```

  **Commit**: YES
  - Message: `feat(ui): add quality filter and fix default date`
  - Files: `static/js/app.js`

---

- [ ] 8. 前端：无感刷新 + 标题修改

  **What to do**:
  - 修改自动刷新逻辑 - 使用 Promise 而不阻塞 UI
  - 修改页面标题为 "alertlogs"
  - 确保刷新时表格数据无缝更新
  
  **Must NOT do**:
  - 不改变分页逻辑

  **Recommended Agent Profile**:
  > **Category**: `visual-engineering` - 前端刷新优化
  >   Reason: 需要修改 JS 刷新逻辑
  > **Skills**: `[]`

  **Parallelization**:
  - **Can Run In Parallel**: NO (依赖 T6,T7)
  - **Parallel Group**: Wave 2
  - **Blocks**: T9
  - **Blocked By**: T6, T7

  **References**:
  - `static/js/app.js` - 当前刷新逻辑
  - `templates/index.html` - 标题位置

  **Acceptance Criteria**:
  - [ ] 页面标题显示 "alertlogs"
  - [ ] 自动刷新不显示加载遮罩
  - [ ] 用户操作时刷新不中断

  **QA Scenarios**:
  ```
  Scenario: 页面标题
    Tool: Playwright
    Steps:
      1. 访问页面
    Expected Result: 标题显示 "alertlogs"

  Scenario: 无感刷新
    Tool: Playwright
    Steps:
      1. 登录后让页面自动刷新
      2. 在刷新期间尝试点击分页
    Expected Result: 刷新不阻塞用户操作
  ```

  **Commit**: YES
  - Message: `feat(ui): seamless refresh and rename to alertlogs`
  - Files: `templates/index.html`, `static/js/app.js`

---

- [ ] 9. 端到端功能验证

  **What to do**:
  - 测试完整流程：登录 → 查看统计 → 筛选 → 登出
  - 验证所有验收标准
  
  **Must NOT do**:
  - 不修改任何代码

  **Recommended Agent Profile**:
  > **Category**: `unspecified-high` - 综合验证
  >   Reason: 需要全面测试
  > **Skills**: `["playwright"]`

  **Parallelization**:
  - **Can Run In Parallel**: NO (最终验证)
  - **Parallel Group**: Wave 3
  - **Blocks**: None
  - **Blocked By**: T8

  **References**:
  - 所有修改的文件

  **Acceptance Criteria**:
  - [ ] 登录功能正常
  - [ ] 质量筛选生效
  - [ ] 统计筛选生效
  - [ ] 前端默认1天
  - [ ] 标题正确
  - [ ] Webhook 端点保持公开

  **QA Scenarios**:
  ```
  Scenario: 完整流程测试
    Tool: Playwright + Bash
    Steps:
      1. 访问首页 → 显示登录
      2. 登录 → 进入主界面
      3. 查看统计 → 显示正确
      4. 筛选 quality → 结果正确
      5. 登出 → 跳转登录页
      6. 测试 webhook 公开 → 返回 200
  ```

  **Commit**: NO
  - Message: ``

---

## 最终验证波次

- [ ] F1. **Plan Compliance Audit** — `oracle`
  验证所有 Must Have 已实现，Must NOT Have 未实现

- [ ] F2. **Code Quality Review** — `unspecified-high`
  运行 `go build`，检查代码规范

- [ ] F3. **Real Manual QA** — `unspecified-high`
  执行端到端测试

- [ ] F4. **Scope Fidelity Check** — `deep`
  确认所有需求已覆盖

---

## 提交策略

每个任务独立提交：
- `feat(auth): add JWT middleware`
- `feat(auth): add login endpoint`
- `feat(filter): add quality filter`
- `fix(alert): return severities from DB`
- `feat(stats): add filter params`
- `feat(ui): add login page`
- `feat(ui): quality filter + date default`
- `feat(ui): seamless refresh + rename`

---

## 成功标准

### 验证命令
```bash
# 登录
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# 访问受保护端点
TOKEN="your_token"
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/alerts/stats

# Webhook 保持公开
curl -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -d '{"alerts":[...]}'

# Quality 筛选
curl "http://localhost:8080/api/alerts?quality=critical"
```

### 最终检查清单
- [ ] JWT 认证工作正常
- [ ] /webhook 保持公开
- [ ] Quality 筛选生效
- [ ] 统计支持筛选
- [ ] 前端默认显示1天
- [ ] 标题为 alertlogs
- [ ] 无感刷新
