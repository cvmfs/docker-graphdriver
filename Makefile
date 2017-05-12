repo=$(shell pwd | xargs basename)

shell:
	mkdir -p /tmp/dind > /dev/null
	docker run -u $(shell whoami)  --privileged -it -v/tmp/dind:/tmp -v$(shell pwd):/work/$(repo) nhardi/docker_graphdriver_plugins:dev dumb-init tmux

unit_test:
	mkdir -p /tmp/dind > /dev/null
	docker run --privileged -it -v/tmp/dind:/tmp -v$(shell pwd):/work/$(repo) golang bash
