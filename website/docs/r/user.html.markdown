---
layout: "mysql"
page_title: "MySQL: mysql_user"
sidebar_current: "docs-mysql-resource-user"
description: |-
  Creates and manages a user on a MySQL server.
---

# mysql\_user

The ``mysql_user`` resource creates and manages a user on a MySQL
server.

~> **Note:** The password for the user is provided in plain text, and is
obscured by an unsalted hash in the state
[Read more about sensitive data in state](/docs/state/sensitive-data.html).
Care is required when using this resource, to avoid disclosing the password.

## Example Usage

```hcl
resource "mysql_user" "jdoe" {
  user               = "jdoe"
  host               = "example.com"
  plaintext_password = "password"
}
```

## Example Usage with an Authentication Plugin

```hcl
resource "mysql_user" "nologin" {
  user               = "nologin"
  host               = "example.com"
  auth_plugin        = "mysql_no_login"
}
```

## Argument Reference

The following arguments are supported:

* `user` - (Required) The name of the user.
* `host` - (Optional) The source host of the user. Defaults to "localhost".
* `plaintext_password` - (Optional) The password for the user. This must be provided in plain text, so the data source for it must be secured. An _unsalted_ hash of the provided password is stored in state. Conflicts with `auth_plugin`.
* `password` - (Optional) Deprecated alias of `plaintext_password`, whose value is *stored as plaintext in state*. Prefer to use `plaintext_password` instead, which stores the password as an unsalted hash. Conflicts with `auth_plugin`.
* `auth_plugin` - (Optional) Use an [authentication plugin][ref-auth-plugins] to authenticate the user instead of using password authentication.  Description of the fields allowed in the block below. Conflicts with `password` and `plaintext_password`.  
* `tls_option` - (Optional) An TLS-Option for the `CREATE USER` or `ALTER USER` statement. The value is suffixed to `REQUIRE`. A value of 'SSL' will generate a `CREATE USER ... REQUIRE SSL` statement. See the [MYSQL `CREATE USER` documentation](https://dev.mysql.com/doc/refman/5.7/en/create-user.html) for more. Ignored if MySQL version is under 5.7.0.

[ref-auth-plugins]: https://dev.mysql.com/doc/refman/5.7/en/authentication-plugins.html

The `auth_plugin` value supports:

* `AWSAuthenticationPlugin` - Allows the use of IAM authentication with [Amazon
  Aurora][ref-amazon-aurora]. For more details on how to use IAM auth with
  Aurora, see [here][ref-aurora-using-iam].

[ref-amazon-aurora]: https://aws.amazon.com/rds/aurora/
[ref-aurora-using-iam]: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/UsingWithRDS.IAMDBAuth.html#UsingWithRDS.IAMDBAuth.Creating

* `mysql_no_login` - Uses the MySQL No-Login Authentication Plugin. The
  No-Login Authentication Plugin must be active in MySQL. For more information,
  see [here][ref-mysql-no-login].

[ref-mysql-no-login]: https://dev.mysql.com/doc/refman/5.7/en/no-login-pluggable-authentication.html

## Attributes Reference

The following attributes are exported:

* `user` - The name of the user.
* `password` - The password of the user.
* `id` - The id of the user created, composed as "username@host".
* `host` - The host where the user was created.

## Attributes Reference

No further attributes are exported.
