In order to build the plugins is necessary to install the dependencies of this project and also the one of docker.

I have used [`dep`](https://golang.github.io/dep/) to install the dependencies of this project, simply running `dep ensure` on the root.

It will populate the `vendor/` dir.

The it is necessary to install the dependencies of docker.

So we move inside `vendor/github.com/docker/docker/` and there should be a file called `vendor.conf` which tracks each dependencies along with their commit hash.

To install them I believe there are several tool but I simply used the first one from google: [`trash`](https://github.com/rancher/trash).

To recap.

From the root of the project:

```
dep ensure
cd vendor/github.com/docker/docker/
trash
```

And now it should be possible to build the two plugins.

Again from the project root

```
cd plugins/overlay2_cvmfs
go build
cd ../aufs_cvmfs
go build
```
