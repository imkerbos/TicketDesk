package service

import (
	"strings"
	"testing"
)

// TestGenerateToken 验证 token 生成的格式和长度
func TestGenerateToken(t *testing.T) {
	plain, hash, prefix, err := generateToken()
	if err != nil {
		t.Fatalf("generateToken err: %v", err)
	}
	if !strings.HasPrefix(plain, TokenPrefix) {
		t.Errorf("plain 应以 %s 开头, 实际 %s", TokenPrefix, plain)
	}
	if len(plain) != len(TokenPrefix)+TokenBodyLen {
		t.Errorf("plain 长度应为 %d, 实际 %d", len(TokenPrefix)+TokenBodyLen, len(plain))
	}
	if len(hash) != 64 {
		t.Errorf("hash 应为 64 hex 字符, 实际 %d", len(hash))
	}
	if prefix != plain[:16] {
		t.Errorf("prefix 不一致: prefix=%s, plain[:16]=%s", prefix, plain[:16])
	}
}

// TestHashTokenDeterministic 验证相同输入得到相同 hash，不同输入得到不同 hash
func TestHashTokenDeterministic(t *testing.T) {
	h1 := hashToken("td_pat_abc")
	h2 := hashToken("td_pat_abc")
	if h1 != h2 {
		t.Errorf("hash 应确定性：同输入应得相同 hash")
	}
	h3 := hashToken("td_pat_abd")
	if h1 == h3 {
		t.Errorf("不同 token 应有不同 hash")
	}
}

// TestIsAPIToken 验证 API token 格式判断
func TestIsAPIToken(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"td_pat_a8f9k2m3n4p5q6r7s8t9u0v1w2x3y4z5", true}, // 正好 39 字符
		{"td_pat_short", false},
		{"random_string", false},
		{"", false},
	}
	for _, c := range cases {
		if got := IsAPIToken(c.in); got != c.want {
			t.Errorf("IsAPIToken(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}
