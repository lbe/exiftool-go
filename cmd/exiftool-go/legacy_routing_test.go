package main

import "testing"

func TestIsLegacyPipelineInvocation(t *testing.T) {
	t.Helper()

	cases := []struct {
		name string
		args []string
		want bool
	}{
		{"empty", nil, true},
		{"dir_only", []string{"/tmp"}, true},
		{"flags_and_dir", []string{"-l", "INFO", "/tmp"}, true},
		{"long_eq", []string{"--log-level=WARN", "--workers=2", "d"}, true},
		{"exiftool_ver", []string{"-ver"}, false},
		{"mixed_exiftool", []string{"-l", "INFO", "-json", "x.jpg"}, false},
		{"incomplete_l", []string{"-l"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()
			if got := isLegacyPipelineInvocation(tc.args); got != tc.want {
				t.Fatalf("isLegacyPipelineInvocation(%q) = %v, want %v", tc.args, got, tc.want)
			}
		})
	}
}
