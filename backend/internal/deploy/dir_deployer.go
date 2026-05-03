package deploy

import (
	"context"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// DirProvider 将 output 目录的内容复制到用户指定的本地目录
// 支持两种路径模式：
//   - basePath 模式：设置了 basePath（如 /GrideaPro/），将 href="/xxx" 替换为 href="/GrideaPro/xxx"（绝对路径，适用于 CDN 子目录）
//   - 相对路径模式：未设置 basePath，按文件深度将 href="/xxx" 替换为 href="../../xxx"（适用于 file:/// 协议）
type DirProvider struct{}

func NewDirProvider() *DirProvider {
	return &DirProvider{}
}

func (p *DirProvider) Deploy(ctx context.Context, outputDir string, setting *domain.Setting, logger LogFunc) error {
	targetDir := setting.OutputDir()
	if targetDir == "" {
		return fmt.Errorf("输出目录未配置")
	}

	// 校验目标目录
	info, err := os.Stat(targetDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("目标目录不存在: %s", targetDir)
		}
		return fmt.Errorf("无法访问目标目录: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("目标路径不是目录: %s", targetDir)
	}

	// 检查 context 取消
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	logger(fmt.Sprintf("正在复制文件到: %s", targetDir))

	// 递归复制
	count, err := copyDir(outputDir, targetDir)
	if err != nil {
		return fmt.Errorf("复制文件失败: %w", err)
	}

	// 重写 HTML 文件中的路径
	basePath := setting.BasePath()
	var htmlCount int
	if basePath != "" {
		// basePath 模式：将 href="/xxx" 替换为 href="/basePath/xxx"（绝对路径）
		logger(fmt.Sprintf("正在重写路径为基础路径模式: %s", basePath))
		htmlCount, err = rewriteHtmlPathsBasePath(targetDir, basePath)
	} else {
		// 相对路径模式：按深度计算 ../（file:/// 兼容）
		logger("正在重写路径以支持 file:// 协议...")
		htmlCount, err = rewriteHtmlPaths(targetDir)
	}
	if err != nil {
		return fmt.Errorf("路径重写失败: %w", err)
	}

	logger(fmt.Sprintf("部署完成，共复制 %d 个文件，重写 %d 个 HTML 文件", count, htmlCount))
	return nil
}

// rewriteHtmlPathsBasePath 遍历目标目录中的所有 HTML 文件，
// 将 href="/xxx" 和 src="/xxx" 替换为 href="/basePath/xxx"（绝对路径前缀模式，适用于 CDN 子目录）
func rewriteHtmlPathsBasePath(rootDir, basePath string) (int, error) {
	// 规范化 basePath：确保以 / 开头且以 / 结尾
	basePath = strings.TrimPrefix(basePath, "/")
	basePath = strings.TrimSuffix(basePath, "/")
	prefix := "/" + basePath + "/"

	count := 0
	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(strings.ToLower(d.Name()), ".html") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("读取 %s 失败: %w", path, err)
		}

		content := string(data)
		rewritten := rewritePaths(content, prefix)

		if rewritten != content {
			if err := os.WriteFile(path, []byte(rewritten), 0644); err != nil {
				return fmt.Errorf("写入 %s 失败: %w", path, err)
			}
			count++
		}
		return nil
	})
	return count, err
}

// rewriteHtmlPaths 遍历目标目录中的所有 HTML 文件，
// 将 href="/..." 和 src="/..." 替换为按文件深度计算的相对路径
func rewriteHtmlPaths(rootDir string) (int, error) {
	count := 0
	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(strings.ToLower(d.Name()), ".html") {
			return nil
		}

		// 计算从 rootDir 到当前文件的相对目录深度
		rel, err := filepath.Rel(rootDir, filepath.Dir(path))
		if err != nil {
			return err
		}

		depth := 0
		if rel != "." {
			depth = len(strings.Split(rel, string(filepath.Separator)))
		}

		// depth=0: prefix="./", depth=2: prefix="../../"
		var prefix string
		if depth == 0 {
			prefix = "./"
		} else {
			prefix = strings.Repeat("../", depth)
		}

		// 读取、替换、写回
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("读取 %s 失败: %w", path, err)
		}

		content := string(data)
		rewritten := rewritePaths(content, prefix)

		if rewritten != content {
			if err := os.WriteFile(path, []byte(rewritten), 0644); err != nil {
				return fmt.Errorf("写入 %s 失败: %w", path, err)
			}
			count++
		}
		return nil
	})
	return count, err
}

// rewritePaths 将 HTML 中的绝对路径替换为基于 prefix 的相对路径
func rewritePaths(html, prefix string) string {
	// 替换 href="/ 和 src="/（双引号），不可能匹配到 https:// 等外部 URL
	replacements := []struct{ old, new string }{
		{`href="/`, `href="` + prefix},
		{`src="/`, `src="` + prefix},
	}

	for _, r := range replacements {
		html = strings.ReplaceAll(html, r.old, r.new)
	}

	return html
}

// copyDir 递归复制 src 目录内容到 dst（覆盖已有文件）
func copyDir(src, dst string) (int, error) {
	count := 0
	err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(dst, rel)

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		// 复制文件
		if err := copyFile(path, targetPath); err != nil {
			return fmt.Errorf("复制 %s 失败: %w", rel, err)
		}
		count++
		return nil
	})
	return count, err
}

// copyFile 复制单个文件
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	// 复制权限
	info, err := srcFile.Stat()
	if err == nil {
		_ = dstFile.Chmod(info.Mode())
	}
	return nil
}
