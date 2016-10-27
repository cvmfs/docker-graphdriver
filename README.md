About
=====
This utility can help importing existing Docker images into cvmfs.

How to use
==========
`go run docker2cvmfs.go library/ubuntu`

This will download all layers for the given image in predefined location which
is currently `/tmp/layers`. Names of downloaded files are hashes (digests) of
layers. Files are itself just tar.gz archives which can be unpacked as usual.
