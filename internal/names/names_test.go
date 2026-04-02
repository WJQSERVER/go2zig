package names

import "testing"

func TestExported(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"rename_user": "RenameUser",
		"user_id":     "UserID",
		"ok":          "OK",
		"go_api":      "GoAPI",
		"http_server": "HTTPServer",
	}

	for input, want := range tests {
		if got := Exported(input); got != want {
			t.Fatalf("Exported(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestLowerCamel(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"rename_user": "renameUser",
		"user_id":     "userID",
		"ok":          "ok",
		"http_server": "httpServer",
	}

	for input, want := range tests {
		if got := LowerCamel(input); got != want {
			t.Fatalf("LowerCamel(%q) = %q, want %q", input, got, want)
		}
	}
}
