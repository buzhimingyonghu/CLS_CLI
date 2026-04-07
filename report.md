# CLS CLI 自然语言交互能力建设 —— 阶段成果汇报

## 背景

基于腾讯云日志服务（CLS）的 cls-cli 工具，本阶段在原有基础上新增了仪表盘相关 API 支持，并结合 CodeBuddy Code 的 Skill 能力，实现了通过**自然语言**操作 CLS 仪表盘的完整交互流程。

---

## 一、新增 CLI 命令

### 1. 仪表盘管理模块 `cls-cli dashboard`

| 命令 | 说明 |
|------|------|
| `+list` | 列出仪表盘，支持按名称（精确匹配）、ID、文件夹、标签过滤 |
| `+create` | 创建仪表盘，支持指定文件夹、从本地 JSON 文件导入数据 |
| `+export` | 导出仪表盘到本地 JSON 文件，Data 字段格式化为可读对象 |
| `+delete` | 删除仪表盘（带二次确认） |
| `+folders` | 列出所有仪表盘文件夹 |

### 2. 关键参数说明

**`+list`**
```bash
cls-cli dashboard +list --name "仪表盘名称"     # 精确匹配名称
cls-cli dashboard +list --folder-id <id>        # 按文件夹过滤
cls-cli dashboard +list --format table          # 表格格式输出
```

**`+create`**
```bash
cls-cli dashboard +create --name "名称" --folder-id <id>
cls-cli dashboard +create --name "名称" --folder-id <id> --from-file ./dashboard.json
```

**`+export`**
```bash
cls-cli dashboard +export --name "仪表盘名称"              # 导出到 ./仪表盘名.json
cls-cli dashboard +export --name "仪表盘名称" --output ./path/to/file.json
```

### 3. 导出格式说明

导出的 JSON 文件结构清晰，`Data` 字段被解析为完整 JSON 对象，而非转义字符串，便于阅读和二次导入：

```json
{
  "DashboardId": "dashboard-xxxx",
  "DashboardName": "仪表盘名称",
  "FolderId": "xxxx",
  "FolderName": "文件夹名称",
  "Data": {
    "panels": [...],
    "templating": {...},
    "time": [...],
    "timezone": "browser"
  },
  "DashboardTopicInfos": [...],
  "CreateTime": "2026-04-07 15:29:42"
}
```

---

## 二、自然语言交互成果

通过 CodeBuddy Code + cls-cli，用户可以直接用自然语言完成仪表盘的全生命周期管理，无需记忆命令格式。

### 2.1 创建空仪表盘

> 用户："创建一个 test_api_2 的仪表盘"

AI 自动引导完成：选择创建方式 → 展示文件夹列表 → 用户选择 → 执行创建

【截图占位：创建空仪表盘交互过程】

---

### 2.2 导出仪表盘

> 用户："帮我保存 test_2 的仪表盘到当前目录"

AI 自动执行导出，生成格式化 JSON 文件。

【截图占位：导出仪表盘交互过程】

---

### 2.3 根据导出文件创建仪表盘

> 用户："根据本地的 test_2.json，创建一个 test_api_3 的仪表盘"

AI 自动展示文件夹列表，用户选择目标文件夹后，从本地 JSON 导入数据完成创建。

【截图占位：从 JSON 文件导入创建仪表盘交互过程】

---

### 2.4 其他自然语言操作示例

【截图占位：其他交互示例】

---

## 三、Skill 能力：创建仪表盘引导流程

基于 CodeBuddy Code 的 Skill 机制，定义了 `cls-dashboard` 专属技能，触发后自动执行标准化创建流程：

### 流程设计

```
Step 1: 确认仪表盘名称
        ↓ 未提供则询问
Step 2: 选择创建方式
        [1] 创建空仪表盘
        [2] 从本地 JSON 文件导入 → 询问文件路径
        ↓
Step 3: 展示全部文件夹列表，用户选择
        ↓
Step 4: 执行创建命令
```

### Skill 文件位置

```
.codebuddy/skills/cls-dashboard/SKILL.md
```

### Skill 核心能力

- 自动列出所有文件夹（完整展示，不省略）
- 支持空仪表盘和 JSON 导入两种创建方式
- 引导式交互，降低用户操作门槛
- 兼容新旧两种 JSON 格式（Data 为对象或字符串均可）

---

## 四、总结

| 能力项 | 状态 |
|--------|------|
| 仪表盘 CRUD（list/create/delete） | ✅ 完成 |
| 文件夹查询（folders） | ✅ 完成 |
| 仪表盘导出（export，格式化 JSON） | ✅ 完成 |
| 从本地 JSON 导入创建 | ✅ 完成 |
| 自然语言交互（CodeBuddy Code） | ✅ 完成 |
| 创建仪表盘引导 Skill | ✅ 完成 |

---

## 五、后续计划

- 支持 ModifyDashboard（修改仪表盘内容）
- 支持更多 CLS API（日志主题、告警等）
- 完善 Skill，覆盖更多常用操作场景
