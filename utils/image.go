package utils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrEmptyFilename     = errors.New("filename is empty")
	ErrUnallowedFilename = errors.New("unallowed file name")
	ErrUnallowedExt      = errors.New("only allowed to get image files")
	ErrUnallowedPath     = errors.New("unallowed path")
	ErrFileNotFound      = errors.New("file not found")
	ErrIsDirectory       = errors.New("directory not allowed")
	ErrDeleteFailed      = errors.New("failed to delete file")
)

// ValidateAndResolveImagePath 验证并解析图片文件路径
// 参数:
//   - filename: 用户提供的文件名（不含路径）
//   - baseDir: 允许的根目录（如 avatarDir）
//
// 返回:
//   - cleanPath: 清理后的绝对路径，可直接用于文件操作
//   - ext: 小写扩展名（如 ".png"），可用于设置 Content-Type
//   - error: 验证失败时的具体错误
func ValidateAndResolveImagePath(filename, baseDir string) (cleanPath string, ext string, err error) {
	if filename == "" {
		return "", "", ErrEmptyFilename
	}

	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return "", "", ErrUnallowedFilename
	}

	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
		".avif": true,
	}
	ext = strings.ToLower(filepath.Ext(filename))
	if !allowedExts[ext] {
		return "", "", ErrUnallowedExt
	}

	safePath := filepath.Join(baseDir, filename)
	cleanPath, err = filepath.Abs(safePath)
	if err != nil {
		return "", "", fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	absBaseDir, err := filepath.Abs(baseDir)
	if err != nil {
		return "", "", fmt.Errorf("failed to resolve base directory: %w", err)
	}

	if !strings.HasPrefix(cleanPath, absBaseDir+string(filepath.Separator)) && cleanPath != absBaseDir {
		return "", "", ErrUnallowedPath
	}

	fileInfo, err := os.Stat(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", "", ErrFileNotFound
		}
		return "", "", fmt.Errorf("failed to stat file: %w", err)
	}
	if fileInfo.IsDir() {
		return "", "", ErrIsDirectory
	}

	return cleanPath, ext, nil
}

// DeleteImageFile 安全删除指定图片文件
// 参数:
//   - filename: 用户提供的文件名（不含路径）
//   - baseDir: 允许的根目录（如 avatarDir）
//
// 返回:
//   - error: 删除失败时的具体错误
func DeleteImageFile(filename, baseDir string) error {
	cleanPath, _, err := ValidateAndResolveImagePath(filename, baseDir)
	if err != nil {
		return err
	}

	if err := os.Remove(cleanPath); err != nil {
		if os.IsNotExist(err) {
			return ErrFileNotFound
		}
		return fmt.Errorf("%w: %v", ErrDeleteFailed, err)
	}

	return nil
}
