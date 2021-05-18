# GTS: Genome Transformation Subprograms
An application and library package for genome flat-file manipulation.

## Quick Start
1. Extracting CDS sequences from a GenBank file:

```sh
$ gts extract CDS <file> > cds_sequences.gb
```

2. Retrieving basic information such as the composition of a sequence:

```sh
$ gts summary <file>
```

3. Getting feature attributes:

```sh
$ gts query <file>
```

Install the GTS CLI tools and try it out yourself!
For a more elaborate walkthrough, check out the [GTS Cookbook](https://github.com/go-gts/gts/wiki/cookbook) to discover the many functionalities of GTS.

## Installing the GTS CLI tools
### Binary Download
By far the easiest way to quickly use the GTS CLI tools (no library) is to simply download the binary executable file.
Simply navigate to the latest release (https://github.com/go-gts/gts/releases/latest) and download one of the `.tar.gz` files for your platform.
Darwin x86_64 is targeted for macOS users, Linux i386 is targeted for 32 bit Linux-based OS users, and Linux x86_64 is targeted for 64 bit Linux-based OS users.
The Linux binaries work on many popular Linux distributions including, but not limited to, Ubuntu, Debian, and CentOS.

### Package Managers (Recommended)
We recommend users who just want the GTS CLI tools to install via their favorite package managers.
GTS currently supports [Homebrew](https://brew.sh), [Anaconda/Miniconda](https://www.anaconda.com), dpkg, and yum.

#### With Homebrew
```sh
$ brew install go-gts/gts/gts-bio
```

#### With Anaconda/Miniconda
```sh
$ conda install -c ktnyt gts-bio
```

#### With dpkg
1. Download the deb package with the command of your choice.
```sh
# 32 bit with wget
$ wget https://github.com/go-gts/gts/releases/download/v0.27.1/gts_0.27.1_linux_386.deb
# 32 bit with curl
$ curl -LO https://github.com/go-gts/gts/releases/download/v0.27.1/gts_0.27.1_linux_386.deb
# 64 bit with wget
$ wget https://github.com/go-gts/gts/releases/download/v0.27.1/gts_0.27.1_linux_amd64.deb
# 64 bit with curl
$ curl -LO https://github.com/go-gts/gts/releases/download/v0.27.1/gts_0.27.1_linux_amd64.deb
```

2. Install the deb package with dpkg.
```sh
# 32 bit
$ dpkg --install gts_0.27.1_linux_386.deb
# 64 bit
$ dpkg --install gts_0.27.1_linux_amd64.deb
```

3. Remove the deb package file.
```sh
# 32 bit
$ rm gts_0.27.1_linux_386.deb
# 64 bit
$ rm gts_0.27.1_linux_amd64.deb
```

#### With yum
```sh
# 32 bit
$ yum install -y https://github.com/go-gts/gts/releases/download/v0.27.1/gts_0.27.1_linux_386.rpm
# 64 bit
$ yum install -y https://github.com/go-gts/gts/releases/download/v0.27.1/gts_0.27.1_linux_amd64.rpm
```

## Shell Completions
GTS provides bash and zsh completion scripts for better usability. The bash completion will be installed in `/usr/local/etc/bash_completion.d` with Homebrew, `/etc/bash_completion.d` with dpkg/yum, and `$CONDA_ROOT/share/bash-completion/completions` with conda (see [conda-bash-completion](https://github.com/tartansandal/conda-bash-completion) for more details on using bash completion with conda). The zsh completion will be installed in `/usr/local/share/zsh/site-functions` with Homebrew, dpkg and yum, and `$CONDA_ROOT/share/zsh/site-functions` with conda. Adding `fpath+=$CONDA_ROOT/share/zsh/site_fucntions` to your `.zshrc` before calling `compinit` will enable zsh completion. Be sure to have the envionrment variable `CONDA_ROOT` be set appropriately.

If you want to set up completions manually, download them from the following URLs.

- https://github.com/go-gts/gts/releases/download/v0.27.1/gts-completion.bash
- https://github.com/go-gts/gts/releases/download/v0.27.1/gts-completion.zsh

## Using the GTS library
The GTS library requires the use of [Go Modules](https://blog.golang.org/using-go-modules). Therefore a Go distribution with version 1.13 or later is highly recommended. To use the GTS library in your project, initialize your module as per protocol and type the following command:

```sh
$ go get github.com/go-gts/gts/...@latest
```
