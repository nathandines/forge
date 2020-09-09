# Forge

[![CircleCI](https://img.shields.io/circleci/project/github/nathandines/forge/master.svg)](https://circleci.com/gh/nathandines/forge)
[![GitHub Release](https://img.shields.io/github/release/nathandines/forge.svg)](https://github.com/nathandines/forge/releases/latest)
[![Github All Releases](https://img.shields.io/github/downloads/nathandines/forge/total.svg)](https://github.com/nathandines/forge/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/nathandines/forge)](https://goreportcard.com/report/github.com/nathandines/forge)
[![Docker Image](https://img.shields.io/badge/docker-nathandines%2Fforge-blue.svg)](https://hub.docker.com/r/nathandines/forge/)

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
- YAML and JSON formatted stack policies
- Apply stack policies to nested stacks
- Deploy using an assumed IAM role (often used to deploy stacks to other
  accounts)
  - Includes support for MFA specified on the command line or in `~/.aws/config`
- Enable Termination Protection at deployment time
- Define multiple parameter files to merge/override parameters
- Override specific parameters on the command line

## Available Parameters

To see what options are available to you, execute `forge --help` for the latest
help applicable to your version of _Forge_

## Installation

### macOS

On macOS, just use Homebrew to install and you're done!

```sh
brew tap nathandines/tap
brew install forge
```

### Windows

On Windows, just use Chocolatey to install and you're done!

```powershell
choco install forge
```

### Other

Go to the [latest release page on
GitHub](https://github.com/nathandines/forge/releases/latest) to download the
latest stable version.

Next, move the downloaded binary to a directory which is on your path, and
rename it to `forge`. On *nix systems, `~/bin` or `/usr/local/bin` are good
options depending on whether you want to restrict the install to just your user
or install it system-wide. On Windows systems, a similar pattern is advised;
`%USERPROFILE%\bin` for a single user, or `%PROGRAMDATA%\bin` for multiple
users.

The final step to installation is to make sure the directory you installed
_Forge_ to is on the PATH. See [this
page](https://stackoverflow.com/questions/14637979/how-to-permanently-set-path-on-linux)
for instructions on setting the PATH on Linux and Mac. [This
page](https://stackoverflow.com/questions/1618280/where-can-i-set-path-to-make-exe-on-windows)
contains instructions for setting the PATH on Windows.

#### Adding bash or zsh completion (optional)

_Forge_ has the capability to generate shell completion for bash and zsh. Run
one of the following commands (adjusting the destination for the output file as
required for your machine) to enable shell completion for _Forge_ on your
system.

If you're not sure which shell you use, you probably use bash.

```sh
forge gen-bash-completion > /etc/bash_completion.d/forge
# or
forge gen-zsh-completion > ~/.zsh_completions.d/forge
```

## Feature Usage

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

##### parameters1.yml

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

##### parameters2.yml

```yaml
---
DomainName: foobar.com
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
  --parameters-file ./parameters1.yml \
  --parameters-file ./parameters2.yml
```

During deployment, you'll see event output of the creation of the stack. After
deployment, upon logging into your AWS account, you should be able to see a new
DHCP option set which has been deployed with the tags and parameters defined
above.

### Development

#### Requirements

- GNU Make
- [Go v1.11+](https://golang.org/)

#### Build

```sh
make build
```

#### Test

```sh
make test
```

#### Linting

```sh
make lint
```

#### Update Dependencies

```sh
make update-deps
```

#### Clean workspace

```sh
make clean
```

#### Change AWS Service Endpoints

You can currently change the service endpoints for both CloudFormation and STS by setting the following environment variables when running _Forge_:

- AWS_ENDPOINT_CLOUDFORMATION
- AWS_ENDPOINT_IAM
- AWS_ENDPOINT_STS
