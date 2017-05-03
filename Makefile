shell:
	mkdir -p /tmp/dind > /dev/null
	docker run -u $(shell whoami)  --privileged -it -v/tmp/dind:/tmp -v$(shell pwd):/work/$(shell pwd | xargs basename) gitlab.cern.ch/nhardi/dev dumb-init tmux
