package models

import "fmt"

type Scope struct {
	Name   string
	Path   string
	Branch string
	Ext    string
}

func (s *Scope) CreateFilePath(zone string) string {
	return fmt.Sprintf("%s/%s.%s", s.Path, zone, s.Ext)
}
func (s *Scope) GetBranch(fallback string) string {
	if s.Branch != "" {
		return s.Branch
	} else {
		return fallback
	}
}
