# Builds the CernVM-FS graphdriver plugin container for Docker
# Requires a Debian based distribution with
# make, git, curl, gcc, xz-utils, libfuse

IMAGE_NAME = cvmfs-graphdriver-plugin
IMAGE_ARCH = amd64
IMAGE_VERSION = 0.12
IMAGE_TARBALL = $(IMAGE_NAME)-$(IMAGE_VERSION).tar.bz2
BB_VERSION = 1.23.2
CVMFS_VERSION = 2.4.2
CVMFS_RELEASE = 1+debian9.1
CVMFS_CONFIG_VERSION = 1.1
CVMFS_CONFIG_RELEASE = 1
GRAPHDRIVER_AUFS_VERSION = 0.6
GRAPHDRIVER_OVERLAY2_VERSION = 0.6

BB_BASE_URL = http://ecsft.cern.ch/dist/cernvm
CVMFS_BASE_URL = https://ecsft.cern.ch/dist/cvmfs/cvmfs-$(CVMFS_VERSION)
CVMFS_CONFIG_BASE_URL = https://ecsft.cern.ch/dist/cvmfs/cvmfs-config
GRAPHDRIVER_AUFS_BASE_URL = https://ecsft.cern.ch/dist/cvmfs/docker-graphdriver/aufs_cvmfs
GRAPHDRIVER_OVERLAY2_BASE_URL = https://ecsft.cern.ch/dist/cvmfs/docker-graphdriver/overlay2_cvmfs

GIT_SRC = ../src
BB_SRC = $(BB_BASE_URL)/busybox-$(BB_VERSION).tar.gz
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
  $(ROOTFS)/usr/bin/aufs_cvmfs \
  $(ROOTFS)/usr/bin/overlay2_cvmfs \
  $(ROOTFS)/usr/bin/entrypoint.sh
	cat /dev/null > $(ROOTFS)/.cvmfs.deb
	cat /dev/null > $(ROOTFS)/.cvmfs-config.deb
	cat /dev/null > $(ROOTFS)/.busybox-$(BB_VERSION).tar.gz
	rm -rf $(ROOTFS)/.busybox-$(BB_VERSION)/*
	echo "Image version:            $(IMAGE_VERSION)" >> $@
	echo "Git revision:             $(shell cd $(GIT_SRC) && git rev-parse HEAD)" >> $@
	echo "Busybox version:          $(BB_VERSION)" >> $@
	echo "CernVM-FS client:         $(CVMFS_CONFIG_SRC)" >> $@
	echo "CernVM-FS version:        $(CVMFS_SRC)" >> $@
	echo "AUFS graph driver:        $(GRAPHDRIVER_AUFS_SRC)" >> $@
	echo "Overlay2 graph driver:    $(GRAPHDRIVER_OVERLAY2_SRC)" >> $@

$(ROOTFS):
	mkdir -p $@

$(ROOTFS)/.busybox-$(BB_VERSION).tar.gz: | $(ROOTFS)
	curl -o $@ $(BB_SRC)

$(ROOTFS)/.busybox-$(BB_VERSION): $(ROOTFS)/.busybox-$(BB_VERSION).tar.gz
	cd $(ROOTFS) && tar xvf .busybox-$(BB_VERSION).tar.gz
	mv $(ROOTFS)/busybox-$(BB_VERSION) $(ROOTFS)/.busybox-$(BB_VERSION)

$(ROOTFS)/busybox: | $(ROOTFS)/.busybox-$(BB_VERSION)
	cp $(GIT_SRC)/data/busybox-config $(ROOTFS)/.busybox-$(BB_VERSION)/.config
	$(MAKE) -j $(shell nproc) -C $(ROOTFS)/.busybox-$(BB_VERSION)
	mv $(ROOTFS)/.busybox-$(BB_VERSION)/busybox $@

$(ROOTFS)/.done_minbase: $(ROOTFS)/busybox
	mkdir -p $(ROOTFS)/bin
	mkdir -p $(ROOTFS)/cvmfs_ext_config
	mkdir -p $(ROOTFS)/minio_ext_config
	mkdir -p $(ROOTFS)/dev
	mkdir -p $(ROOTFS)/lib/modules
	mkdir -p $(ROOTFS)/mnt
	mkdir -p $(ROOTFS)/proc
	mkdir -p $(ROOTFS)/tmp
	mkdir -p $(ROOTFS)/usr/bin
	mkdir -p $(ROOTFS)/var/lib/cvmfs
	mkdir -p $(ROOTFS)/var/log
	mkdir -p $(ROOTFS)/var/run/cvmfs
	ln -fs lib $(ROOTFS)/lib64
	for CMD in $(shell $(ROOTFS)/busybox --list | sort -u); do \
	    ln -fvs /busybox $(ROOTFS)/bin/$$CMD; \
	  done
	touch $(ROOTFS)/.done_minbase

$(MANIFEST): $(ROOTFS)/.done_minbase
	cp $(GIT_SRC)/data/plugin-config.json $@

$(ROOTFS)/usr/bin/entrypoint.sh: $(ROOTFS)/.done_minbase
	cp $(GIT_SRC)/plugins/entrypoint.sh $@

$(ROOTFS)/.cvmfs-config.deb: | $(ROOTFS)
	curl -o $@ $(CVMFS_CONFIG_SRC)

$(ROOTFS)/.cvmfs.deb: | $(ROOTFS)
	curl -o $@ $(CVMFS_SRC)

$(ROOTFS)/usr/bin/cvmfs2: $(ROOTFS)/.done_minbase | $(ROOTFS)/.cvmfs-config.deb $(ROOTFS)/.cvmfs.deb
	cd $(ROOTFS) && ar p .cvmfs-config.deb data.tar.gz | tar zx
	cd $(ROOTFS) && ar p .cvmfs.deb data.tar.xz | tar Jx
	rm -rf $(ROOTFS)/etc/bash_completion.d
	rm -rf $(ROOTFS)/usr/share
	rm -f $(ROOTFS)/usr/bin/cvmfs_talk \
	  $(ROOTFS)/usr/bin/cvmfs_config \
	  $(ROOTFS)/sbin/mount.cvmfs \
	  $(ROOTFS)/etc/auto.cvmfs \
	  $(ROOTFS)/usr/libexec/auto.cvmfs
	$(GIT_SRC)/ci/container/libs.sh $(ROOTFS)
	ldconfig -r $(ROOTFS)

$(ROOTFS)/usr/bin/aufs_cvmfs: $(ROOTFS)/.done_minbase
	curl -o $@ $(GRAPHDRIVER_AUFS_SRC)
	chmod +x $@

$(ROOTFS)/usr/bin/overlay2_cvmfs: $(ROOTFS)/.done_minbase
	curl -o $@ $(GRAPHDRIVER_OVERLAY2_SRC)
	chmod +x $@

