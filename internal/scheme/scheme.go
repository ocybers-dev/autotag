package scheme

import (
	_type "autotag/type"
	"fmt"
	"github.com/hashicorp/go-version"
)

// Strategy 版本号生成策略接口
type Strategy interface {
	// GenerateTags 根据提交列表生成标签信息
	// 返回需要创建的标签列表
	GenerateTags(commitList []_type.CommitInfo) ([]_type.TagInfo, error)

	// Name 返回策略名称
	Name() string
}

// 版本号升级接口
type bumper interface {
	bump(*version.Version) (*version.Version, error)
}

type major struct{}
type minor struct{}
type patch struct{}

var (
	majorBumper major
	minorBumper minor
	patchBumper patch
)

func (m major) bump(cv *version.Version) (*version.Version, error) {
	segments := cv.Segments()

	vString := fmt.Sprintf("%d", segments[0]+1)
	for index := range segments {
		if index == 0 {
			continue
		}
		if index == 1 || index == 2 {
			vString += ".0"
		}
	}
	return version.NewVersion(vString)
}

func (m minor) bump(cv *version.Version) (*version.Version, error) {
	segments := cv.Segments()
	vString := fmt.Sprintf("%d", segments[0])
	if len(segments) >= 2 {
		vString += fmt.Sprintf(".%d", segments[1]+1)
	}
	if len(segments) >= 3 {
		for index, value := range segments {
			if index > 2 {
				vString += fmt.Sprintf(".%d", value)
			}
		}
	}
	return version.NewVersion(vString)
}

func (m patch) bump(cv *version.Version) (*version.Version, error) {
	segments := cv.Segments()
	vString := cv.String()
	if len(segments) >= 3 {
		vString = fmt.Sprintf("%d.%d.%d", segments[0], segments[1], segments[2]+1)
		for index, value := range segments {
			if index > 2 {
				vString += fmt.Sprintf(".%d", value)
			}
		}
	}
	return version.NewVersion(vString)
}
