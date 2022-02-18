package v1

import (
	"time"
)

// LicenseToken a license token parsed from an OpenFaaS Ltd JWT
type LicenseToken struct {
	Email    string
	Name     string
	Expires  time.Time
	Products []string
}
