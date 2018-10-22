---
layout: "mysql"
page_title: "MySQL: mysql_grant"
sidebar_current: "docs-mysql-resource-grant"
description: |-
  Creates and manages privileges given to a user on a MySQL server
---

# mysql\_grant

The ``mysql_grant`` resource creates and manages privileges given to
a user on a MySQL server.

## Example Usage

```hcl
resource "mysql_user" "jdoe" {
  user     = "jdoe"
  host     = "example.com"
  password = "password"
}

resource "mysql_grant" "jdoe" {
  user       = "${mysql_user.jdoe.user}"
  host       = "${mysql_user.jdoe.host}"
  database   = "app"
  privileges = ["SELECT", "UPDATE"]
}
```

## Argument Reference

The following arguments are supported:

* `user` - (Required) The name of the user.
* `host` - (Optional) The source host of the user. Defaults to "localhost".
* `database` - (Required) The database to grant privileges on.
* `table` - (Optional) Which table to grant `privileges` on. Defaults to `*`, which is all tables.
* `privileges` - (Required) A list of privileges to grant to the user. Refer to a list of privileges (such as [here](https://dev.mysql.com/doc/refman/5.5/en/grant.html)) for applicable privileges.
* `tls_option` - (Optional) An TLS-Option for the `GRANT` statement. The value is suffixed to `REQUIRE`. A value of 'SSL' will generate a `GRANT ... REQUIRE SSL` statement. See the [MYSQL `GRANT` documentation](https://dev.mysql.com/doc/refman/5.7/en/grant.html) for more. Ignored if MySQL version is under 5.7.0.
* `grant` - (Optional) Whether to also give the user privileges to grant the same privileges to other users.

## Attributes Reference

No further attributes are exported.
