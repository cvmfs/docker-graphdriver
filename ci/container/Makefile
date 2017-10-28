# Builds the CernVM-FS graphdriver plugin container for Docker
# Requires a Debian based distribution with
# git, debootstrap, fakeroot, fakechroot

IMAGE_NAME = cvmfs-graphdriver-plugin
IMAGE_ARCH = amd64
IMAGE_DISTRO = stretch
IMAGE_VERSION = 0.5
IMAGE_TARBALL = $(IMAGE_NAME)-$(IMAGE_VERSION).tar.bz2
CVMFS_VERSION = 2.4.2
CVMFS_RELEASE = 1+debian9.1
CVMFS_CONFIG_VERSION = 1.1
CVMFS_CONFIG_RELEASE = 1
GRAPHDRIVER_AUFS_VERSION = 0.2
GRAPHDRIVER_OVERLAY2_VERSION = 0.2

CVMFS_BASE_URL = https://ecsft.cern.ch/dist/cvmfs/cvmfs-$(CVMFS_VERSION)
CVMFS_CONFIG_BASE_URL = https://ecsft.cern.ch/dist/cvmfs/cvmfs-config
GRAPHDRIVER_AUFS_BASE_URL = https://ecsft.cern.ch/dist/cvmfs/docker-graphdriver/aufs_cvmfs
GRAPHDRIVER_OVERLAY2_BASE_URL = https://ecsft.cern.ch/dist/cvmfs/docker-graphdriver/overlay2_cvmfs

GIT_SRC = ../src
CVMFS_SRC = $(CVMFS_BASE_URL)/cvmfs_$(CVMFS_VERSION)~$(CVMFS_RELEASE)_$(IMAGE_ARCH).deb
CVMFS_CONFIG_SRC = $(CVMFS_CONFIG_BASE_URL)/cvmfs-config-graphdriver_$(CVMFS_CONFIG_VERSION)-$(CVMFS_CONFIG_RELEASE)_all.deb
GRAPHDRIVER_AUFS_SRC = $(GRAPHDRIVER_AUFS_BASE_URL)/$(GRAPHDRIVER_AUFS_VERSION)/aufs_cvmfs
GRAPHDRIVER_OVERLAY2_SRC = $(GRAPHDRIVER_OVERLAY2_BASE_URL)/$(GRAPHDRIVER_OVERLAY2_VERSION)/overlay2_cvmfs

ROOTFS = $(IMAGE_NAME)-$(IMAGE_VERSION)/rootfs
MANIFEST = $(IMAGE_NAME)-$(IMAGE_VERSION)/config.json

all: $(IMAGE_TARBALL)

clean:
	rm -f $(IMAGE_TARBALL)
	rm -rf $(IMAGE_NAME)-$(IMAGE_VERSION)

$(IMAGE_TARBALL): $(ROOTFS)/CONTENTS $(MANIFEST)
	tar cfJ $@ $(IMAGE_NAME)-$(IMAGE_VERSION)
	rm -rf $(IMAGE_NAME)-$(IMAGE_VERSION)

$(ROOTFS)/CONTENTS: $(ROOTFS)/usr/bin/cvmfs2 \
  $(ROOTFS)/usr/local/bin/aufs_cvmfs \
  $(ROOTFS)/usr/local/bin/overlay2_cvmfs \
  $(ROOTFS)/usr/local/bin/entrypoint.sh
	cat /dev/null > $(ROOTFS)/.cvmfs.deb
	cat /dev/null > $(ROOTFS)/.cvmfs-config.deb
	fakechroot fakeroot chroot $(ROOTFS) apt-get clean
	echo "Image version:            $(IMAGE_VERSION)" >> $@
	echo "Git revision:             $(shell cd $(GIT_SRC) && git rev-parse HEAD)" >> $@
	echo "Debian distribution:      $(IMAGE_DISTRO)" >> $@
	echo "CernVM-FS client:         $(CVMFS_CONFIG_SRC)" >> $@
	echo "CernVM-FS version:        $(CVMFS_SRC)" >> $@
	echo "AUFS graph driver:        $(GRAPHDRIVER_AUFS_SRC)" >> $@
	echo "Overlay2 graph driver:    $(GRAPHDRIVER_OVERLAY2_SRC)" >> $@

$(ROOTFS)/.done_minbase:
	mkdir $(IMAGE_NAME)-$(IMAGE_VERSION)
	fakechroot fakeroot debootstrap --variant=minbase $(IMAGE_DISTRO) $(ROOTFS)
	touch $(ROOTFS)/.done_minbase

$(MANIFEST): $(ROOTFS)/.done_minbase
	cp $(GIT_SRC)/data/plugin-config.json $@

$(ROOTFS)/usr/local/bin/entrypoint.sh: $(ROOTFS)/.done_minbase
	cp $(GIT_SRC)/plugins/entrypoint.sh $@

$(ROOTFS)/.cvmfs-config.deb: $(ROOTFS)/.done_minbase
	curl -o $@ $(CVMFS_CONFIG_SRC)

$(ROOTFS)/.cvmfs.deb: $(ROOTFS)/.done_minbase
	curl -o $@ $(CVMFS_SRC)

$(ROOTFS)/usr/bin/cvmfs2: | $(ROOTFS)/.cvmfs-config.deb $(ROOTFS)/.cvmfs.deb
	fakechroot fakeroot chroot $(ROOTFS) dpkg -i .cvmfs-config.deb
	fakechroot fakeroot chroot $(ROOTFS) dpkg -i .cvmfs.deb || true
	fakechroot fakeroot chroot $(ROOTFS) apt-get install -fy
	ln -sf /bin/kmod $(ROOTFS)/sbin/modprobe
	ln -sf /lib/x86_64-linux-gnu/ld-2.24.so $(ROOTFS)/lib64/ld-linux-x86-64.so.2
	rm -f $(ROOTFS)/var/run
	ln -s /run $(ROOTFS)/var/run
	mkdir $(ROOTFS)/run/cvmfs
	chmod 0755 $(ROOTFS)/bin/mount
	chmod 0755 $(ROOTFS)/bin/umount
	chmod 0755 $(ROOTFS)/bin/fusermount

$(ROOTFS)/usr/local/bin/aufs_cvmfs: $(ROOTFS)/.done_minbase
	curl -o $@ $(GRAPHDRIVER_AUFS_SRC)
	chmod +x $@

$(ROOTFS)/usr/local/bin/overlay2_cvmfs: $(ROOTFS)/.done_minbase
	curl -o $@ $(GRAPHDRIVER_OVERLAY2_SRC)
	chmod +x $@

