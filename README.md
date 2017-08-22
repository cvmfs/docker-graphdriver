# docker-graphdriver
CernVM-FS Graph Driver for Docker

# Introduction

Docker is a popular containerization technology both in industry and in science
because it makes containers easy to create, manage and share. Docker images are
composed of shared layers to save space and bandwidth, but the price for
obtaining image is very high. For instance, to start a cluster of 1000 nodes
using an image of 1 GB your network has to transfer 1 TB at once and as quickly
as possible.

On average only 6% of the image is required [1]. We developed a Docker plugin
[2] to refine data reuse granularity from layers to files. Additionally, the
download is delayed until the file is accessed for the first time. Using this
approach and CernVM-FS [3], only a fraction of the image is initially
transferred and containers start instantly.

This project is presented at ACAT 2017 in Seattle. You can get the poster in PDF
here: [Making Containers Lazy with Docker and CernVM-FS](https://cernbox.cern.ch/index.php/s/ztVY6EgRua5IxIj).

# Quickstart

You can obtain and test this plugin this way:

```
$ docker plugin install cvmfs/overlay2-graphdriver
$ docker plugin enable cvmfs/overlay2-graphdriver

# Restart Docker daemon with flags --experimental -s cvmfs/overlay2-graphdriver
$ docker run -it cvmfs/thin_ubuntu echo "Hello ACAT 2017"
```

*Notice:* make sure that the overlay Linux kernel module is available and loaded
(`sudo modprobe overlay`).

# Project Status

This project is in an early experimental phase. It can be used in realistic
scenarios but it hasn't yet reached production quality.

# Contact

If you have any questions or need help with testing, please to write to
nikola.hardi@cern.ch (link sends e-mail).

# References

  1. [Slacker: Fast Distribution with Lazy Docker Containers](https://www.usenix.org/node/194431)
  2. https://github.com/cvmfs/docker-graphdriver
  3. [The Evolution of Global Scale Filesystems for Scientific Software Distribution](http://ieeexplore.ieee.org/document/7310920/?arnumber=7310920)
  4. https://gitlab.cern.ch/cloud-infrastructure/docker-volume-cvmfs/
  5. http://singularity.lbl.gov
