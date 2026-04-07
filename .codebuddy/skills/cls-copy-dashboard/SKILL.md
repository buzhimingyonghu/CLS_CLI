---
name: cls-copy-dashboard
description: 引导用户完成 CLS 仪表盘的跨账号复制，支持修改数据源 TopicId 以适配目标账号。当用户提到"复制仪表盘"、"把仪表盘迁移到另一个账号"、"跨账号同步仪表盘"、"把 A 账号的仪表盘复制到 B 账号"，或者需要在不同腾讯云账号之间共享仪表盘配置时，务必使用此 skill。即使用户只说"帮我把这个仪表盘给另一个账号用"，也应触发此 skill。
allowed-tools: Bash
---

# CLS 跨账号仪表盘复制向导

你是 CLS 跨账号仪表盘复制专家，通过 cls-cli 完成从源账号导出仪表盘、修改数据源、在目标账号创建的完整流程。

---

## 前置检查

执行前先确认 cls-cli 可用：
```bash
cls-cli config list 2>&1
```

列出当前所有已配置账号，确认源账号和目标账号都已配置。如未配置，引导用户运行 `cls-setup` skill 完成初始化。

---

## 完整复制流程

### 第一步：确认源账号和仪表盘

询问用户：
```
请提供以下信息：
1. 源账号别名（运行 cls-cli config list 查看）：
2. 要复制的仪表盘名称：
3. 目标账号别名：
4. 新仪表盘名称（在目标账号中的名字）：
```

### 第二步：切换到源账号并导出仪表盘

切换源账号：
```bash
cls-cli config use <源账号别名> 2>&1
```

验证切换成功：
```bash
cls-cli config show 2>&1
```

导出仪表盘到本地（文件名用仪表盘名称）：
```bash
cls-cli dashboard +export --name "<仪表盘名称>" --output ./<仪表盘名称>.json 2>&1
```

导出成功后，读取并展示当前数据源信息：
```bash
python3 -c "
import json
with open('./<仪表盘名称>.json') as f:
    d = json.load(f)
print('当前数据源：')
for item in d['Data']['templating']['list']:
    if item.get('type') == 'datasource':
        print(f'  {item[\"label\"]}（{item[\"name\"]}）：TopicId = {item[\"default\"][\"TopicId\"]}')
"
```

### 第三步：询问是否修改数据源

向用户展示：
```
导出的仪表盘包含以下数据源：

  会话数据源（ds）    ：TopicId = <当前值>
  指标数据源（metricDs）：TopicId = <当前值>

是否需要修改为目标账号的 TopicId？
[1] 不修改，直接使用原 TopicId 创建
[2] 修改数据源 TopicId
```

**如果用户选择 [2]**，依次询问：
```
请提供目标账号的【会话数据源】新 TopicId（留空则不修改）：
请提供目标账号的【指标数据源】新 TopicId（留空则不修改）：
```

收到后执行修改（只修改非空的字段）：
```bash
python3 -c "
import json

filepath = './<仪表盘名称>.json'
ds_topic_id = '<新会话数据源TopicId>'       # 空则填 ''
metric_topic_id = '<新指标数据源TopicId>'   # 空则填 ''

with open(filepath) as f:
    d = json.load(f)

for item in d['Data']['templating']['list']:
    if item.get('type') == 'datasource':
        if item.get('name') == 'ds' and ds_topic_id:
            item['default']['TopicId'] = ds_topic_id
        if item.get('name') == 'metricDs' and metric_topic_id:
            item['default']['TopicId'] = metric_topic_id

with open(filepath, 'w') as f:
    json.dump(d, f, ensure_ascii=False, indent=2)

print('✓ 数据源已更新：')
for item in d['Data']['templating']['list']:
    if item.get('type') == 'datasource':
        print(f'  {item[\"label\"]}（{item[\"name\"]}）：TopicId = {item[\"default\"][\"TopicId\"]}')
"
```

修改完成后展示更新结果，请用户确认。

### 第四步：切换到目标账号

```bash
cls-cli config use <目标账号别名> 2>&1
```

验证切换：
```bash
cls-cli config show 2>&1
cls-cli dashboard +folders 2>&1
```

### 第五步：选择目标文件夹

列出目标账号所有文件夹，**必须全部展示，不能省略**：
```bash
cls-cli dashboard +folders 2>&1
```

解析输出，将所有文件夹按序号完整列出，格式如下：
```
  1. 文件夹名称A  (ID: folder-xxxx)
  2. 文件夹名称B  (ID: folder-yyyy)
  ...（全部列出，不省略）
  0. 不放入任何文件夹
```

使用 AskUserQuestion 工具让用户选择目标文件夹。

### 第六步：在目标账号创建仪表盘

根据用户选择的文件夹，组装创建命令：

**指定了文件夹：**
```bash
cls-cli dashboard +create --name "<新仪表盘名称>" --folder-id "<目标文件夹ID>" --from-file ./<仪表盘名称>.json 2>&1
```

**不放入文件夹：**
```bash
cls-cli dashboard +create --name "<新仪表盘名称>" --from-file ./<仪表盘名称>.json 2>&1
```

向用户确认后执行，创建成功后展示结果：
```
✓ 跨账号复制完成！

  源账号    ：<源账号别名>
  源仪表盘  ：<原仪表盘名称>
  目标账号  ：<目标账号别名>
  新仪表盘  ：<新仪表盘名称>（ID: dashboard-xxxx）
  目标文件夹：<文件夹名称>（无则显示"无"）
  会话数据源：<TopicId>（未修改则显示原值）
  指标数据源：<TopicId>（未修改则显示原值）
```

---

## 注意事项

- 导出的 JSON 文件会保留在当前目录，可重复用于多次复制
- `--from-file` 支持三种格式自动识别：新导出格式（Data 为对象）、旧格式（Data 为字符串）、DescribeDashboards 完整响应
- 如果目标账号中已存在同名仪表盘，会报错 `DashboardNameConflict`，换一个名字重试
- 复制完成后建议切换回默认账号：`cls-cli config use default`
