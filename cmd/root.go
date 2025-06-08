package cmd

import (
	"autotag/internal"
	"autotag/type"
	"fmt"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"strings"
	"time"
)

var rootCmd = &cobra.Command{
	Use: " autotag",
	Long: `    ___   __  ____________  _________   ______
   /   | / / / /_  __/ __ \/_  __/   | / ____/
  / /| |/ / / / / / / / / / / / / /| |/ / __  
 / ___ / /_/ / / / / /_/ / / / / ___ / /_/ /  
/_/  |_\____/ /_/  \____/ /_/ /_/  |_\____/   
`,
	Run: func(cmd *cobra.Command, args []string) {
		// 初始化日志
		logger, _ := initLogger(_type.Opt.Verbose)
		defer logger.Sync()
		zap.ReplaceGlobals(logger)
		// 解析元数据
		p2MetaFlag(cmd)
		// 业务逻辑处理
		err := internal.Run()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolP("help", "h", false, "显示该命令的帮助信息")
	rootCmd.Flags().BoolVarP(&_type.Opt.Verbose, "verbose", "v", false, "启用详细日志输出")
	rootCmd.Flags().StringVarP(&_type.Opt.RepoPath, "internal", "r", ".", "指定要操作的 Git 仓库路径，默认是当前目录")
	rootCmd.Flags().StringVarP(&_type.Opt.Branch, "branch", "b", "main", "要分析的 Git 分支名称，默认为 main")
	rootCmd.Flags().StringVarP(&_type.Opt.Scheme, "scheme", "s", "default", "提交信息解析方案，支持 autotag 或 conventional")
	rootCmd.Flags().StringVarP(&_type.Opt.Prefix, "prefix", "p", "", "提交版本信息前缀，默认为空")
	rootCmd.Flags().StringSliceP("meta", "m", nil, "附加元数据，格式为 key=value，可以多次使用")
	rootCmd.Flags().BoolVarP(&_type.Opt.DryRun, "dry-run", "d", false, "启用干运行模式，只打印将要执行的操作，不实际提交")
}

// p2MetaFlag 解析(parse)命令行参数中的 meta 标志， 打印(print)命令行接收参数
func p2MetaFlag(cmd *cobra.Command) {
	// parse meta flag
	if _type.Opt.Meta == nil {
		_type.Opt.Meta = make(map[string]string)
	}

	metaSlice, err := cmd.Flags().GetStringSlice("meta")
	if err != nil {
		zap.L().Error("获取 meta 参数失败", zap.Error(err))
	}
	for _, item := range metaSlice {
		parts := strings.SplitN(item, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			val := parts[1]
			_type.Opt.Meta[key] = val
		} else {
			zap.L().Warn("meta 参数格式不正确", zap.String("item", item))
		}
	}

	// print received parameters - 极简风格
	const (
		Reset  = "\033[0m"
		Bold   = "\033[1m"
		Blue   = "\033[34m"
		Green  = "\033[32m"
		Yellow = "\033[33m"
		Cyan   = "\033[36m"
	)

	fmt.Printf("%s%s=== 运行参数 ===%s\n", Bold, Blue, Reset)
	fmt.Printf("  • RepoPath: %s%s%s\n", Green, _type.Opt.RepoPath, Reset)
	fmt.Printf("  • Branch:   %s%s%s\n", Green, _type.Opt.Branch, Reset)
	fmt.Printf("  • Scheme:   %s%s%s\n", Green, _type.Opt.Scheme, Reset)
	fmt.Printf("  • Prefix:   %s%s%s\n", Green, _type.Opt.Prefix, Reset)
	fmt.Printf("  • Verbose:  %s%v%s\n", Green, _type.Opt.Verbose, Reset)
	fmt.Printf("  • DryRun:   %s%v%s\n", Green, _type.Opt.DryRun, Reset)
	fmt.Printf("%s元数据:%s\n", Yellow, Reset)
	if len(_type.Opt.Meta) == 0 {
		fmt.Printf("  %s(无元数据)%s\n", Cyan, Reset)
	} else {
		for k, v := range _type.Opt.Meta {
			fmt.Printf("  • %s%s%s = %s%s%s\n", Cyan, k, Reset, Green, v, Reset)
		}
	}
}

func initLogger(verbose bool) (*zap.Logger, error) {
	var cfg zap.Config
	if verbose {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
	}

	cfg.Encoding = "console"
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoder(func(time.Time, zapcore.PrimitiveArrayEncoder) {})
	cfg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	return cfg.Build()
}
