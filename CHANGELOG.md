## 1.5.0 (Unreleased)

BUG FIXES:

* Lazily connect to MySQL servers. ([#43](https://github.com/terraform-providers/terraform-provider-mysql/issues/43))
* Add retries to MySQL server connection logic. ([#43](https://github.com/terraform-providers/terraform-provider-mysql/issues/43))
* Migrated to Go modules for dependencies and `vendor/` management.

IMPROVEMENTS:

* Provider now has a `tls` option that configures TSL for server connections. ([#43](https://github.com/terraform-providers/terraform-provider-mysql/issues/43))
* `r/mysql_user`: Added the `tls_option` attribute, which allows to restrict the MySQL users to a specific MySQL-TLS-Encryption. ([#26](https://github.com/terraform-providers/terraform-provider-mysql/issues/40))
* `r/mysql_grant`: Added the `tls_option` attribute, which allows to restrict the MySQL grant to a specific MySQL-TLS-Encryption. ([#26](https://github.com/terraform-providers/terraform-provider-mysql/issues/40))
* `r/mysql_grant`: Added a `table` argument that allows `GRANT` statements to be scoped to a single table.

## 1.1.0 (March 28, 2018)

IMPROVEMENTS:

* `resource/user`: Added the `auth_plugin` attribute, which allows for the use of authentication plugins when creating MySQL users. ([#26](https://github.com/terraform-providers/terraform-provider-mysql/issues/26))

## 1.0.1 (January 03, 2018)

BUG FIXES:

* Supporting MySQL versions containing a `~` by updating `hashicorp/go-version` ([#27](https://github.com/terraform-providers/terraform-provider-mysql/issues/27))

## 1.0.0 (November 03, 2017)

UPGRADE NOTES:

* This provider is now using a different underlying library to access MySQL (See [#16](https://github.com/terraform-providers/terraform-provider-mysql/issues/16)). This should be a drop-in replacement for all of the functionality exposed by this provider, but just in case it is suggested to test cautiously after upgrading (review plans before applying, etc) in case of any edge-cases in interactions with specific versions of MySQL.

IMPROVEMENTS:

* `mysql_user` now has a `plaintext_password` argument which stores its value in state as an _unsalted_ hash. This deprecates `password`, which stores its value in the state in cleartext. Since the hash is unsalted, some care is still warranted to secure the state, and a strong password should be used to reduce the chance of a successful brute-force attack on the hash. ([#9](https://github.com/terraform-providers/terraform-provider-mysql/issues/9))

BUG FIXES:

* Fix grant option SQL Statement ([#12](https://github.com/terraform-providers/terraform-provider-mysql/issues/12))

## 0.1.0 (June 21, 2017)

NOTES:

* Same functionality as that of Terraform 0.9.8. Repacked as part of [Provider Splitout](https://www.hashicorp.com/blog/upcoming-provider-changes-in-terraform-0-10/)
