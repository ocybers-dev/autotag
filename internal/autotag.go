package internal

import (
	"autotag/internal/git"
	"autotag/internal/scheme"
	_type "autotag/type"
	"fmt"
)

func Run() error {
	// 1. 根据配置直接创建策略实例
	var strategy scheme.Strategy

	switch _type.Opt.Scheme {
	case "default":
		strategy = &scheme.DefaultStrategy{}
	default:
		return fmt.Errorf("不支持的tag策略: [%s], 请检查'-s'参数再次尝试", _type.Opt.Scheme)
	}

	// 2. 初始化仓库，获取指定分支的提交以及提交对应的标签
	commitInfos, err := git.GetCommitAndTags()
	if err != nil {
		return err
	}

	// 4. 使用策略生成标签信息
	tagInfos, err := strategy.GenerateTags(commitInfos)
	if err != nil {
		return err
	}

	// 5. 根据生成的标签信息推送提交
	err = git.CreatePushTags(tagInfos)
	if err != nil {
		return err
	}

	return nil
}
