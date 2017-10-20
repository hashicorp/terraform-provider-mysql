package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/krogon-dp/terraform-provider-mysql/mysql"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: mysql.Provider})
}
