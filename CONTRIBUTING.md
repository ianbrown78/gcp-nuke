# Contributing

Thank you for wanting to contribute to *gcp-nuke*.

Because of the amount of GCP services and their rate of change, we rely on your
participation. For the same reason we can only act retroactive on changes of
GCP services. Otherwise, it would be a fulltime job to keep up with GCP.


## How Can I Contribute?

### Some Resource Is Not Supported by *gcp-nuke*

If a resource is not yet supported by *gcp-nuke*, you have two options to
resolve this:

* File [an issue](https://github.com/ianbrown78/gcp-nuke/issues/new) and describe
  which resource is missing. This way someone can take care of it.
* Add the resource yourself and open a Pull Request. Please follow the
  guidelines below to see how to create such a resource.


### Some Resource Does Not Get Deleted

Please check the following points before creating a bug issue:

* Is the resource actually supported by *gcp-nuke*? If not, please follow the
  guidelines above.
* Are there permission problems? In this case *gcp-nuke* will print errors
  that usually contain the status code `403`.
* Did you just get scared by an error that was printed? *gcp-nuke* does not
  know about dependencies between resources. To work around this it will just
  retry deleting all resources in multiple iterations. Therefore, it is normal
  that there are a lot of dependency errors in the first one. The iterations
  are separated by lines starting with `Removal requested: ` and only the
  errors in the last block indicate actual errros.

File [an issue](https://github.com/ianbrown78/gcp-nuke/issues/new) and describe
as accurately as possible how to generate the resource on GCP that cause the
errors in *gcp-nuke*. Ideally this is provided in a reproducible way like
a Terraform template or GCP CLI commands.


### I Have Ideas to Improve *gcp-nuke*

You should take these steps if you have an idea how to improve *gcp-nuke*:

1. Check the [issues page](https://github.com/ianbrown78/gcp-nuke/issues),
   whether someone already had the same or a similar idea.
2. Also check the [closed
   issues](https://github.com/ianbrown78/gcp-nuke/issues?utf8=%E2%9C%93&q=is%3Aissue),
   because this might have already been implemented, but not yet released. Also,
   the idea might not be viable for unobvious reasons.
3. Join the discussion, if there is already an related issue. If this is not
   the case, open a new issue and describe your idea. Afterwards, we can
   discuss this idea and form a proposal.


### I Just Have a Question

Please use our mailing list for questions: gcp-nuke@googlegroups.com. You can
also search in the mailing list archive, whether someone already had the same
problem: https://groups.google.com/d/forum/gcp-nuke


## Resource Guidelines

### Consider Pagination

Most GCP resources are paginated and all resources should handle that.


### Use Properties Instead of String Functions

Currently, each resource can offer two functions to describe itself, that are
used by the user to identify it and by *gcp-nuke* to filter it.

The String function is deprecated:

```go
String() string
```

The Properties function should be used instead:

```go
Properties() types.Properties
```

The interface for the String function is still there, because not all resources
are migrated yet. Please use the Properties function for new resources.


## Styleguide

### Go

#### Code Format

Like almost all Go projects, we are using `go fmt` as a single source of truth
for formatting the source code. Please use `go fmt` before committing any
change.


### Git

#### Setup Email

We prefer having the commit linked to the GitHub account, that is creating the
Pull Request. To make this happen, *git* must be configured with an email, that
is registered with a GitHub account.

To set the email for all git commits, you can use this command:

```
git config --global user.email "email@example.com"
```

If you want to change the email only for the *gcp-nuke* repository, you can
skip the `--global` flag. You have to make sure that you are executing this in
the *gcp-nuke* directory:

```
git config user.email "email@example.com"
```

If you already committed something with a wrong email, you can use this command:

```
git commit --amend --author="Author Name <email@address.com>"
```

This changes the email of the latest commit. If you have multiple commits in
your branch, please squash them and change the author afterwards.