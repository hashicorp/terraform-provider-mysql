package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/sfdc-pcg/terraform-provider-mysql/mysql"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: mysql.Provider})
}
