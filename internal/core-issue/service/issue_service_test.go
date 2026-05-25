package service

import (
	"context"
	"errors"
	"mime/multipart"
	"net/textproto"
	"strings"
	"testing"
)

// makeHeader 构造一个仅含 Filename / Size 的 multipart.FileHeader, 用于入口校验测试
func makeHeader(name string, size int64) *multipart.FileHeader {
	return &multipart.FileHeader{
		Filename: name,
		Header:   textproto.MIMEHeader{},
		Size:     size,
	}
}

// TestCreateIssueWithAttachments_RejectsTooLarge 验证文件大小超限时立即返回 ErrFileTooLarge
// 入口校验应在 attachmentService nil 检查之前
func TestCreateIssueWithAttachments_RejectsTooLarge(t *testing.T) {
	files := []*multipart.FileHeader{
		makeHeader("a.png", MaxAttachmentSize+1),
	}
	s := &issueService{} // attachmentService 为 nil, 但不应被调用
	_, err := s.CreateIssueWithAttachments(context.Background(), nil, files, 1)
	if err == nil {
		t.Fatal("期望返回错误, 实际为 nil")
	}
	if !errors.Is(err, ErrFileTooLarge) {
		t.Fatalf("期望 ErrFileTooLarge, 实际: %v", err)
	}
}

// TestCreateIssueWithAttachments_RejectsBadExt 验证不允许扩展名时返回 ErrInvalidFileType
func TestCreateIssueWithAttachments_RejectsBadExt(t *testing.T) {
	files := []*multipart.FileHeader{
		makeHeader("evil.exe", 100),
	}
	s := &issueService{}
	_, err := s.CreateIssueWithAttachments(context.Background(), nil, files, 1)
	if err == nil {
		t.Fatal("期望返回错误, 实际为 nil")
	}
	if !errors.Is(err, ErrInvalidFileType) {
		t.Fatalf("期望 ErrInvalidFileType, 实际: %v", err)
	}
}

// TestCreateIssueWithAttachments_RejectsWhenAttachmentSvcNil 验证有合法文件且 attachmentService 未注入时返回错误
// 即: 文件校验通过后, 才检查 attachmentService nil
func TestCreateIssueWithAttachments_RejectsWhenAttachmentSvcNil(t *testing.T) {
	files := []*multipart.FileHeader{makeHeader("a.png", 10)}
	s := &issueService{}
	_, err := s.CreateIssueWithAttachments(context.Background(), nil, files, 1)
	if err == nil {
		t.Fatal("期望返回错误, 实际为 nil")
	}
	if !strings.Contains(err.Error(), "附件服务") {
		t.Fatalf("期望错误信息含 '附件服务', 实际: %v", err)
	}
}
