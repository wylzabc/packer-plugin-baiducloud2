package main

import (
	"fmt"
	"os"

	bccbuilder "github.com/hashicorp/packer-plugin-baiducloud/builder/bcc"
	"github.com/hashicorp/packer-plugin-baiducloud/version"
	"github.com/hashicorp/packer-plugin-sdk/plugin"
)

// main the function where execution of the program begins
func main() {
	pps := plugin.NewSet()
	pps.RegisterBuilder("bcc", new(bccbuilder.Builder))
	pps.SetVersion(version.PluginVersion)

	err := pps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
