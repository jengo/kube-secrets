# kube-secrets [![Build Status](https://travis-ci.org/jengo/kube-secrets.svg?branch=master)](https://travis-ci.org/jengo/kube-secrets) [![Coverage Status](https://coveralls.io/repos/github/jengo/kube-secrets/badge.svg?branch=HEAD)](https://coveralls.io/github/jengo/kube-secrets?branch=HEAD)
Command line editor for Kubernetes secrets

Updating Kubernetes secrets is honestly a pain in the ass. This utility simplifies the process by allowing you to modify keys using your favorite editor.  The base64 encoding is transparent.

This repo is setup to run and build the application inside a Docker container.  For development, simply run make depends shell.  From there you can edit on your host.  And run the application with go run kube-secrets.go.

*NOTE THIS IS A VERY EARLY STAGE APPLICATION, USE WITH CAUTION*

## Examples
### Create or update Kubernetes secret by key
```
kube-secrets update test_data/sample.yml MYSQL_ROOT_PASSWORD
```

### Create new Kubernetes secret file
```
kube-secrets create mysql-secret.yml MYSQL_ROOT_PASSWORD
```

## Show value of Kubernetes secret by key
```
kube-secrets show test_data/sample.yml MYSQL_ROOT_PASSWORD
```
