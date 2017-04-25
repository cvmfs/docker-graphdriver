shell:
	docker run -u nikola  --privileged -it -v/tmp:/tmp -v$(shell pwd):/work/$(shell pwd | xargs basename) atlantic777/dev tmux
