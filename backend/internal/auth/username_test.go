package auth

import "testing"

func TestUsernameOK(t *testing.T) {
	cases := []struct {
		in string
		ok bool
	}{
		{"abc", true},
		{"A_b-9", true},
		{"a.very.long.name-32chars_________", false},
		{"1abc", false}, // 不能數字開頭
		{"ab", false},   // 太短
		{"中文", false},   // 非英數
		{"a b", false},  // 空白
		{"a/b", false},  // 非允許字元
	}

	for _, c := range cases {
		if usernameOK(c.in) != c.ok {
			t.Fatalf("usernameOK(%q) expected %v", c.in, c.ok)
		}
	}
}
