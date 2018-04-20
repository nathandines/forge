# Forge

[![CircleCI](https://img.shields.io/circleci/project/github/nathandines/forge/master.svg)](https://circleci.com/gh/nathandines/forge)
[![GitHub Release](https://img.shields.io/github/release/nathandines/forge.svg)](https://github.com/nathandines/forge/releases/latest)
[![Github All Releases](https://img.shields.io/github/downloads/nathandines/forge/total.svg)](https://github.com/nathandines/forge/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/nathandines/forge)](https://goreportcard.com/report/github.com/nathandines/forge)
[![Docker Image](https://img.shields.io/badge/docker-nathandines%2Fforge-blue.svg)](https://hub.docker.com/r/nathandines/forge/)

## What is Forge, and why do I need it?

_Forge_ is a simple tool which makes deploying CloudFormation stacks a bit
easier in continuous delivery environments. In contrast to
[awscli](https://github.com/aws/aws-cli), _Forge_ provides a mechanism to
create, update and delete CloudFormation stacks synchronously, while also
providing output on the stack events, and an exit code reflecting the final
state of the stack deployment.

## Features

- Parameters and Tags defined as YAML/JSON files which contain a key-value
  object
- Lists in Parameter files will be collapsed into `CommaDelimitedLists` and
  passed into CloudFormation
- Only required parameters in a parameter file will be used, meaning you can
  share parameter files between stacks for common usage
- Automatic discovery and passthrough of CloudFormation capabilities (i.e.
  `CAPABILITY_IAM` and `CAPABILITY_NAMED_IAM`)
- Synchronous execution of actions against CloudFormation stacks
- Exit codes based on stack status
- Running stack event output on the command line
- Dynamically Create or Update stacks based on existing stack status
- Acceptance of "No updates to be performed." as a non-erroneous state
- Environment Variable Substitution in Parameter and Tag files

More features are currently on the roadmap, which can be [found on
Trello](https://trello.com/b/ECuGN86A)

## Available Parameters

To see what options are available to you, execute `forge --help` for the latest
help applicable to your version of _Forge_

## Getting Started

### Installation

```sh
go get -u github.com/nathandines/forge
```

### Using Environment Variables in Parameter or Tag files

Environment variables can be referenced within parameter and tag files by using
the following format: ``'{{ env `variable_name` }}'`` (the backticks **MUST**
surround the variable name). This is because under the covers, _Forge_ uses the
Golang text templating engine, with an additional function (`env`) to assist
with environment variable references.

**YAML Note:** The curly braces must be quoted when using YAML to ensure that
the field is interpreted as a string

#### Example

```yaml
---
Environment: '{{ env `ENVIRONMENT` }}'
Owner Email: '{{ env `USER` }}@example.com'
```

### Example: Deploying a stack with tags and parameters

#### Requirements

- Forge installed on your machine and available in your `PATH`
- AWS Account with permissions to create a DHCP Option Set through
  CloudFormation

#### Setting up your environment

Start in an empty folder. Create the following files which will cover your tags,
parameters, and CloudFormation template.

##### tags.yml

```yaml
---
Tag One: This is an example tag
CostAllocationTag: Cost Center
```

##### parameters.yml

```yaml
---
DomainName: example.com
DNSServers:
  - 10.0.0.1
  - 10.0.0.2
  - 10.0.0.3
  - 10.0.0.4
UnrelatedParameter: This Will Not Be Used
```

##### cfn_template.yml

```yaml
---
Parameters:
  DomainName:
    Type: String
  DNSServers:
    Type: CommaDelimitedList

Resources:
  DHCPOptions:
    Type: AWS::EC2::DHCPOptions
    Properties:
      DomainName: !Ref DomainName
      DomainNameServers: !Ref DNSServers
```

#### Deploying the example stack

Firstly, authenticate your CLI environment to AWS. _Forge_ uses [environment
variables to authenticate to AWS
services](https://docs.aws.amazon.com/cli/latest/userguide/cli-environment.html).
You could choose to use a tool such as
[awskeyring](https://github.com/vibrato/awskeyring) to setup your environment,
or reference an awscli profile using `AWS_DEFAULT_PROFILE`.

##### Deploying the stack

Once you're authenticated to the AWS services, you can now deploy your stack

```sh
forge deploy --stack-name test-stack \
  --template-file ./cfn_template.yml \
  --tags-file ./tags.yml \
  --parameters-file ./parameters.yml
```

During deployment, you'll see event output of the creation of the stack. After
deployment, upon logging into your AWS account, you should be able to see a new
DHCP option set which has been deployed with the tags and parameters defined
above.

### Development

#### Build

```sh
make
```

#### Test

```sh
make test
```

#### Linting

```sh
make lint
```

#### Clean workspace

```sh
make clean
```
