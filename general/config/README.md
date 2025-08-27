# WME Projects Configuration

This repository contains all of the configurations for things like, topic partitions configuration, what projects and namespaces are being supported by the system. With golang package to manipulate the data.

# Testing

A Go module is needed for `go test`, so run the following:

```sh
$ go mod init wikimedia-enterprise/general/config && go tidy
$ go test .
```
