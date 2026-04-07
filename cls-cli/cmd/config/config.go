package config

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tencentcloud/cls-cli/internal/cmdutil"
	"github.com/tencentcloud/cls-cli/internal/core"
	"github.com/tencentcloud/cls-cli/internal/output"
	"golang.org/x/term"
)

func NewCmdConfig(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "配置管理",
		Long:  "管理 cls-cli 的认证配置和默认参数",
	}

	cmd.AddCommand(newCmdConfigInit(f))
	cmd.AddCommand(newCmdConfigShow(f))
	cmd.AddCommand(newCmdConfigAdd(f))
	cmd.AddCommand(newCmdConfigList(f))
	cmd.AddCommand(newCmdConfigUse(f))
	return cmd
}

func newCmdConfigInit(f *cmdutil.Factory) *cobra.Command {
	var (
		secretID  string
		secretKey string
		region    string
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "初始化配置（设置默认账号的 SecretId、SecretKey、Region）",
		Long: `交互式或通过参数配置腾讯云 API 凭证，写入 default 账号。

示例:
  # 交互式配置
  cls-cli config init

  # 直接传参
  cls-cli config init --secret-id AKIDxxx --secret-key yyy --region ap-guangzhou`,
		RunE: func(cmd *cobra.Command, args []string) error {
			reader := bufio.NewReader(os.Stdin)

			if secretID == "" {
				fmt.Print("SecretId: ")
				input, _ := reader.ReadString('\n')
				secretID = strings.TrimSpace(input)
			}
			if secretKey == "" {
				fmt.Print("SecretKey: ")
				input, _ := reader.ReadString('\n')
				secretKey = strings.TrimSpace(input)
			}
			if region == "" {
				fmt.Print("默认 Region (如 ap-guangzhou): ")
				input, _ := reader.ReadString('\n')
				region = strings.TrimSpace(input)
				if region == "" {
					region = "ap-guangzhou"
				}
			}

			if secretID == "" || secretKey == "" {
				return fmt.Errorf("SecretId 和 SecretKey 不能为空")
			}

			cfg := &core.CliConfig{
				SecretID:  secretID,
				SecretKey: secretKey,
				Region:    region,
			}

			if err := core.SaveConfig(cfg); err != nil {
				return fmt.Errorf("保存配置失败: %w", err)
			}

			output.PrintSuccess(fmt.Sprintf("配置已保存到 %s", core.GetConfigPath()))
			return nil
		},
	}

	cmd.Flags().StringVar(&secretID, "secret-id", "", "腾讯云 SecretId")
	cmd.Flags().StringVar(&secretKey, "secret-key", "", "腾讯云 SecretKey")
	cmd.Flags().StringVar(&region, "region", "", "默认 Region (如 ap-guangzhou)")
	return cmd
}

func newCmdConfigAdd(f *cmdutil.Factory) *cobra.Command {
	var alias string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "添加或更新一个账号配置",
		Long: `交互式添加账号（SecretId、SecretKey 使用终端静默输入，不回显）。

示例:
  cls-cli config add --alias work
  cls-cli config add --alias default`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if alias == "" {
				return fmt.Errorf("必须指定 --alias <别名>")
			}

			reader := bufio.NewReader(os.Stdin)

			fmt.Print("SecretId: ")
			secretID, err := readPassword()
			if err != nil {
				// 静默输入失败，退而使用普通输入
				input, _ := reader.ReadString('\n')
				secretID = strings.TrimSpace(input)
			} else {
				fmt.Println() // 换行
			}

			fmt.Print("SecretKey: ")
			secretKey, err := readPassword()
			if err != nil {
				input, _ := reader.ReadString('\n')
				secretKey = strings.TrimSpace(input)
			} else {
				fmt.Println()
			}

			fmt.Print("Region (如 ap-guangzhou): ")
			regionInput, _ := reader.ReadString('\n')
			region := strings.TrimSpace(regionInput)
			if region == "" {
				region = "ap-guangzhou"
			}

			if secretID == "" || secretKey == "" {
				return fmt.Errorf("SecretId 和 SecretKey 不能为空")
			}

			// 加载现有配置（不存在则新建）
			mc, err := core.LoadRawConfig()
			if err != nil {
				mc = &core.MultiConfig{
					Current:  alias,
					Profiles: map[string]core.Profile{},
				}
			}

			mc.Profiles[alias] = core.Profile{
				SecretID:  secretID,
				SecretKey: secretKey,
				Region:    region,
			}

			// 如果是第一个账号，自动设为当前
			if len(mc.Profiles) == 1 || mc.Current == "" {
				mc.Current = alias
			}

			if err := core.SaveRawConfig(mc); err != nil {
				return fmt.Errorf("保存配置失败: %w", err)
			}

			output.PrintSuccess(fmt.Sprintf("账号 [%s] 已保存", alias))
			return nil
		},
	}

	cmd.Flags().StringVar(&alias, "alias", "", "账号别名（必填）")
	return cmd
}

func newCmdConfigList(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "列出所有已保存的账号别名",
		RunE: func(cmd *cobra.Command, args []string) error {
			mc, err := core.LoadRawConfig()
			if err != nil {
				return err
			}

			if len(mc.Profiles) == 0 {
				fmt.Println("暂无账号，请运行 cls-cli config add --alias <别名> 添加")
				return nil
			}

			aliases := make([]string, 0, len(mc.Profiles))
			for alias := range mc.Profiles {
				aliases = append(aliases, alias)
			}
			sort.Strings(aliases)

			for _, alias := range aliases {
				p := mc.Profiles[alias]
				if alias == mc.Current {
					fmt.Printf("* %-20s  %s  (当前)\n", alias, p.Region)
				} else {
					fmt.Printf("  %-20s  %s\n", alias, p.Region)
				}
			}
			return nil
		},
	}
}

func newCmdConfigUse(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "use <别名>",
		Short: "切换当前账号",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			alias := args[0]

			mc, err := core.LoadRawConfig()
			if err != nil {
				return err
			}

			if _, ok := mc.Profiles[alias]; !ok {
				return fmt.Errorf("账号 [%s] 不存在，请先运行 cls-cli config add --alias %s", alias, alias)
			}

			mc.Current = alias
			if err := core.SaveRawConfig(mc); err != nil {
				return fmt.Errorf("保存配置失败: %w", err)
			}

			fmt.Printf("已切换到账号 [%s]\n", alias)
			return nil
		},
	}
}

func newCmdConfigShow(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "显示当前账号别名和 Region",
		RunE: func(cmd *cobra.Command, args []string) error {
			mc, err := core.LoadRawConfig()
			if err != nil {
				return err
			}

			current := mc.Current
			if current == "" {
				current = "default"
			}

			profile, ok := mc.Profiles[current]
			if !ok {
				return fmt.Errorf("当前账号 [%s] 不存在", current)
			}

			output.FormatOutput(map[string]interface{}{
				"current_alias": current,
				"region":        profile.Region,
				"config_dir":    core.GetConfigDir(),
			}, f.Format)
			return nil
		},
	}
}

// readPassword 使用终端静默输入（不回显），返回输入内容
func readPassword() (string, error) {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return "", fmt.Errorf("非终端环境")
	}
	b, err := term.ReadPassword(fd)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

func maskSecret(s string) string {
	if len(s) <= 8 {
		return "****"
	}
	return s[:4] + strings.Repeat("*", len(s)-8) + s[len(s)-4:]
}
