package models

import (
	"fmt"
	"testing"
)

func TestScope_CreateFilePath(t *testing.T) {

	testCases := []struct {
		want string
		path string
		ext  string
		zone string
	}{
		{
			want: "zones/example.com.yaml",
			path: "/zones",
			ext:  "yaml",
			zone: "example.com",
		},
		{
			want: "example.com.yml",
			path: "",
			ext:  "yml",
			zone: "example.com",
		},
		{
			want: "example.com.yml",
			path: "/",
			ext:  "yml",
			zone: "example.com",
		},
		{
			want: "zones/example.com.yaml",
			path: "/zones/",
			ext:  "yaml",
			zone: "example.com",
		},
		{
			want: "zones/example.com.yaml",
			path: " /zones/ ",
			ext:  " yaml ",
			zone: "example.com",
		},
	}

	for i, test := range testCases {

		s := NewScope(
			fmt.Sprintf("CreateFilePath_%d", i),
			test.path,
			"main",
			test.ext,
		)

		got := s.CreateFilePath(test.zone)

		if got != test.want {
			t.Errorf("%s: Path dont match want %q got %q", s.Name, test.want, got)
		}
	}

}

func TestScope_GetBranch(t *testing.T) {

	testCases := []struct {
		want     string
		branch   string
		fallback string
	}{
		{
			want:     "main",
			branch:   "main",
			fallback: "",
		},
		{
			want:     "main",
			branch:   "",
			fallback: "main",
		},
		{
			want:     "main",
			branch:   "  main  ",
			fallback: "",
		},
		{
			want:     "  main  ", // Fallback should not be trimmed
			branch:   "   ",
			fallback: "  main  ",
		},
	}

	for i, test := range testCases {

		s := NewScope(
			fmt.Sprintf("GetBranch_%d", i),
			"/",
			test.branch,
			".yaml",
		)

		got := s.GetBranch(test.fallback)

		if got != test.want {
			t.Errorf("%s: Path dont match want %q got %q", s.Name, test.want, got)
		}
	}
}
