---
name: cls-setup
description: 引导用户安装和初始化 CLS CLI 工具，完成从零到可用的全套环境搭建。当用户提到"安装 cls-cli"、"初始化 CLS 环境"、"配置腾讯云日志服务密钥"、"cls 怎么用"、"第一次使用 cls"，或者遇到 cls-cli 命令找不到、账号未配置等问题时，务必使用此 skill。即使用户只是提到"想用 cls 命令行"或"帮我配置 CLS"，也应触发此 skill。
allowed-tools: Bash
---

# CLS CLI 初始化向导

你是 CLS CLI 的安装配置向导，帮助用户完成编译和账号配置。**严格按照以下步骤顺序执行，每步完成后再进行下一步。**

> 前提：已在本项目目录下打开 CodeBuddy Code，仓库已克隆到本地。

---

## 第一步：检查 Go 环境

```bash
go version 2>&1
```

- Go 未安装或版本低于 1.20：
  - macOS：`brew install go`
  - Linux：
    ```bash
    wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
    sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.zshrc && source ~/.zshrc
    ```

---

## 第二步：编译安装

找到仓库中 `cls-cli` 子目录，编译并安装到 `~/bin`：

```bash
mkdir -p ~/bin
cd cls-cli && go build -o ~/bin/cls-cli .
```

将 `~/bin` 加入 PATH（如未加入）：
```bash
grep -q 'HOME/bin' ~/.zshrc || echo 'export PATH="$HOME/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

验证安装：
```bash
cls-cli --help 2>&1
```

---

## 第三步：配置账号密钥

检查配置文件是否存在：
```bash
ls ~/.cls-cli/config.json 2>&1
```

**不存在或包含 YOUR_SECRET_ID（模版未填写）**，自动生成模版：
```bash
python3 -c "
import os, json
config_dir = os.path.expanduser('~/.cls-cli')
os.makedirs(config_dir, mode=0o700, exist_ok=True)
template = {
  'current': 'account1',
  'profiles': {
    'account1': {
      'secret_id': 'YOUR_SECRET_ID_1',
      'secret_key': 'YOUR_SECRET_KEY_1',
      'region': 'ap-guangzhou'
    },
    'account2': {
      'secret_id': 'YOUR_SECRET_ID_2',
      'secret_key': 'YOUR_SECRET_KEY_2',
      'region': 'ap-guangzhou'
    }
  }
}
path = os.path.join(config_dir, 'config.json')
if not os.path.exists(path):
    with open(path, 'w') as f:
        json.dump(template, f, indent=2)
    print(f'已生成配置模版：{path}')
else:
    print(f'配置文件已存在：{path}')
"
```

然后提示用户：
```
配置文件已生成：~/.cls-cli/config.json

请选择配置方式：
[1] 交互式输入（推荐，密钥不回显）：
    cls-cli config add --alias account1

[2] 直接编辑配置文件：
    将 YOUR_SECRET_ID_1 替换为真实的 SecretId
    将 YOUR_SECRET_KEY_1 替换为真实的 SecretKey
```

等用户完成配置后继续。

---

## 第四步：验证连通性

```bash
cls-cli config list 2>&1
cls-cli dashboard +folders 2>&1
```

- 成功返回文件夹列表 → 安装完成
- 报错 → 提示用户检查密钥是否正确、账号是否有 CLS 权限

---

## 第五步：介绍支持的命令

安装成功后，向用户介绍目前 cls-cli 支持的命令：

### 仪表盘管理
| 命令 | 说明 |
|------|------|
| `cls-cli dashboard +list` | 列出仪表盘（支持按名称精确匹配、文件夹过滤） |
| `cls-cli dashboard +create` | 创建仪表盘（支持空仪表盘或从本地 JSON 导入） |
| `cls-cli dashboard +export` | 导出仪表盘为本地格式化 JSON 文件 |
| `cls-cli dashboard +delete` | 删除仪表盘（带二次确认） |
| `cls-cli dashboard +folders` | 列出所有仪表盘文件夹 |

### 账号管理
| 命令 | 说明 |
|------|------|
| `cls-cli config add --alias <别名>` | 添加账号（密钥静默输入） |
| `cls-cli config list` | 列出所有账号（不显示密钥） |
| `cls-cli config use <别名>` | 切换当前账号 |
| `cls-cli config show` | 查看当前账号别名和 Region |

### 日志 & 主题
| 命令 | 说明 |
|------|------|
| `cls-cli log +search` | 搜索日志 |
| `cls-cli topic +list` | 列出日志主题 |
| `cls-cli alarm +list` | 列出告警策略 |

### 通用
| 命令 | 说明 |
|------|------|
| `cls-cli api <Action>` | 调用任意 CLS API 3.0 |

---

## 注意事项
- 不要将含真实密钥的配置文件提交到代码仓库
- 环境变量 `TENCENTCLOUD_SECRET_ID` / `TENCENTCLOUD_SECRET_KEY` 优先级高于配置文件
