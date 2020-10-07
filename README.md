# GTS: Genome Transformation Subprograms
An application and library package for genome manipulation.

## Installing the GTS CLI tools
### Package Managers (Recommended)
We recommend users who just want the GTS CLI tools (no library) to install via their favorite package managers.
GTS currently supports [Homebrew](https://brew.sh), [Anaconda/Miniconda](https://www.anaconda.com), apt, and yum.

#### With Homebrew
```sh
$ brew install go-gts/gts/gts-bio
```

#### With Anaconda/Miniconda
```sh
$ conda install -c ktnyt gts-bio
```

#### With apt
1. Download the deb package with the command of your choice.
```sh
# 32 bit with wget
$ wget https://github.com/go-gts/gts/releases/download/v0.24.1/gts_0.24.1_linux_386.deb
# 32 bit with curl
$ curl -LO https://github.com/go-gts/gts/releases/download/v0.24.1/gts_0.24.1_linux_386.deb
# 64 bit with wget
$ wget https://github.com/go-gts/gts/releases/download/v0.24.1/gts_0.24.1_linux_amd64.deb
# 64 bit with curl
$ curl -LO https://github.com/go-gts/gts/releases/download/v0.24.1/gts_0.24.1_linux_amd64.deb
```

2. Install the deb package with dpkg.
```sh
# 32 bit
$ dpkg --install gts_0.24.1_linux_386.deb
# 64 bit
$ dpkg --install gts_0.24.1_linux_amd64.deb
```

3. Remove the deb package file.
```sh
# 32 bit
$ rm gts_0.24.1_linux_386.deb
# 64 bit
$ rm gts_0.24.1_linux_amd64.deb
```

#### With yum
```sh
# 32 bit
$ yum install -y https://github.com/go-gts/gts/releases/download/v0.24.1/gts_0.24.1_linux_386.rpm
# 64 bit
$ yum install -y https://github.com/go-gts/gts/releases/download/v0.24.1/gts_0.24.1_linux_amd64.rpm
```

## Shell Completions
GTS provides bash and zsh completion scripts for better usability. The bash completion will be installed in `/usr/local/etc/bash_completion.d` with Homebrew, `/etc/bash_completion.d` with apt/yum, and `$CONDA_ROOT/share/bash-completion/completions` with conda (see [conda-bash-completion](https://github.com/tartansandal/conda-bash-completion) for more details on using bash completion with conda). The zsh completion will be installed in `/usr/local/share/zsh/site-functions` with Homebrew, apt and yum, and `$CONDA_ROOT/share/zsh/site-functions` with conda. Adding `fpath+=$CONDA_ROOT/share/zsh/site_fucntions` to your `.zshrc` before calling `compinit` will enable zsh completion. Be sure to have the envionrment variable `CONDA_ROOT` be set appropriately.

If you want to set up completions manually, download them from the following URLs.

- https://github.com/go-gts/gts/releases/download/v0.24.1/gts-completion.bash
- https://github.com/go-gts/gts/releases/download/v0.24.1/gts-completion.zsh

## Using the GTS library
The GTS library requires the use of [Go Modules](https://blog.golang.org/using-go-modules). Therefore a Go distribution with version 1.13 or later is highly recommended. To use the GTS library in your project, initialize your module as per protocol and type the following command:

```sh
$ go get github.com/go-gts/gts/...@latest
```
