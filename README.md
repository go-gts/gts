# GTS: Genomics Tools and Subprograms
An application and library package for genome manipulation.

## Installing the GTS CLI tools
### Package Managers (Recommended)
We recommend users who just want the GTS CLI tools (no library) to install via their favorite package managers.
GTS currently supports [Homebrew](https://brew.sh), apt, and yum.

#### With Homebrew
```sh
$ brew install go-gts/gts/gts-bio
```

#### With apt
1. Download the deb package with the command of your choice.
```sh
# 32 bit with wget
$ wget https://github.com/go-gts/gts/releases/download/v0.14.0/gts_0.14.0_linux_386.deb
# 32 bit with curl
$ curl -LO https://github.com/go-gts/gts/releases/download/v0.14.0/gts_0.14.0_linux_386.deb
# 64 bit with wget
$ wget https://github.com/go-gts/gts/releases/download/v0.14.0/gts_0.14.0_linux_amd64.deb
# 64 bit with curl
$ curl -LO https://github.com/go-gts/gts/releases/download/v0.14.0/gts_0.14.0_linux_amd64.deb
```

2. Install the deb package with dpkg.
```sh
# 32 bit
$ dpkg --install gts_0.14.0_linux_386.deb
# 64 bit
$ dpkg --install gts_0.14.0_linux_amd64.deb
```

3. Remove the deb package file.
```sh
# 32 bit
$ rm gts_0.14.0_linux_386.deb
# 64 bit
$ rm gts_0.14.0_linux_amd64.deb
```

#### With yum
```sh
# 32 bit
$ yum install -y https://github.com/go-gts/gts/releases/download/v0.14.0/gts_0.14.0_linux_386.rpm
# 64 bit
$ yum install -y https://github.com/go-gts/gts/releases/download/v0.14.0/gts_0.14.0_linux_amd64.rpm
```

## Using the GTS library
The GTS library requires the use of [Go Modules](https://blog.golang.org/using-go-modules). Therefore a Go distribution with version 1.13 or later is highly recommended. To use the GTS library in your project, initialize your module as per protocol and type the following command:

```sh
$ go get github.com/go-gts/gts/...@latest
```
