// Package storage 提供文件存储服务
package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// LocalStorage 本地文件存储
type LocalStorage struct {
	basePath string
}

// NewLocalStorage 创建本地存储实例
func NewLocalStorage(basePath string) (*LocalStorage, error) {
	// 确保基础路径存在
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base path: %w", err)
	}

	return &LocalStorage{
		basePath: basePath,
	}, nil
}

// Save 保存文件
// 返回相对路径（相对于basePath）
func (s *LocalStorage) Save(file io.Reader, filename string) (string, error) {
	// 生成唯一的文件路径（使用时间戳避免冲突）
	relPath := filepath.Join("attachments", filename)
	fullPath := filepath.Join(s.basePath, relPath)

	// 确保目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// 创建文件
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	// 复制内容
	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("failed to copy file: %w", err)
	}

	return relPath, nil
}

// Delete 删除文件
func (s *LocalStorage) Delete(relPath string) error {
	fullPath := filepath.Join(s.basePath, relPath)
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// GetFullPath 获取文件的完整路径
func (s *LocalStorage) GetFullPath(relPath string) string {
	return filepath.Join(s.basePath, relPath)
}

// Exists 检查文件是否存在
func (s *LocalStorage) Exists(relPath string) bool {
	fullPath := filepath.Join(s.basePath, relPath)
	_, err := os.Stat(fullPath)
	return err == nil
}
