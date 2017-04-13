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

## Argument Reference

The following arguments are supported:

* `user` - (Required) The name of the user.

* `host` - (Optional) The source host of the user. Defaults to "localhost".

* `plaintext_password` - (Optional) The password for the user. This must be
  provided in plain text, so the data source for it must be secured.
  An _unsalted_ hash of the provided password is stored in state.

* `password` - (Optional) Deprecated alias of `plaintext_password`, whose
  value is *stored as plaintext in state*. Prefer to use `plaintext_password`
  instead, which stores the password as an unsalted hash.

## Attributes Reference

No further attributes are exported.
