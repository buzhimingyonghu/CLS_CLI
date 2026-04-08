# CLS CLI

腾讯云日志服务 (Cloud Log Service) 命令行工具，专为 **AI Agent** 和开发者设计。

> 安装后配合 CodeBuddy Code / Claude Code / Cursor 等 AI 编程工具使用，用自然语言即可完成日志检索、仪表盘管理、告警配置等操作——无需记忆任何命令。

## CodeBuddy使用示例

使用 CodeBuddy Code 打开本项目目录后，执行以下命令完成初始化：

```
# 初始化（新用户首次使用）——会自动完成编译、配置、验证等步骤
/cls_setup

```
---

## 环境要求

- Go 1.20+
- macOS / Linux
- 腾讯云账号（SecretId / SecretKey）

---

## 快速开始

### 1. 安装 Go（如未安装）

**macOS：**
```bash
brew install go
```

**Linux：**
```bash
# 下载并安装 Go 1.21
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

验证安装：
```bash
go version
```

### 2. 编译安装

```bash
cd cls-cli

# 编译并安装到 ~/bin（推荐，无需 sudo）
mkdir -p ~/bin
go build -o ~/bin/cls-cli .

# 加入 PATH（写入 ~/.zshrc 或 ~/.bashrc）
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

验证安装：
```bash
cls-cli --version
```

### 3. 配置腾讯云密钥

配置文件位于 `~/.cls-cli/config.json`，支持多账号管理。

**方式一：交互式添加账号（推荐，密钥不回显）**
```bash
# 添加账号（别名自定义，如 default、account1、work 等）
cls-cli config add --alias default

# 列出所有账号
cls-cli config list

# 切换账号
cls-cli config use default
```

**方式二：直接编辑配置文件模版**

复制以下模版到 `~/.cls-cli/config.json`，填入真实密钥：
```json
{
  "current": "account1",
  "profiles": {
    "account1": {
      "secret_id": "YOUR_SECRET_ID_1",
      "secret_key": "YOUR_SECRET_KEY_1",
      "region": "ap-guangzhou"
    },
    "account2": {
      "secret_id": "YOUR_SECRET_ID_2",
      "secret_key": "YOUR_SECRET_KEY_2",
      "region": "ap-guangzhou"
    }
  }
}
```

> **安全提示**：请勿将包含真实密钥的配置文件提交到代码仓库。

**方式三：环境变量（临时覆盖）**
```bash
export TENCENTCLOUD_SECRET_ID=xxx
export TENCENTCLOUD_SECRET_KEY=xxx
export CLS_DEFAULT_REGION=ap-guangzhou
```

查看当前账号：
```bash
cls-cli config show
```

### 4. 验证连通性

```bash
cls-cli topic +list
```

---

## 使用方式

### 推荐：自然语言 + AI Agent

在 CodeBuddy Code 等工具中直接说：

```
"帮我查一下最近1小时 topic xxx 里有没有 ERROR 日志"
"创建一个仪表盘，名字叫业务大盘，放到测试文件夹"
"导出 OpenClaw 成本治理仪表盘到本地"
"根据本地的 dashboard.json 创建一个新仪表盘"
```

AI 会自动调用 cls-cli 完成操作。

### 直接使用 CLI

```bash
# 搜索日志
cls-cli log +search --topic-id xxx --query "level:ERROR" --from "1 hour ago"

# 列出日志主题
cls-cli topic +list

# 列出仪表盘
cls-cli dashboard +list --format table

# 导出仪表盘
cls-cli dashboard +export --name "仪表盘名称"

# 创建仪表盘
cls-cli dashboard +create --name "名称" --folder-id <id>

# 从本地 JSON 导入创建仪表盘
cls-cli dashboard +create --name "名称" --folder-id <id> --from-file ./dashboard.json

# 列出仪表盘文件夹
cls-cli dashboard +folders
```

---

## 命令总览

### 日志 & 主题

| 域 | 快捷命令 | 说明 |
|---|---|---|
| `log` | `+search` `+context` `+tail` `+histogram` `+download` | 日志检索与分析 |
| `topic` | `+list` `+create` `+info` `+delete` `+logsets` | 日志主题管理 |

### 仪表盘

| 命令 | 参数 | 说明 |
|---|---|---|
| `dashboard +list` | `--name` `--id` `--folder-id` `--format` `--offset` `--limit` | 列出仪表盘（名称精确匹配） |
| `dashboard +create` | `--name`（必填）`--folder-id` `--from-file` `--data` `--tag` | 创建仪表盘 |
| `dashboard +export` | `--name`（必填）`--output` | 导出仪表盘为本地 JSON 文件 |
| `dashboard +delete` | `--id`（必填） | 删除仪表盘（带确认） |
| `dashboard +folders` | `--name` `--id` `--limit` | 列出仪表盘文件夹 |

**`+export` 导出格式说明：**

导出的 JSON 文件中 `Data` 字段被解析为完整对象（非转义字符串），便于阅读和二次导入：

```json
{
  "DashboardId": "dashboard-xxxx",
  "DashboardName": "仪表盘名称",
  "FolderId": "xxxx",
  "FolderName": "文件夹名称",
  "Data": {
    "panels": [...],
    "templating": {...},
    "time": ["now-7d", "now"],
    "timezone": "browser"
  },
  "DashboardTopicInfos": [...]
}
```

`+create --from-file` 支持以下三种 JSON 格式自动识别：
1. `+export` 导出格式（`Data` 为对象）
2. 旧格式（`Data` 为转义字符串）
3. `DescribeDashboards` API 完整响应格式

### 告警 & 采集

| 域 | 快捷命令 | 说明 |
|---|---|---|
| `alarm` | `+list` `+history` `+create` `+delete` `+notices` | 告警策略管理 |
| `machinegroup` / `mg` | `+list` `+create` `+delete` `+info` `+status` | 机器组管理 |
| `collector` / `col` | `+list` `+create` `+delete` `+info` `+guide` | 采集配置管理 |
| `loglistener` / `ll` | `+install` `+init` `+start` `+stop` `+restart` `+status` `+uninstall` `+check` | LogListener 管理 |

### 通用

| 命令 | 说明 |
|---|---|
| `config add --alias <别名>` | 交互式添加账号（密钥不回显） |
| `config list` | 列出所有账号，`*` 标记当前账号 |
| `config use <别名>` | 切换当前账号 |
| `config show` | 查看当前账号信息（不显示密钥） |
| `config init` | 初始化腾讯云密钥配置（旧式单账号） |
| `api <Action>` | 通用 API 调用，支持所有 CLS API 3.0 |

---

## CodeBuddy Code Skills

Skills 文件位于项目的 `.codebuddy/skills/` 目录下，使用 CodeBuddy Code 进入项目目录后即可直接调用。

在 CodeBuddy Code 中内置了三个 Skill，配合自然语言使用更高效：

| Skill | 调用方式 | 说明 |
|---|---|---|
| `cls-setup` | `/cls-setup` | 新用户初始化向导：编译 → 生成配置模版 → 验证连通性 → 介绍命令 |
| `dashboard-create` | `/dashboard-create [仪表盘名称]` | 交互式创建仪表盘：自动列出所有文件夹供选择，支持从本地 JSON 导入 |
| `cls-copy-dashboard` | `/cls-copy-dashboard` | 跨账号复制仪表盘：导出 → 修改数据源 TopicId → 切换账号 → 选文件夹 → 创建 |


---

## 架构设计

```
用户 (自然语言)
    ↓
AI Agent (CodeBuddy Code / Claude Code / Cursor)
    ↓
cls-cli (命令行工具)
    ↓
腾讯云 CLS API 3.0
```

- **两层命令体系**：`+` 前缀快捷命令（语义化）+ `api` 通用命令（覆盖全量 API）
- **极简依赖**：仅依赖 `cobra`，签名算法自行实现，无腾讯云 SDK 依赖
- **灵活时间解析**：支持 `15 minutes ago`、`today`、`2024-01-01` 等多种格式
- **多输出格式**：`--format json|pretty|table|csv`
- **Dry-run 模式**：`--dry-run` 预览请求，不实际执行

---

## License

MIT
