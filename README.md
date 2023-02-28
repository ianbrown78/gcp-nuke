# gcp-nuke

![Build Status](https://github.com/ianbrown78/gcp-nuke/.github/workflows/ci.yaml/badge.svg?branch=master)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://github.com/ianbrown78/gcp-nuke/blob/master/LICENSE)
[![GitHub release](https://img.shields.io/github/release/ianbrown78/gcp-nuke.svg)](https://github.com/ianbrown78/gcp-nuke/releases)
[![Docker Hub](https://img.shields.io/docker/pulls/ianbrown78/gcp-nuke)](https://hub.docker.com/repository/docker/ianbrown78/gcp-nuke)

Remove all resources from a GCP account.

> **Development Status** *gcp-nuke* is stable, but it is likely that not all GCP
resources are covered by it. Be encouraged to add missing resources and create
a Pull Request or to create an [Issue](https://github.com/ianbrown78/gcp-nuke/issues/new).

## Caution!

Be aware that *gcp-nuke* is a very destructive tool, hence you have to be very
careful while using it. Otherwise, you might delete production data.

**We strongly advise you to not run this application on any GCP project, where
you cannot afford to lose all resources.**

To reduce the blast radius of accidents, there are some safety precautions:

1. By default, *gcp-nuke* only lists all nukeable resources. You need to add
   `--no-dry-run` to actually delete resources.
2. *gcp-nuke* asks you twice to confirm the deletion by entering the project
   alias. The first time is directly after the start and the second time after
   listing all nukeable resources.
3. To avoid errors, your service account must have owner access to the project
   you are attempting to nuke the resources of. Otherwise, gcp-nuke will error
   and abort.
5. The config file contains a blocklist field. If the project ID of the project
   you want to nuke is part of this blocklist, *gcp-nuke* will abort. It is
   recommended, that you add every production account to this blocklist.
6. To ensure you don't just ignore the blocklisting feature, the blocklist must
   contain at least one project ID. By default, this is a non-existent project ID.
7. The config file contains project specific settings (eg. filters). The
   project you want to nuke must be explicitly listed there.
8. To ensure to not accidentally delete a random project, it is required to
   specify a config file. It is recommended to have only a single config file
   and add it to a central repository. This way the account blocklist is way
   easier to manage and keep up to date.

Feel free to create an issue, if you have any ideas to improve the safety
procedures.


## Use Cases

* We are testing our [Terraform](https://www.terraform.io/) code with Jenkins.
  Sometimes a Terraform run fails during development and messes up the account.
  With *gcp-nuke* we can simply clean up the failed project, so it can be reused
  for the next build.
* Our platform developers have their own GCP Projects where they can create
  their own Kubernetes clusters for testing purposes. With *gcp-nuke* it is
  very easy to clean up these projects at the end of the day and keep the costs
  low.

## Releases

We usually release a new version once enough changes came together and have
been tested for a while.

You can find Linux, macOS and Windows binaries on the
[releases page](https://github.com/ianbrown78/gcp-nuke/releases), but we also
provide containerized versions on [docker.io/ianbrown78/gcp-nuke](https://hub.docker.com/r/ianbrown78/gcp-nuke).  
Images are available for multiple architectures (amd64, arm64 & armv7).


## Usage

```
NAME:
   gcp-nuke - The GCP project cleanup tool with added radiation

USAGE:
   e.g. gcp-nuke --project test-nuke-123456 --dryrun --no-keep-project

VERSION:
   v0.1.0

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --project value   GCP project id to nuke (required)
   --no-dryrun       Do not perform a dryrun (default: false)
   --timeout value   Timeout for removal of a single resource in seconds (default: 400)
   --polltime value  Time for polling resource deletion status in seconds (default: 10)
   --no-keep-project Do not keep the project, destroy it with the resources.
   --help, -h        show help (default: false)
   --version, -v     print the version (default: false)
```

Example dryrun:

```buildoutcfg
./gcp-nuke --project gcp-nuke-test
2019/12/23 13:53:14 [Info] Retrieving zones for project: gcp-nuke-test
2019/12/23 13:53:14 [Info] Retrieving regions for project: gcp-nuke-test
2019/12/23 13:53:15 [Info] Timeout 400 seconds. Polltime 10 seconds. Dry run :true
2019/12/23 13:53:16 [Info] Retrieving list of resources for ContainerGKEClusters
2019/12/23 13:53:16 [Info] Retrieving list of resources for ComputeInstanceGroupsRegion
2019/12/23 13:53:16 [Info] Retrieving list of resources for ComputeDisks
2019/12/23 13:53:16 [Info] Retrieving list of resources for ComputeInstances
2019/12/23 13:53:16 [Info] Retrieving list of resources for ComputeRegionAutoScalers
2019/12/23 13:53:16 [Info] Retrieving list of resources for ComputeInstanceTemplates
2019/12/23 13:53:16 [Info] Retrieving list of resources for ComputeInstanceGroupsZone
2019/12/23 13:53:16 [Info] Retrieving list of resources for ComputeZoneAutoScalers
2019/12/23 13:53:16 [Dryrun] Resource type ComputeInstanceTemplates with resources [instance-template-1] would be destroyed [project: gcp-nuke-test]
2019/12/23 13:53:16 [Dryrun] [Skip] Resource type ContainerGKEClusters has nothing to destroy [project: gcp-nuke-test]
2019/12/23 13:53:22 [Dryrun] [Skip] Resource type ComputeRegionAutoScalers has nothing to destroy [project: gcp-nuke-test]
2019/12/23 13:53:22 [Dryrun] [Skip] Resource type ComputeInstanceGroupsRegion has nothing to destroy [project: gcp-nuke-test]
2019/12/23 13:53:32 [Dryrun] [Skip] Resource type ComputeZoneAutoScalers has nothing to destroy [project: gcp-nuke-test]
2019/12/23 13:53:32 [Dryrun] [Skip] Resource type ComputeInstances has nothing to destroy [project: gcp-nuke-test]
2019/12/23 13:53:32 [Dryrun] [Skip] Resource type ComputeDisks has nothing to destroy [project: gcp-nuke-test]
2019/12/23 13:53:33 [Dryrun] Resource type ComputeInstanceGroupsZone with resources [instance-group-1] would be destroyed [project: gcp-nuke-test]
2019/12/23 13:53:33 -- Deletion complete for project gcp-nuke-test (dry-run: true) --
```

As you see *gcp-nuke* now tries to delete all resources which aren't filtered,
without caring about the dependencies between them. This results in API errors
which can be ignored. These errors are shown at the end of the *gcp-nuke* run,
if they keep to appear.

*gcp-nuke* retries deleting all resources until all specified ones are deleted
or until there are only resources with errors left.

### GCP Credentials

*gcp-nuke* always uses either the attached credentials file (GOOGLE_APPLICATION_CREDENTIALS)
or the Application Default Credentials (ADC). Which you use is entirely up to you.

To use attached credentials, simply download the JSON key file and run:
```shell
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/your/key/file.json
```

To use ADC, follow the documentation [here](https://cloud.google.com/docs/authentication/provide-credentials-adc)

## Install

### For Mac
`brew install gcp-nuke`

### Use Released Binaries

The easiest way of installing it, is to download the latest
[release](https://github.com/ianbrown78/gcp-nuke/releases) from GitHub.

#### Example for Linux Intel/AMD

Download and extract
`$ wget -c https://github.com/ianbrown78/gcp-nuke/releases/download/v1.0.0/gcp-nuke-v1.0.0-linux-amd64.tar.gz -O - | sudo tar -xz -C $HOME/bin`

Run
`$ gcp-nuke-v1.0.0-linux-amd64`

### Compile from Source

To compile *gcp-nuke* from source you need a working
[Golang](https://golang.org/doc/install) development environment. The sources
must be cloned to `$GOPATH/src/github.com/ianbrown78/gcp-nuke`.

Also, you need to install [GNU
Make](https://www.gnu.org/software/make/).

Then you just need to run `make build` to compile a binary into the project
directory or `make install` go install *gcp-nuke* into `$GOPATH/bin`. With
`make xc` you can cross compile *gcp-nuke* for other platforms.

### Docker

You can run *gcp-nuke* with Docker by using a command like this:

```bash
$ docker run \
    --rm -it \
    docker.io/ianbrown78/resources-nuke:v1.0.0
```

To make it work, you need to supply the flags to the command as per the help.

Make sure you use the latest version in the image tag. Alternatively you can use
`main` for the latest development version, but be aware that this is more
likely to break at any time.


## Testing

### Unit Tests

To unit test *gcp-nuke*, some tests require [gomock](https://github.com/golang/mock) to run.
This will run via `go generate ./...`, but is automatically run via `make test`.
To run the unit tests:

```bash
make test
```


## Contact Channels

Feel free to create a GitHub Issue for any bug reports or feature requests.
Please use our mailing list for questions: gcp-nuke@googlegroups.com. You can
also search in the mailing list archive, whether someone already had the same
problem: https://groups.google.com/d/forum/gcp-nuke

## Contribute

You can contribute to *gcp-nuke* by forking this repository, making your
changes and creating a Pull Request against our repository. If you are unsure
how to solve a problem or have other questions about a contributions, please
create a GitHub issue.