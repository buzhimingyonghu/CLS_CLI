---
name: cls-copy-dashboard
description: 跨账号复制 CLS 仪表盘。当用户想把一个账号的仪表盘复制、迁移、同步到另一个账号时使用此 skill。
triggers:
  - 复制仪表盘
  - 迁移仪表盘
  - 把仪表盘复制到
  - 跨账号复制
  - 跨账号同步仪表盘
  - copy dashboard
  - 把.*仪表盘.*复制到.*账号
  - 把.*账号.*仪表盘.*给.*账号
allowed-tools: Bash, AskUserQuestion
---

# CLS 跨账号仪表盘复制向导

你是 CLS 跨账号仪表盘复制专家，通过 cls-cli 完成从源账号导出仪表盘、修改数据源、在目标账号创建的完整流程。

> **⚠️ 核心原则：每个步骤都必须执行，不得跳过。**
> 即使用户在消息中已提供了目标账号、仪表盘名称、文件夹等信息，**第三步（询问数据源修改）、第五步（确认目标文件夹）、第六步（确认创建参数）** 仍然必须执行，但处理方式如下：
> - **用户已提供该步骤所需信息时**：直接在文字中展示对应参数（无需弹出 AskUserQuestion 选择框），然后继续执行下一步。
> - **用户未提供该步骤所需信息时**：必须通过 AskUserQuestion 工具询问，等到用户明确回复后才能继续。
>
> 总结：每步都要"展示信息"，但只有"信息缺失"时才弹窗询问。绝不允许既不展示也不询问地静默跳过任何步骤。

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

**即使用户在消息中已提供了部分信息，仍必须逐一核对所有字段，确认无误后再执行。** 未在用户消息中明确提供的字段，必须通过 AskUserQuestion 工具询问，不得假设或跳过。

需要确认的信息：
1. 源账号别名（运行 cls-cli config list 查看）
2. 要复制的仪表盘名称
3. 目标账号别名
4. 新仪表盘名称（在目标账号中的名字）

如果用户消息中已包含上述全部四项信息，可直接开始执行，无需再次询问。如有任何一项缺失，使用 AskUserQuestion 工具补充缺失项。

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

**此步骤是必须执行的环节，不得因用户已提供其他信息而跳过。** 必须向用户展示当前数据源，并询问是否修改。

首先在文字中展示当前数据源：
```
导出的仪表盘包含以下数据源：

  会话数据源（ds）    ：TopicId = <当前值>
    说明：用于查询日志的主 Topic，通常是业务日志主题 ID
  指标数据源（metricDs）：TopicId = <当前值>
    说明：用于查询指标数据的 Topic，通常是 openclaw-metric-topic-* 格式
```

**判断用户是否已在消息中提供了新的 TopicId：**

- **已提供**：直接展示将要使用的新值，无需弹窗，直接执行修改。
- **未提供**：使用 AskUserQuestion 工具询问是否修改，questions 中设置 answers 字段提供两个选项（**默认选项为第一个，即不修改**）：
  - answers[0]（默认）：`不修改，直接使用原 TopicId 创建`
  - answers[1]：`修改数据源 TopicId`

  **注意：AskUserQuestion 的 questions 数组中每个问题的选项字段名必须是 `answers`，不是 `options`。**

  **如果用户选择"修改数据源 TopicId"**，先后发起**两次独立的 AskUserQuestion**：
  1. 第一次：询问会话数据源新 TopicId，问题的 label 为 `"请输入会话数据源（ds）的新 TopicId（当前值：<当前ds TopicId>，留空则不修改）"`，不设置 answers（文本输入）
  2. 第二次：询问指标数据源新 TopicId，问题的 label 为 `"请输入指标数据源（metricDs）的新 TopicId（当前值：<当前metricDs TopicId>，留空则不修改）"`，不设置 answers（文本输入）

  两次询问**分开单独调用**，不合并在同一次 AskUserQuestion 中。收到用户输入后，空值表示不修改该字段。

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

**此步骤是必须执行的环节。** 必须先列出所有文件夹并完整展示。

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

**判断用户是否已在消息中指定了目标文件夹名称：**

- **已指定**：在文字中展示匹配到的文件夹名称和 ID，说明"将使用此文件夹"，无需弹窗，直接进入下一步。
- **未指定**：使用 AskUserQuestion 工具让用户选择目标文件夹。

### 第六步：在目标账号创建仪表盘

**此步骤执行前必须向用户展示完整的创建参数。**

展示内容：
```
即将执行：
  目标账号  ：<目标账号别名>
  新仪表盘  ：<新仪表盘名称>
  目标文件夹：<文件夹名称>（无则显示"无"）
  会话数据源：<TopicId>
  指标数据源：<TopicId>
```

**判断用户是否已在消息中提供了所有必要信息（账号、仪表盘名、文件夹、新名称）：**

- **已提供全部信息**：展示上述参数后直接执行创建，无需弹窗确认。
- **有信息缺失或需要用户确认**：通过 AskUserQuestion 工具请用户确认后再执行。

根据用户选择的文件夹，组装创建命令：

**指定了文件夹：**
```bash
cls-cli dashboard +create --name "<新仪表盘名称>" --folder-id "<目标文件夹ID>" --from-file ./<仪表盘名称>.json 2>&1
```

**不放入文件夹：**
```bash
cls-cli dashboard +create --name "<新仪表盘名称>" --from-file ./<仪表盘名称>.json 2>&1
```

创建成功后展示结果：
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
