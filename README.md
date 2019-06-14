[![Build Status](https://travis-ci.org/terraform-providers/terraform-provider-mysql.svg?branch=master)](https://travis-ci.org/terraform-providers/terraform-provider-mysql)

Terraform Provider
==================

- Website: https://www.terraform.io
- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
- Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)

<img src="https://cdn.rawgit.com/hashicorp/terraform-website/master/content/source/assets/images/logo-hashicorp.svg" width="600px">

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 0.12.x
-	[Go](https://golang.org/doc/install) 1.12 (to build the provider plugin)

Usage
-----

For Terraform 0.12+ compatibility, the configuration should specify version 1.6 or higher:

```hcl
provider "mysql" {
  version = "~> 1.6"
}
```

Building The Provider
---------------------

Clone repository to: `$GOPATH/src/github.com/terraform-providers/terraform-provider-mysql`

```sh
$ mkdir -p $GOPATH/src/github.com/terraform-providers; cd $GOPATH/src/github.com/terraform-providers
$ git clone git@github.com:terraform-providers/terraform-provider-mysql
```

Enter the provider directory and build the provider

```sh
$ cd $GOPATH/src/github.com/terraform-providers/terraform-provider-mysql
$ make build
```

Using the provider
----------------------
## Fill in for each provider

Developing the Provider
---------------------------

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.12+ is *required*). You'll also need to correctly setup a [GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding `$GOPATH/bin` to your `$PATH`.

To compile the provider, run `make build`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

```sh
$ make bin
...
$ $GOPATH/bin/terraform-provider-mysql
...
```

In order to test the provider, you can simply run `make test`.

```sh
$ make test
```

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```sh
$ make testacc
```

If you want to run the Acceptance tests on your own machine with a MySQL in Docker:

```bash
$ docker run --rm --name some-mysql -e MYSQL_ROOT_PASSWORD=my-secret-pw -d -p 3306:3306 mysql:5.7
$ # wait for a few seconds to let MySQL stand up, check the logs with: docker logs -f some-mysql
$ export MYSQL_USERNAME=root
$ export MYSQL_ENDPOINT=localhost:3306
$ export MYSQL_PASSWORD=my-secret-pw
$ mysql -h localhost -u root -p -e "INSTALL PLUGIN mysql_no_login SONAME 'mysql_no_login.so';"
$ make testacc
$ docker rm -f some-mysql
```
