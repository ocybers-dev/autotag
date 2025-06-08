package scheme

import (
	_type "autotag/type"
	"fmt"
	"github.com/hashicorp/go-version"
	"go.uber.org/zap"
	"os"
	"regexp"
	"strings"
)

// DefaultStrategy 默认策略实现
type DefaultStrategy struct{}

func (s *DefaultStrategy) Name() string {
	return "default"
}

// GenerateTags 根据提交列表生成标签信息
func (s *DefaultStrategy) GenerateTags(commitList []_type.CommitInfo) ([]_type.TagInfo, error) {

	var bumperStack []int
	var lastTaggedCommit *_type.CommitInfo
	var lastVersion string

	for i, commit := range commitList {
		if len(commit.Tags) > 0 {
			// 如果当前提交有标签，记录最后一个标签的提交信息
			lastTaggedCommit = &commitList[i]
			lastVersion = commit.Tags[0]
			break
		}
		bumperType := s.analyzeBumpType(commit)
		bumperStack = append(bumperStack, bumperType)
	}
	if lastTaggedCommit == nil {
		zap.L().Warn("没有找到带标签的提交，程序退出。")
		os.Exit(0)
	}
	tv, err := version.NewVersion(lastVersion)
	if err != nil {
		return nil, fmt.Errorf("创建当前版本号失败: %w", err)
	}
	zap.L().Info("成功获取当前版本号：", zap.String("version", tv.String()))
	for i := len(bumperStack) - 1; i >= 0; i-- {
		bumper := bumperStack[i]
		switch bumper {
		case 0:
			tv, err = majorBumper.bump(tv)
		case 1:
			tv, err = minorBumper.bump(tv)
		case 2:
			tv, err = patchBumper.bump(tv)
		default:
			return nil, fmt.Errorf("未知的版本升级类型: %d", bumper)
		}
		if err != nil {
			return nil, fmt.Errorf("版本号升级失败: %w", err)
		}
	}

	zap.L().Info("成功生成升级版本号：", zap.String("version", tv.String()))
	var tags []_type.TagInfo

	tags = append(tags, _type.TagInfo{
		Version:    "v" + tv.String(),
		CommitHash: commitList[0].Hash,
		Message:    "autotag default scheme:" + tv.String(),
	})

	return tags, nil
}

// analyzeBumpType 分析提交信息以确定版本升级类型
// 返回值: 0=major, 1=minor, 2=patch
func (s *DefaultStrategy) analyzeBumpType(commit _type.CommitInfo) int {
	// 获取提交消息，统一转换为小写便于匹配
	message := strings.ToLower(commit.Commit.Message)

	// 创建正则表达式
	majorRegex := regexp.MustCompile(`(?i)(\[major\]|#major)`)
	minorRegex := regexp.MustCompile(`(?i)(\[minor\]|#minor)`)
	patchRegex := regexp.MustCompile(`(?i)(\[patch\]|#patch)`)

	// 按优先级检查：major > minor > patch
	if majorRegex.MatchString(message) {
		return 0 // major bump
	}

	if minorRegex.MatchString(message) {
		return 1 // minor bump
	}

	if patchRegex.MatchString(message) {
		return 2 // patch bump
	}

	// 默认返回 patch bump
	return 2
}
