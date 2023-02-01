# GCP-Nuke

## Background

Inspired by aws-nuke and cloud-nuke.

This tool was created out of my personal frustration with cleaning up GCP projects. 

But why?

Many reasons:

The behaviour of gcloud projects delete is to disable a project - pending a 30 day wait time for any resource removal. Sometimes, you want to just want to remove everything; e.g. SharedVPCs - gcloud project delete on service projects can cause "ghost subnets" on the Host project. Yes, you end up with undeletable subnets due to VM resources, and end up having to 'undelete' the gcp project. Google support's solution? Well ofcourse, "just don't do it" - https://cloud.google.com/vpc/docs/deprovisioning-shared-vpc.

Additionally, I've found Terraform destroy of some of my colleagues' wizard level terraform modules fail occasionally, so it's always neat to see what's not been deleted via a dryrun.

## Usage

```
NAME:
   gcp-nuke - The GCP project cleanup tool with added radiation

USAGE:
   e.g. gcp-nuke --project test-nuke-123456 --dryrun --keep-project

VERSION:
   v0.1.0

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --project value   GCP project id to nuke (required)
   --dryrun          Perform a dryrun instead (default: false)
   --timeout value   Timeout for removal of a single resource in seconds (default: 400)
   --polltime value  Time for polling resource deletion status in seconds (default: 10)
   --keep-project    Keep the project, just destroy the resources.
   --help, -h        show help (default: false)
   --version, -v     print the version (default: false)
```

Example dryrun

```
./gcp-nuke --project test-nuke-123456 --dryrun
2019/12/23 13:53:14 [Info] Retrieving zones for project: test-nuke-123456
2019/12/23 13:53:14 [Info] Retrieving regions for project: test-nuke-123456
2019/12/23 13:53:15 [Info] Timeout 400 seconds. Polltime 10 seconds. Dry run :true
2019/12/23 13:53:16 [Info] Retrieving list of resources for ContainerGKEClusters
2019/12/23 13:53:16 [Info] Retrieving list of resources for ComputeInstanceGroupsRegion
2019/12/23 13:53:16 [Info] Retrieving list of resources for ComputeDisks
2019/12/23 13:53:16 [Info] Retrieving list of resources for ComputeInstances
2019/12/23 13:53:16 [Info] Retrieving list of resources for ComputeRegionAutoScalers
2019/12/23 13:53:16 [Info] Retrieving list of resources for ComputeInstanceTemplates
2019/12/23 13:53:16 [Info] Retrieving list of resources for ComputeInstanceGroupsZone
2019/12/23 13:53:16 [Info] Retrieving list of resources for ComputeZoneAutoScalers
2019/12/23 13:53:16 [Dryrun] Resource type ComputeInstanceTemplates with resources [instance-template-1] would be destroyed [project: test-nuke-123456]
2019/12/23 13:53:16 [Dryrun] [Skip] Resource type ContainerGKEClusters has nothing to destroy [project: test-nuke-123456]
2019/12/23 13:53:22 [Dryrun] [Skip] Resource type ComputeRegionAutoScalers has nothing to destroy [project: test-nuke-123456]
2019/12/23 13:53:22 [Dryrun] [Skip] Resource type ComputeInstanceGroupsRegion has nothing to destroy [project: test-nuke-123456]
2019/12/23 13:53:32 [Dryrun] [Skip] Resource type ComputeZoneAutoScalers has nothing to destroy [project: test-nuke-123456]
2019/12/23 13:53:32 [Dryrun] [Skip] Resource type ComputeInstances has nothing to destroy [project: test-nuke-123456]
2019/12/23 13:53:32 [Dryrun] [Skip] Resource type ComputeDisks has nothing to destroy [project: test-nuke-123456]
2019/12/23 13:53:33 [Dryrun] Resource type ComputeInstanceGroupsZone with resources [instance-group-1] would be destroyed [project: test-nuke-123456]
2019/12/23 13:53:33 -- Deletion complete for project test-nuke-123456 (dry-run: true) --
```

## Roadmap
- Add BigQuery
- Add CloudKMS & Secrets Manager
- Add DataFlow
- Add Cloud Functions
- Add Cloud Run
- Migrate from urfave/cli to spf13/cobra