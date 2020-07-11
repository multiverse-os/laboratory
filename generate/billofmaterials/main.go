package main

import (
	"github.com/bom-maker/cli"
)

const (
	// AppName is the application's name
	AppName = "bom-maker"
	// AppDesc is the application's description
	AppDesc = "Handle Billing-Of-Materials ready for parts' order"
)

var (
	// AppVersion is the version of the app (defined at compile time)
	AppVersion string
)

func main() {
	if AppVersion == "" {
		AppVersion = "master"
	}

	cli.Process(AppName, AppDesc, AppVersion)
}
