package _type

import "github.com/go-git/go-git/v5/plumbing/object"

type Options struct {
	// Verbose 启用详细日志输出（-v 或 --verbose）
	Verbose bool

	// DryRun 启用干运行模式（-d 或 --dry-run），只打印将要执行的操作，不实际提交
	DryRun bool

	// RepoPath 指定要操作的 Git 仓库路径（-r 或 --internal），默认是当前目录 "./"
	RepoPath string

	// Branch 要分析的 Git 分支名称（-b 或 --branch），默认为 main
	Branch string

	// Scheme 提交信息解析方案（-s 或 --scheme）
	Scheme string

	// Prefix 提交版本信息前缀（-p 或 --prefix），默认为空
	Prefix string

	// Meta 附加元数据(-m 或 --meta,例如：  -m "name=string,hobby=Great Sword")，默认为空
	Meta map[string]string
}

var Opt = &Options{}

type CommitInfo struct {
	Commit *object.Commit
	Hash   string
	Tags   []string
}

type TagInfo struct {
	Version    string // 版本号，如 "v1.2.3"
	CommitHash string // 对应的提交哈希
	Message    string // 标签消息（可选）
}
