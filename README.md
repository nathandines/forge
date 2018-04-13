# Forge

[![CircleCI](https://circleci.com/gh/nathandines/forge.svg?style=svg)](https://circleci.com/gh/nathandines/forge)
![GitHub (pre-)release](https://img.shields.io/github/release/nathandines/forge/all.svg)
![Github All Releases](https://img.shields.io/github/downloads/nathandines/forge/total.svg)

## What is Forge, and why do I need it?

_Forge_ is a simple tool which makes deploying CloudFormation stacks a bit more
friendly for continuous delivery environments. In contrast to
[awscli](https://github.com/aws/aws-cli), _Forge_ provides a mechanism to
create, update and delete CloudFormation stacks synchronously, while also
providing output on the stack events, and an exit code reflecting the final
state of the stack deployment.

## Features

- Exit codes based on stack status
- Stack event output on the command line

## Getting Started

### Installation

```sh
go get -u github.com/nathandines/forge
```

### Development

#### Build

```sh
make
```

#### Test

```sh
make test
```

#### Clean workspace

```sh
make clean
```
