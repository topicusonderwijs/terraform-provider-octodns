package models

import "errors"

var (
	SubdomainNotFoundError      = errors.New("subdomain not found in zone")
	SubdomainAlreadyExistsError = errors.New("subdomain already exists in zone")

	TypeNotFoundError      = errors.New("record type not found in subdomain")
	TypeAlreadyExistsError = errors.New("record type already exists in subdomain")
)
