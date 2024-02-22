package models

import (
	"fmt"
	"strings"
)

type Scope struct {
	Name   string
	Path   string
	Branch string
	Ext    string
}

func NewScope(name, path, branch, ext string) Scope {

	// Trim off "/" characters from the front and end
	path = strings.Trim(path, "/ ")

	// Trim off "." characters from the front and end
	ext = strings.Trim(ext, ". ")

	branch = strings.TrimSpace(branch)

	return Scope{
		Name:   name,
		Path:   path,
		Branch: branch,
		Ext:    ext,
	}

}

func (s *Scope) CreateFilePath(zone string) string {
	if s.Path == "" {
		return fmt.Sprintf("%s.%s", zone, s.Ext)
	}
	return fmt.Sprintf("%s/%s.%s", s.Path, zone, s.Ext)

}

func (s *Scope) GetBranch(fallback string) string {
	if s.Branch != "" {
		return s.Branch
	} else {
		return fallback
	}
}
