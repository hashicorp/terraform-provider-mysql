## 0.1.1 (Unreleased)

ENHANCEMENTS:

* `mysql_user` now has a `plaintext_password` argument which stores its value in state as an _unsalted_ hash. This deprecates `password`, which stores its value in the state in cleartext. Since the hash is unsalted, some care is still warranted to secure the state, and a strong password should be used to reduce the chance of a successful brute-force attack on the hash. [GH-9]

BUG FIXES:

* Fix grant option SQL Statement [GH-12]

## 0.1.0 (June 21, 2017)

NOTES:

* Same functionality as that of Terraform 0.9.8. Repacked as part of [Provider Splitout](https://www.hashicorp.com/blog/upcoming-provider-changes-in-terraform-0-10/)
