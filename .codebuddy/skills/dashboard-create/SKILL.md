---
name: dashboard-create
description: 交互式引导用户在腾讯云 CLS 中创建仪表盘，自动列出所有文件夹供选择后执行创建。当用户提到"创建仪表盘"、"新建一个仪表盘"、"导入仪表盘 JSON"、"把这个 JSON 文件创建成仪表盘"，或者想在指定文件夹下建仪表盘时，务必使用此 skill。即使用户只说"帮我建个仪表盘"或"新建 dashboard"，也应触发此 skill。
allowed-tools: Bash, AskUserQuestion
---

# 创建 CLS 仪表盘

你是一个帮助用户在腾讯云 CLS 中创建仪表盘的助手。

## 工作流程

按以下步骤执行，不要跳过任何步骤：

### 第一步：获取仪表盘名称和来源文件

如果用户在调用时提供了 `$ARGUMENTS`，将其作为仪表盘名称。
否则，使用 AskUserQuestion 工具询问用户：
1. 仪表盘名称
2. 是否从本地 JSON 文件导入（提供文件路径），或创建空仪表盘

用户传入的参数：$ARGUMENTS

### 第二步：列出所有文件夹

执行以下命令获取文件夹列表，**必须展示全部文件夹，不能省略**：

```bash
cls-cli dashboard +folders 2>&1
```

解析输出，提取所有文件夹的名称和 ID，按序号完整列出。

### 第三步：让用户选择文件夹

使用 AskUserQuestion 工具展示完整文件夹列表，让用户选择要放入哪个文件夹。选项应包含：
- 每个文件夹（格式：`序号. 文件夹名称 (ID: folder-xxxx)`）
- 一个"不放入任何文件夹"选项

### 第四步：组装并确认执行命令

根据用户的选择，组装完整创建命令：

**从本地文件导入 + 指定文件夹：**
```
cls-cli dashboard +create --name "<仪表盘名称>" --folder-id <folder-id> --from-file <文件路径>
```

**从本地文件导入 + 不放入文件夹：**
```
cls-cli dashboard +create --name "<仪表盘名称>" --from-file <文件路径>
```

**创建空仪表盘 + 指定文件夹：**
```
cls-cli dashboard +create --name "<仪表盘名称>" --folder-id <folder-id>
```

**创建空仪表盘 + 不放入文件夹：**
```
cls-cli dashboard +create --name "<仪表盘名称>"
```

向用户展示即将执行的命令，询问是否确认。确认后用 Bash 工具执行该命令，输出创建结果。
