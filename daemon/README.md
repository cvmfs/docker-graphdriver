# Automatic conversion of docker images into the thin format

This utility will automatically convert normal docker images into the thin
format.

## Vocabulary

There are several concepts to keep track in this process, and none of them is
very common, so before to dive in we can agree on a shared vocabulary.

**Registry** does refer to the docker image registry, with protocol extensions,
common examples are:

    * https://registry.hub.docker.com
    * https://gitlab-registry.cern.ch

**Repository** This specifies a class of images, each image will be indexed,
then by tag or digest. Common examples are:
 
    * library/redis
    * library/ubuntu

**Tag** is a way to identify an image inside a repository, tags are mutable
and may change in a feature. Common examples are:

    * 4
    * 3-alpine

**Digest** is another way to identify images inside a repository, digests are
**immutable**, since they are the result of a hash function to the content of
the image. Thanks to this technique the images are content addressable.
Common examples are:

    * sha256:2aa24e8248d5c6483c99b6ce5e905040474c424965ec866f7decd87cb316b541
    * sha256:d582aa10c3355604d4133d6ff3530a35571bd95f97aadc5623355e66d92b6d2c


An **image** belongs to a repository -- which in turns belongs to a registry --
and it is identified by a tag, or a digest or both, if you can choose is always
better to identify the image using at least the digest.

To unique identify an image so we need to provide all those information:

    1. registry
    2. repository
    3. tag or digest or tag + digest

We will use slash (`/`) to separate the `registry` from the `repository` and
the colon (`/`) to separate the `repository` from the `tag` and the at (`@`) to
separate the `digest` from the tag or from the `repository`.

The final syntax will be:

    REGISTRY/REPOSITORY[:TAG][@DIGEST]

Examples of images are:
    * https://registry.hub.docker.com/library/redis:4
    * https://registry.hub.docker.com/minio/minio@sha256:b1e5dd4a7be831107822243a0675ceb5eabe124356a9815f2519fe02beb3f167
    * https://registry.hub.docker.com/wurstmeister/kafka:1.1.0@sha256:3a63b48894bce633fb2f0d2579e162163367113d79ea12ca296120e90952b463

## Concepts

The converter has a declarative approach. You specify what is your end goal and
it tries to reach it.

The main component of this approach is the **desiderata** which is a triplet
composed by the input image, the output image and in which cvmfs repository you
want to store the data.

    desiderata => (input_image, output_image, cvmfs_repository)

The input image in your desiderata should be as more specific as possible,
ideally specifying both the tag and the digest.

On the other end, you cannot be so specific for the output image, simple
because is impossible to know the digest before to generate the image itself.

Finally we model the repository as an append only structure, deleting
layers could break some images actually running.

## Commands

Here follow the list of commands that the converter understand.

### add-desiderata

```
add-desiderata --input-image $INPUT_IMAGE --output-image $OUTPUT_IMAGE --repository $CVMFS_REPO \
        --user-input $USER_INPUT --user-output $USER_OUTPUT
```

Will add a new `desiderata` to the internal database, then it will try to
convert the regular image into a thin image.

The users are the one that will try to log into the registry, you can add
users (so usernames, password and registry) using the `add-user` command.

### add-image

```
add-image $IMAGE
```

Will add the image to the internal database

### check-image-syntax

```
check-image-syntax $IMAGE
```

Will parse your image and output what it is been able to parse.

### list-images

```
list-images
```

List all the images in the database

### migrate-database

```
migrate-database
```

Apply all the migration to the database up to the newest version of the
software.

At the first run is necessary to run this function.

### download-manifest

```
download-manifest $IMAGE
```

Will try to download the manifest of the image from the repository, if
successful it will print the manifest itself, otherwise it will display the
error. The same internal procedure is used in order to actually convert the
images.

### convert

```
convert
```

This command will try to convert all the desiderata in the internal database.

### loop

```
loop
```

This command is equivalent to call `convert` in an infinite loop, useful to
make sure that all the images are up to date.


## add-desiderata workflow

This section will go into the detail of what happens when you try to add a
desiderata.

The very first step is the parse of both the input and output image, if any of
those parse fails the whole command fails and we immediately return an error.

Then we check if the desiderata we are trying to add is already in the
database, if it is we are not going to add it again and we simply return an
error.

The next step is trying to download the input image manifest, if we are not
able to access the input manifest we return an error.

Finally if every check completed successfully we add the desiderata to the
internal database.

## convert workflow

The goal of convert is to actually create the thin images starting from the
regular one.

In order to convert we iterate for every desiderata.

In general, some desiderata will be already converted while others will need to
be converted ex-novo.

The first step is then to check if the desiderata is already been converted.
In order to do this check, we download the input image manifest and check
against the internal database if the input image digest is already been
converted, if it is we can safely skip such conversion. 

Then, every image is made of different layers, some of them could already be
on the repository.
In order to avoid expensive CVMFS transaction, before to download and ingest
the layer we check if it is already in the repository, if it is we do not
download nor ingest the layer.

The conversion simply ingest every layer in an image, create a thin image and
finally push the thin image to the registry.

Such images can be used by docker with the plugins.

## General workflow

This section explains how this utility is intended to be used.

Internally this utility invokes `cvmfs_server` and `docker` commands, so it is
necessary to use it in a stratum0 that also have docker installed. 

The docker dependency can be dropped, but it would require some amount of work,
so for this first release, as long as it is not a big hurdle, we are going to
keep it.

The first time the utility is launched is necessary to create the SQLite
database, to do so you can call the command `migrate-database` or its alias,
`init`.

This command, create as SQLite database called `docker2cvmfs_archive.sqlite`,
the utility will require this file to always be on `.`, the directory from
where you are calling the utility itself, this requirements will be dropped in
future releases.

Once the database is been created we can start adding users, images and
desideratas.

The conversion is quite straightforward, we first download the input image, we
store each layer on the cvmfs repository, we create the output image and
finally we upload the output image to the registry.

For downloading an image the credentials can be not necessary, while for
uploading it they are mandatory.

Also, you may want to have different users upload different images to the same
docker registry, maybe even one user for image.

The first step is so to call `add-user`.

```
$ ./daemon init
INFO[0000] Made migrations                               n=2
$ ./daemon add-user --username foo --password secret --registry docker.foo.bar.com
$ ./daemon list-users
+------+--------------------+
| USER |      REGISTRY      |
+------+--------------------+
| foo  | docker.foo.bar.com |
+------+--------------------+
```

I wasn't able to figure out a reliable way to get authentication tokens so
to avoid storing the password as clear text in the database, the suggestion at
the moment is to use disposable users with very limited capabilities so that
if the database get compromised (a third party has access to it) we are able to
limit the treats.

The next step is to add a desiderata, to do so:

```
$ ./daemon add-desiderata \
        --input-image https://registry.hub.docker.com/library/redis:4 \
        --output-image https://gitlab-registry.cern.ch/smosciat/containerd/thin/redis:4 \
        --repository cd.cern.ch \
        --user-output smosciat
WARN[0000] Unable to retrieve the password, trying to get the manifest anonymously.  error="sql: no rows in result set"
Auth to: Bearer realm="https://auth.docker.io/token",service="registry.docker.io",scope="repository:library/redis:pull"
https://auth.docker.io/token?scope=repository%3Alibrary%2Fredis%3Apull&service=registry.docker.io

$ ./daemon list-desideratas
+----+----------------+-------------------------------------------------+------------+-----------------+------------------------------------------------------------------+
| ID | INPUT IMAGE ID |                INPUT IMAGE NAME                 | CVMFS REPO | OUTPUT IMAGE ID |                        OUTPUT IMAGE NAME                         |
+----+----------------+-------------------------------------------------+------------+-----------------+------------------------------------------------------------------+
| 1  | 1              | https://registry.hub.docker.com/library/redis:4 | cd.cern.ch | 2               | https://gitlab-registry.cern.ch/smosciat/containerd/thin/redis:4 |
+----+----------------+-------------------------------------------------+------------+-----------------+------------------------------------------------------------------+
```

Of ocurse you can add as many desideratas as you wish.

Now that all the desideratas are in place you can simply start converting them:

```
$ ./daemon convert
```

The above command should provide enough logs to be able to infer what is
happening and to debug any error. 

Make sure that the user is able to start a cvmfs transaction and that is able
to communicate with docker, anyway this errors should be pretty self evidentds
in the logs.

The above command is quite cheap, it avoids to convert an images that is
already been converted and it avoid to download layers that are already been
downloaded, command line flags can change this behaviour if necessary.

You may want to keep the above command running in a loop, hence it will
automatically pick up changes in the input images and start the conversion.

We are basically polling the registries for changings in the input image, again
there was not a reliable and easy way to get updates from the registry, not
even from the one inside CERN that we manage.

In order to run the conversion in a loop you can simply use:

```
$./daemon loop
```

While the daemon is running in a loop you should be able to iteract with the
utility without any issue, so you should be able to add users, images and even
desideratas.

Only be careful to don't leave the CVMFS repository in an inconsistet state
(abort the program Ctrl-C while it is doing a transaction).

