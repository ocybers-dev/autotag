package git

import (
	"autotag/type"
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"os"
	"os/exec"
)

func GetCommitAndTags() ([]_type.CommitInfo, error) {
	// 打开仓库
	repo, err := git.PlainOpen(_type.Opt.RepoPath)
	if err != nil {
		return nil, fmt.Errorf("连接git仓库失败: %w", err)
	}

	// 获取指定分支的引用
	var branchRef *plumbing.Reference
	// 获取指定分支的引用
	branchRef, err = repo.Reference(plumbing.ReferenceName("refs/heads/"+_type.Opt.Branch), true)
	if err != nil {
		return nil, fmt.Errorf("获取分支失败 [branch: %s]: %w", _type.Opt.Branch, err)
	}

	// 获取提交迭代器，按时间倒序排列（最新的在前面）
	commitIter, err := repo.Log(&git.LogOptions{
		From:  branchRef.Hash(),
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return nil, fmt.Errorf("获取提交日志失败: %w", err)
	}
	defer commitIter.Close()

	// 获取所有标签引用
	tagRefs, err := repo.Tags()
	if err != nil {
		return nil, fmt.Errorf("获取标签失败: %w", err)
	}

	// 创建提交哈希到标签的映射
	commitToTags := make(map[string][]string)
	err = tagRefs.ForEach(func(tagRef *plumbing.Reference) error {
		// 获取标签对象
		tagObj, err := repo.TagObject(tagRef.Hash())
		var targetHash plumbing.Hash

		if err != nil {
			// 如果不是注释标签，直接使用引用的哈希
			targetHash = tagRef.Hash()
		} else {
			// 如果是注释标签，获取目标提交的哈希
			targetHash = tagObj.Target
		}

		// 将标签名添加到对应提交的标签列表中
		tagName := tagRef.Name().Short()
		commitToTags[targetHash.String()] = append(commitToTags[targetHash.String()], tagName)

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("处理标签时出错: %w", err)
	}

	// 收集提交和标签信息
	var result []_type.CommitInfo
	var tagCount int
	err = commitIter.ForEach(func(commit *object.Commit) error {
		commitHash := commit.Hash.String()

		// 获取该提交关联的标签
		tags := commitToTags[commitHash]
		if tags == nil {
			tags = []string{} // 确保不是nil切片
		}

		result = append(result, _type.CommitInfo{
			Commit: commit,
			Hash:   commitHash,
			Tags:   tags,
		})
		tagCount += len(tags)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("遍历提交时出错: %w", err)
	}

	fmt.Printf("成功获取提交和标签信息 [branch: %s, commits: %d, tags: %d]\n",
		_type.Opt.Branch, len(result), tagCount)

	if tagCount < 1 {
		return nil, errors.New("没有找到任何标签，请检查仓库是否有标签或提交信息是否符合规范")
	}

	return result, nil
}

func CreatePushTags(tagList []_type.TagInfo) error {
	// 打开仓库
	repo, err := git.PlainOpen(_type.Opt.RepoPath)
	if err != nil {
		return fmt.Errorf("连接git仓库失败: %w", err)
	}

	// 遍历所有需要创建的标签
	var createErrors []string
	for _, tagInfo := range tagList {
		// 创建标签
		hash := plumbing.NewHash(tagInfo.CommitHash)

		// 验证提交是否存在
		commit, err := repo.CommitObject(hash)
		if err != nil {
			errMsg := fmt.Sprintf("获取提交对象失败 [hash: %s, version: %s]: %v",
				tagInfo.CommitHash, tagInfo.Version, err)
			fmt.Println(errMsg)
			createErrors = append(createErrors, errMsg)
			continue
		}

		// 创建标签引用
		tagRef := plumbing.NewHashReference(
			plumbing.ReferenceName("refs/tags/"+tagInfo.Version),
			hash,
		)

		// 将标签写入仓库
		err = repo.Storer.SetReference(tagRef)
		if err != nil {
			errMsg := fmt.Sprintf("创建标签失败 [version: %s, hash: %s]: %v",
				tagInfo.Version, tagInfo.CommitHash, err)
			fmt.Println(errMsg)
			createErrors = append(createErrors, errMsg)
			continue
		}

		fmt.Printf("Tag创建成功\n version: %s, hash: %s, message: %s\n",
			tagInfo.Version, tagInfo.CommitHash, commit.Message)
	}

	// 如果有创建标签的错误，返回汇总错误
	if len(createErrors) > 0 {
		return fmt.Errorf("部分标签创建失败: %d个错误", len(createErrors))
	}

	// 推送标签到远程仓库
	if !_type.Opt.DryRun {
		// 使用 shell 命令推送所有标签
		cmd := exec.Command("git", "push", "--tags")

		// 设置工作目录
		cmd.Dir = _type.Opt.RepoPath

		// 设置环境变量（继承当前进程的环境变量，确保 Git 配置可用）
		cmd.Env = os.Environ()

		// 执行命令并获取输出
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		if err != nil {
			return fmt.Errorf("推送标签失败 [command: git push --tags, workdir: %s, output: %s]: %w",
				_type.Opt.RepoPath, outputStr, err)
		} else {
			fmt.Printf("Tag推送成功\n%s", outputStr)
		}
	}

	return nil
}
