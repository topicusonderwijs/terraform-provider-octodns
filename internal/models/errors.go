package models

import "errors"

var (
	ErrSubdomainNotFound      = errors.New("subdomain not found in zone")
	ErrSubdomainAlreadyExists = errors.New("subdomain already exists in zone")

	ErrTypeNotFound      = errors.New("record type not found in subdomain")
	ErrTypeAlreadyExists = errors.New("record type already exists in subdomain")

	ErrValidateNotAnIP                 = errors.New("value not validated as an IP")
	ErrValidateNotAnIPV4               = errors.New("value not validated as an IPV4")
	ErrValidateNotAnIPV6               = errors.New("value not validated as an IPV6")
	ErrValidateIPNotAllowed            = errors.New("an IP value is not allowed")
	ErrValidateFQDNRequiredTrailingDot = errors.New("fqdn value must end with a dot")
	ErrValidateFQDNForbidTrailingDot   = errors.New("fqdn value may not end with a dot")
	ErrValidateNotAFQDN                = errors.New("value not validated as a fqdn")
)
