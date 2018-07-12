# Automatic conversion of docker images into thin format

This utility will automatically convert normal docker images into the thin
format.

## Vocabulary

There are several concept to keep track in this process, and none of them are
very common, so before to dive in we can agree on a share vocabulary.

**Registry** does refer to the docker image registry, with protocol extensions,
common examples are:
    * https://registry.hub.docker.com
    * https://gitlab-registry.cern.ch

**Repository** This specify a containers of images, each image will be indexed,
then by tag or digest. Common examples are:
    * library/redis
    * library/ubuntu

**Tag** is a way to identify an image inside a repository, tags are mutable
and may change in a feature. Common examples are:
    * 4
    * 3-alpine

**Digest** is another way to identify images inside a repository, digest are
**immutable**, since they are the result of an hash function to the content of
the image. Thanks to this technique the images are content addreassable.
Common examples are:
    * sha256:2aa24e8248d5c6483c99b6ce5e905040474c424965ec866f7decd87cb316b541
    * sha256:d582aa10c3355604d4133d6ff3530a35571bd95f97aadc5623355e66d92b6d2c


An **image** belong to a repository -- which in turns belongs to a registry --
and it is identified by a tag, or a digest or both, if you can choose is always
better to identify the image using at least the digest.

To unique identify an image so we need to provide all those information:
    1. registry
    2. repository
    3. tag or digest or tag + digest

We will use colon (`:`) to separate the `registry` from the `repository` and
again the colon to separate the `repository` from the `tag` and the at (`@`) to
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

    `desiderata => (input_image, output_image, cvmfs_repository)`

The input image in your desiderata should be as more specific as possible,
ideally specifying both the tag and the digest.

On the other end, you cannot be so specific for the output image, simple
because is impossible to know the digest before to generate the image itself.

Finally we use model the repository as an append only structure, deleting
layers could break some images actually running.

## Commands

Here follow the list of commands that the converter understand.

### add-desiderata

**add-desiderata** --input-image $INPUT\_IMAGE --output-image $OUTPUT\_IMAGE --repository $CVMFS\_REPO

Will add a new `desiderata` to the internal database, then it will try to
convert the regular image into a thin image.

### check-image

**check-image-syntax** $IMAGE

Will parse your image and output what it is been able to parse.
