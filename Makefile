shell:
	mkdir -p /tmp/dind > /dev/null
	docker run -u $(shell whoami)  --privileged -it -v/tmp/dind:/tmp -v$(shell pwd):/work/$(shell pwd | xargs basename) atlantic777/dev dumb-init tmux
