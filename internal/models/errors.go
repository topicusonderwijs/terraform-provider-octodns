package models

import "errors"

var (
	SubdomainNotFoundError      = errors.New("subdomain not found in zone")
	SubdomainAlreadyExistsError = errors.New("subdomain already exists in zone")

	TypeNotFoundError      = errors.New("record type not found in subdomain")
	TypeAlreadyExistsError = errors.New("record type already exists in subdomain")

	ValidateNotAnIP                 = errors.New("value not validated as an IP")
	ValidateNotAnIPV4               = errors.New("value not validated as an IPV4")
	ValidateNotAnIPV6               = errors.New("value not validated as an IPV6")
	ValidateIPNotAllowed            = errors.New("an IP value is not allowed")
	ValidateFQDNRequiredTrailingDot = errors.New("fqdn value must end with a dot")
	ValidateFQDNForbidTrailingDot   = errors.New("fqdn value may not end with a dot")
	ValidateNotAFQDN                = errors.New("value not validated as a fqdn")
)
