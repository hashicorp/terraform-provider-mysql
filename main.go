package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/terraform-providers/terraform-provider-mysql/mysql"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: mysql.Provider})
}
