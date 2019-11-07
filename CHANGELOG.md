## 1.9.1 (Unreleased)
## 1.9.0 (November 07, 2019)

FEATURES:
* New provider attribute `proxy` ([#102](https://github.com/terraform-providers/terraform-provider-mysql/pull/102))
* New provider attribute `authentication_plugin` ([#105](https://github.com/terraform-providers/terraform-provider-mysql/pull/105))

IMPROVEMENTS:
* Update documentation regarding environment variables for provider config vars ([#103](https://github.com/terraform-providers/terraform-provider-mysql/pull/103))
* Small documentation update around spacing ([#104](https://github.com/terraform-providers/terraform-provider-mysql/pull/104))

## 1.8.0 (October 02, 2019)

FEATURES:
* Add parameters for mysql connection configuration ([#95](https://github.com/terraform-providers/terraform-provider-mysql/pull/95))

IMPROVEMENTS:
* Remove use of config pkg ([#93](https://github.com/terraform-providers/terraform-provider-mysql/pull/93))
* Migrate provider to new standalone Terraform SDK ([#96](https://github.com/terraform-providers/terraform-provider-mysql/pull/96))

BUG FIXES:
* Disable REQUIRE syntax when `tls_options` is an empty string ([#91](https://github.com/terraform-providers/terraform-provider-mysql/pull/91))

## 1.7.0 (July 24, 2019)

FEATURES:
* Add compatibility to create databases on mariadb instances ([#83](https://github.com/terraform-providers/terraform-provider-mysql/pull/83))

IMPROVEMENTS:
* Replace `satori/go.uuid` with `gofrs/uuid` ([#69](https://github.com/terraform-providers/terraform-provider-mysql/pull/69))
* change pgp_key from optional to required ([#87](https://github.com/terraform-providers/terraform-provider-mysql/pull/87))

## 1.6.0 (June 18, 2019)

FEATURES:
* **Terraform 0.12** compatibility ([#85](https://github.com/terraform-providers/terraform-provider-mysql/pull/85))

IMPROVEMENTS:
* Documentation fix-ups ([#84](https://github.com/terraform-providers/terraform-provider-mysql/pull/84) & [#81](https://github.com/terraform-providers/terraform-provider-mysql/pull/81))

## 1.5.2 (May 29, 2019)

BUG FIXES:

* Regenerate go.sum with correct checksum & pin Go ([#70](https://github.com/terraform-providers/terraform-provider-mysql/issues/70))
* Configure provider before running test ([#72](https://github.com/terraform-providers/terraform-provider-mysql/issues/72))
* Handle revoke of grants properly ([#73](https://github.com/terraform-providers/terraform-provider-mysql/issues/73))

## 1.5.1 (January 24, 2019)

BUG FIXES:

* Fix grant error related to mysql syntax error ([#57](https://github.com/terraform-providers/terraform-provider-mysql/issues/57))

## 1.5.0 (November 07, 2018)

BUG FIXES:

* Lazily connect to MySQL servers. ([#43](https://github.com/terraform-providers/terraform-provider-mysql/issues/43))
* Add retries to MySQL server connection logic. ([#43](https://github.com/terraform-providers/terraform-provider-mysql/issues/43))
* Migrated to Go modules for dependencies and `vendor/` management. ([#44](https://github.com/terraform-providers/terraform-provider-mysql/issues/44))

IMPROVEMENTS:

* Provider now supports MySQL 8. ([#53](https://github.com/terraform-providers/terraform-provider-mysql/issues/53))
* Acceptance tests now ran against MySQL 5.6, 5.7, and 8.0.
* Provider now has a `tls` argument that configures TSL for server connections. ([#43](https://github.com/terraform-providers/terraform-provider-mysql/issues/43))
* `r/mysql_user`: Added the `tls_option` argument, which allows to restrict the MySQL users to a specific MySQL-TLS-Encryption. ([#26](https://github.com/terraform-providers/terraform-provider-mysql/issues/40))
* `r/mysql_grant`: Added the `tls_option` argument, which allows to restrict the MySQL grant to a specific MySQL-TLS-Encryption. ([#26](https://github.com/terraform-providers/terraform-provider-mysql/issues/40))
* `r/mysql_grant`: Added a `table` argument that allows `GRANT` statements to be scoped to a single table. ([#39](https://github.com/terraform-providers/terraform-provider-mysql/issues/30))
* `r/mysql_grant`: Added a `role` argument that allows `GRANT` assign privileges to roles. ([#53](https://github.com/terraform-providers/terraform-provider-mysql/issues/53))
* `r/mysql_grant`: Added a `roles` argument that allows `GRANT` assign roles to a user. ([#53](https://github.com/terraform-providers/terraform-provider-mysql/issues/53))
* `r/mysql_user_password`: Manages a PGP encrypted randomly assigned password for the given MySQL user. ([#50](https://github.com/terraform-providers/terraform-provider-mysql/issues/50))
* `r/mysql_role`: New resource for managing MySQL roles. ([#48](https://github.com/terraform-providers/terraform-provider-mysql/issues/48))

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
