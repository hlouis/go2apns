// Package writer provides methods to write notifications to APNS
// and keep the http/2 connections with APNS
package writer

const DEVELOPMENT_ENV = "development"
const PRODUCTION_ENV = "production"

var hosts = map[string]string{
	DEVELOPMENT_ENV: "api.development.push.apple.com:443",
	PRODUCTION_ENV:  "api.push.apple.com:443",
}

type Writer struct {
	// ApnsEnv use to specify which APNS environment to use
	ApnsEnv string
}
