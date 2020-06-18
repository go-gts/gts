# GTS: The Genomics Tool Suite
A software suite for basic genome flatfile manipulation.

## Installing the GTS CLI tools
### Package Managers (Recommended)
We recommend users who just want the GTS CLI tools (no library) to install via their favorite package managers.
GTS currently supports [Homebrew](https://brew.sh), apt and yum.

#### Homebrew
```sh
$ brew install go-gts/gts/gts-bio
```

#### apt
```sh
# For 32 bit computers.
$ apt-get install https://github.com/go-gts/gts/releases/download/v0.9.5/gts_0.9.5_linux_386.deb
# For 64 bit conmputers.
$ apt-get install https://github.com/go-gts/gts/releases/download/v0.9.5/gts_0.9.5_linux_amd64.deb
```

#### yum
```sh
# For 32 bit computers.
$ yum install https://github.com/go-gts/gts/releases/download/v0.9.5/gts_0.9.5_linux_386.rpm
# For 64 bit computers.
$ yum install https://github.com/go-gts/gts/releases/download/v0.9.5/gts_0.9.5_linux_amd64.rpm
```

## Using the GTS library
The GTS library requires the use of [Go Modules](https://blog.golang.org/using-go-modules). Therefore a Go distribution with version 1.13 or later is highly recommended. To use the GTS library in your project, initialize your module as per protocol and type the following command:

```sh
$ go get github.com/go-gts/gts/...@latest
```
