package dashboard

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tencentcloud/cls-cli/internal/cmdutil"
	"github.com/tencentcloud/cls-cli/internal/output"
)

func RegisterShortcuts(rootCmd *cobra.Command, f *cmdutil.Factory) {
	dashboardCmd := &cobra.Command{
		Use:   "dashboard",
		Short: "仪表盘管理",
		Long:  "仪表盘的查询和管理命令",
	}

	dashboardCmd.AddCommand(newListCmd(f))
	dashboardCmd.AddCommand(newExportCmd(f))
	dashboardCmd.AddCommand(newCreateCmd(f))
	dashboardCmd.AddCommand(newDeleteCmd(f))
	dashboardCmd.AddCommand(newFolderListCmd(f))

	rootCmd.AddCommand(dashboardCmd)
}

func newExportCmd(f *cmdutil.Factory) *cobra.Command {
	var (
		dashboardName string
		outputPath    string
	)

	cmd := &cobra.Command{
		Use:   "+export",
		Short: "导出仪表盘到本地 JSON 文件",
		Long: `根据仪表盘名称精确匹配，导出为格式化的本地 JSON 文件。

示例:
  cls-cli dashboard +export --name "test_2"
  cls-cli dashboard +export --name "test_2" --output ./my-dashboard.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if dashboardName == "" {
				return fmt.Errorf("--name 参数必填")
			}

			clsClient, err := f.CLSClient()
			if err != nil {
				return err
			}

			params := map[string]interface{}{
				"Offset": int64(0),
				"Limit":  int64(100),
				"Filters": []map[string]interface{}{
					{"Key": "dashboardName", "Values": []string{dashboardName}},
				},
			}

			result, err := clsClient.Call("DescribeDashboards", params)
			if err != nil {
				return err
			}

			// 客户端精确过滤
			var matched interface{}
			if resp, ok := result["Response"].(map[string]interface{}); ok {
				if infos, ok := resp["DashboardInfos"].([]interface{}); ok {
					for _, item := range infos {
						if m, ok := item.(map[string]interface{}); ok {
							if m["DashboardName"] == dashboardName {
								matched = item
								break
							}
						}
					}
				}
			}

			if matched == nil {
				return fmt.Errorf("未找到名称为 %q 的仪表盘", dashboardName)
			}

			// 默认文件名
			if outputPath == "" {
				outputPath = fmt.Sprintf("./%s.json", dashboardName)
			}

			// 将 Data 字段从 JSON 字符串解析为对象，方便阅读和再导入
			if m, ok := matched.(map[string]interface{}); ok {
				if dataStr, ok := m["Data"].(string); ok && dataStr != "" {
					var dataObj interface{}
					if err := json.Unmarshal([]byte(dataStr), &dataObj); err == nil {
						m["Data"] = dataObj
					}
				}
			}

			outputBytes, err := json.MarshalIndent(matched, "", "  ")
			if err != nil {
				return fmt.Errorf("序列化失败: %w", err)
			}

			if err := os.WriteFile(outputPath, outputBytes, 0644); err != nil {
				return fmt.Errorf("写入文件失败: %w", err)
			}

			fmt.Fprintf(f.IOStreams.Out, "✓ 已导出到 %s\n", outputPath)
			return nil
		},
	}

	cmd.Flags().StringVar(&dashboardName, "name", "", "仪表盘名称（精确匹配，必填）")
	cmd.Flags().StringVar(&outputPath, "output", "", "输出文件路径（默认 ./dashboard-<name>.json）")
	return cmd
}

func newListCmd(f *cmdutil.Factory) *cobra.Command {
	var (
		dashboardID     string
		dashboardName   string
		dashboardRegion string
		folderID        string
		tagKey          string
		offset          int64
		limit           int64
	)

	cmd := &cobra.Command{
		Use:   "+list",
		Short: "列出仪表盘",
		Long: `列出当前账号下的仪表盘列表，支持多种过滤条件。

示例:
  cls-cli dashboard +list
  cls-cli dashboard +list --id dashboard-xxxx
  cls-cli dashboard +list --name "业务大盘"
  cls-cli dashboard +list --folder-id folder-xxxx
  cls-cli dashboard +list --tag-key mykey
  cls-cli dashboard +list --format table`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if f.DryRun {
				fmt.Fprintf(f.IOStreams.Out, "DRY RUN:\n  Action: DescribeDashboards\n")
				return nil
			}

			clsClient, err := f.CLSClient()
			if err != nil {
				return err
			}

			params := map[string]interface{}{
				"Offset": offset,
				"Limit":  limit,
			}

			var filters []map[string]interface{}
			if dashboardID != "" {
				filters = append(filters, map[string]interface{}{
					"Key":    "dashboardId",
					"Values": []string{dashboardID},
				})
			}
			if dashboardName != "" {
				filters = append(filters, map[string]interface{}{
					"Key":    "dashboardName",
					"Values": []string{dashboardName},
				})
			}
			if dashboardRegion != "" {
				filters = append(filters, map[string]interface{}{
					"Key":    "dashboardRegion",
					"Values": []string{dashboardRegion},
				})
			}
			if tagKey != "" {
				filters = append(filters, map[string]interface{}{
					"Key":    "tagKey",
					"Values": []string{tagKey},
				})
			}
			if folderID != "" {
				filters = append(filters, map[string]interface{}{
					"Key":    "folderId",
					"Values": []string{folderID},
				})
			}
			if len(filters) > 0 {
				params["Filters"] = filters
			}

		result, err := clsClient.Call("DescribeDashboards", params)
		if err != nil {
			return err
		}

		// dashboardName 是模糊匹配，客户端做精确过滤
		if dashboardName != "" {
			if resp, ok := result["Response"].(map[string]interface{}); ok {
				if infos, ok := resp["DashboardInfos"].([]interface{}); ok {
					filtered := make([]interface{}, 0)
					for _, item := range infos {
						if m, ok := item.(map[string]interface{}); ok {
							if m["DashboardName"] == dashboardName {
								filtered = append(filtered, item)
							}
						}
					}
					resp["DashboardInfos"] = filtered
				}
			}
		}

		formatDashboardListResult(result, f.Format)
		return nil
		},
	}

	cmd.Flags().StringVar(&dashboardID, "id", "", "按仪表盘 ID 过滤")
	cmd.Flags().StringVar(&dashboardName, "name", "", "按仪表盘名称模糊搜索")
	cmd.Flags().StringVar(&dashboardRegion, "dashboard-region", "", "按仪表盘地域过滤")
	cmd.Flags().StringVar(&folderID, "folder-id", "", "按文件夹 ID 过滤")
	cmd.Flags().StringVar(&tagKey, "tag-key", "", "按标签键过滤")
	cmd.Flags().Int64Var(&offset, "offset", 0, "分页偏移量")
	cmd.Flags().Int64Var(&limit, "limit", 20, "返回条数（最大100）")
	return cmd
}

func newCreateCmd(f *cmdutil.Factory) *cobra.Command {
	var (
		dashboardName string
		data          string
		fromFile      string
		folderID      string
		tags          []string
	)

	cmd := &cobra.Command{
		Use:   "+create",
		Short: "创建仪表盘",
		Long: `创建一个新的仪表盘。

示例:
  cls-cli dashboard +create --name "我的仪表盘"
  cls-cli dashboard +create --name "业务大盘" --data '{}'
  cls-cli dashboard +create --name "业务大盘" --from-file ./dashboard.json
  cls-cli dashboard +create --name "业务大盘" --folder-id folder-xxxx
  cls-cli dashboard +create --name "业务大盘" --tag "env=prod" --tag "team=backend"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if dashboardName == "" {
				return fmt.Errorf("--name 参数必填")
			}

			if f.DryRun {
				fmt.Fprintf(f.IOStreams.Out, "DRY RUN:\n  Action: CreateDashboard\n  DashboardName: %s\n", dashboardName)
				return nil
			}

			clsClient, err := f.CLSClient()
			if err != nil {
				return err
			}

			// 从文件导入 Data 字段
			if fromFile != "" {
				fileBytes, err := os.ReadFile(fromFile)
				if err != nil {
					return fmt.Errorf("读取文件失败: %w", err)
				}
				var fileJSON map[string]interface{}
				if err := json.Unmarshal(fileBytes, &fileJSON); err != nil {
					return fmt.Errorf("解析 JSON 文件失败: %w", err)
				}
			// 支持三种格式：
			// 1. DescribeDashboards 完整响应（Response.DashboardInfos[0].Data 是字符串）
			// 2. 旧导出格式（Data 是字符串）
			// 3. 新导出格式（Data 是对象，需序列化回字符串）
			if resp, ok := fileJSON["Response"].(map[string]interface{}); ok {
				if infos, ok := resp["DashboardInfos"].([]interface{}); ok && len(infos) > 0 {
					if info, ok := infos[0].(map[string]interface{}); ok {
						if d, ok := info["Data"].(string); ok {
							data = d
						}
					}
				}
			} else if d, ok := fileJSON["Data"].(string); ok {
				// Data 是字符串（旧格式）
				data = d
			} else if d, ok := fileJSON["Data"]; ok && d != nil {
				// Data 是对象（新导出格式），序列化回字符串
				dataBytes, err := json.Marshal(d)
				if err != nil {
					return fmt.Errorf("序列化 Data 字段失败: %w", err)
				}
				data = string(dataBytes)
			}
				if data == "" {
					return fmt.Errorf("文件中未找到 Data 字段")
				}
			}

			params := map[string]interface{}{
				"DashboardName": dashboardName,
			}
			if data != "" {
				params["Data"] = data
			}
			if folderID != "" {
				params["FolderId"] = folderID
			}
			if len(tags) > 0 {
				tagList := make([]map[string]string, 0, len(tags))
				for _, t := range tags {
					// 解析 key=value 格式
					for i, c := range t {
						if c == '=' {
							tagList = append(tagList, map[string]string{
								"Key":   t[:i],
								"Value": t[i+1:],
							})
							break
						}
					}
				}
				if len(tagList) > 0 {
					params["Tags"] = tagList
				}
			}

			result, err := clsClient.Call("CreateDashboard", params)
			if err != nil {
				return err
			}

			output.FormatOutput(result, f.Format)
			output.PrintSuccess("仪表盘创建成功")
			return nil
		},
	}

	cmd.Flags().StringVar(&dashboardName, "name", "", "仪表盘名称（必填）")
	cmd.Flags().StringVar(&data, "data", "", "仪表盘配置数据（JSON 字符串）")
	cmd.Flags().StringVar(&fromFile, "from-file", "", "从本地 JSON 文件导入仪表盘配置（支持 DescribeDashboards 响应格式）")
	cmd.Flags().StringVar(&folderID, "folder-id", "", "文件夹 ID，指定仪表盘所在文件夹")
	cmd.Flags().StringArrayVar(&tags, "tag", nil, "标签，格式 key=value，可多次使用")
	return cmd
}

func newDeleteCmd(f *cmdutil.Factory) *cobra.Command {
	var dashboardID string

	cmd := &cobra.Command{
		Use:   "+delete",
		Short: "删除仪表盘",
		Long: `删除指定的仪表盘。此操作不可逆！

示例:
  cls-cli dashboard +delete --id dashboard-xxxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if dashboardID == "" {
				return fmt.Errorf("--id 参数必填")
			}

			if f.DryRun {
				fmt.Fprintf(f.IOStreams.Out, "DRY RUN:\n  Action: DeleteDashboard\n  DashboardId: %s\n", dashboardID)
				return nil
			}

			if !f.ConfirmAction(fmt.Sprintf("即将删除仪表盘 %s，此操作不可逆！", dashboardID)) {
				fmt.Fprintln(f.IOStreams.ErrOut, "已取消删除操作")
				return nil
			}

			clsClient, err := f.CLSClient()
			if err != nil {
				return err
			}

			_, err = clsClient.Call("DeleteDashboard", map[string]interface{}{
				"DashboardId": dashboardID,
			})
			if err != nil {
				return err
			}

			output.PrintSuccess(fmt.Sprintf("仪表盘 %s 已删除", dashboardID))
			return nil
		},
	}

	cmd.Flags().StringVar(&dashboardID, "id", "", "仪表盘 ID（必填）")
	return cmd
}

func newFolderListCmd(f *cmdutil.Factory) *cobra.Command {
	var (
		folderID   string
		folderName string
		limit      int64
	)

	cmd := &cobra.Command{
		Use:   "+folders",
		Short: "列出仪表盘文件夹",
		Long: `列出当前账号下的仪表盘文件夹列表。

示例:
  cls-cli dashboard +folders
  cls-cli dashboard +folders --id folder-xxxx
  cls-cli dashboard +folders --name "我的文件夹"
  cls-cli dashboard +folders --format table`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if f.DryRun {
				fmt.Fprintf(f.IOStreams.Out, "DRY RUN:\n  Action: DescribeFolders\n")
				return nil
			}

			clsClient, err := f.CLSClient()
			if err != nil {
				return err
			}

			params := map[string]interface{}{
				"Limit": limit,
			}

			var filters []map[string]interface{}
			if folderID != "" {
				filters = append(filters, map[string]interface{}{
					"Key":    "folderId",
					"Values": []string{folderID},
				})
			}
			if folderName != "" {
				filters = append(filters, map[string]interface{}{
					"Key":    "folderName",
					"Values": []string{folderName},
				})
			}
			if len(filters) > 0 {
				params["Filters"] = filters
			}

			result, err := clsClient.Call("DescribeFolders", params)
			if err != nil {
				return err
			}

			formatFolderListResult(result, f.Format)
			return nil
		},
	}

	cmd.Flags().StringVar(&folderID, "id", "", "按文件夹 ID 过滤")
	cmd.Flags().StringVar(&folderName, "name", "", "按文件夹名称过滤")
	cmd.Flags().Int64Var(&limit, "limit", 1000, "返回条数")
	return cmd
}

func formatFolderListResult(data interface{}, format output.Format) {
	if format == output.FormatTable || format == output.FormatCSV {
		dataMap, ok := data.(map[string]interface{})
		if !ok {
			output.FormatOutput(data, format)
			return
		}
		resp, ok := dataMap["Response"].(map[string]interface{})
		if !ok {
			output.FormatOutput(data, format)
			return
		}
		if folders, ok := resp["Data"].([]interface{}); ok {
			rows := make([]interface{}, 0, len(folders))
			for _, f := range folders {
				if m, ok := f.(map[string]interface{}); ok {
					rows = append(rows, map[string]interface{}{
						"Id":   m["Id"],
						"Name": m["Name"],
					})
				}
			}
			output.FormatOutput(rows, format)
			return
		}
	}
	output.FormatOutput(data, format)
}

func formatDashboardListResult(data interface{}, format output.Format) {
	if format == output.FormatTable || format == output.FormatCSV {
		dataMap, ok := data.(map[string]interface{})
		if !ok {
			output.FormatOutput(data, format)
			return
		}
		resp, ok := dataMap["Response"].(map[string]interface{})
		if !ok {
			output.FormatOutput(data, format)
			return
		}
		if dashboards, ok := resp["DashboardInfos"].([]interface{}); ok {
			rows := make([]interface{}, 0, len(dashboards))
			for _, d := range dashboards {
				if m, ok := d.(map[string]interface{}); ok {
					rows = append(rows, map[string]interface{}{
						"DashboardId":   m["DashboardId"],
						"DashboardName": m["DashboardName"],
						"CreateTime":    m["CreateTime"],
						"UpdateTime":    m["UpdateTime"],
					})
				}
			}
			output.FormatOutput(rows, format)
			return
		}
	}
	output.FormatOutput(data, format)
}
