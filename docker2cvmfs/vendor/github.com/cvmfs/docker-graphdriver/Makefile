repo=$(shell pwd | xargs basename)

.PHONY: shell
shell:
	mkdir -p /tmp/dind > /dev/null
	docker run -u $(shell whoami) --privileged -it -v/tmp/dind:/tmp -v$(shell pwd):/work/$(repo) atlantic777/cvmfs-dev dumb-init tmux

unit_test:
	mkdir -p /tmp/dind > /dev/null
	docker run --privileged -it -v/tmp/dind:/tmp -v$(shell pwd):/work/$(repo) golang bash
