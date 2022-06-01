package common

var NAME = "prom-am-operator"
var SUMMARY = "0.1.0-main"
var BRANCH = "main"
var VERSION = "0.1.0"
var COMMIT = "dirty"

// AppVersion --
var AppVersion AppVersionInfo

// AppVersionInfo --
type AppVersionInfo struct {
	Name    string
	Version string
	Branch  string
	Summary string
	Commit  string
}

func init() {
	AppVersion = AppVersionInfo{
		Name:    NAME,
		Version: VERSION,
		Branch:  BRANCH,
		Summary: SUMMARY,
		Commit:  COMMIT,
	}
}
