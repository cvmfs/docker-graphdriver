#!/bin/bash
# source: 
# https://www.digitalocean.com/community/tutorials/how-to-set-up-a-private-docker-registry-on-ubuntu-14-04

ifndef FQDN
$(error FQDN variable is not set. Set it like this make all FQDN=something.com)
endif

CA_NAME=devdockerCA
ROOT_KEY=$(CA_NAME).key
ROOT_CRT=$(CA_NAME).crt

SERVER_KEY=$(FQDN).key
SERVER_CSR=$(FQDN).csr
SERVER_CRT=$(FQDN).crt

DAYS_VALID=31

all: cert

cert: $(SERVER_CRT)

$(ROOT_KEY): 
	openssl genrsa -out $(ROOT_KEY) 2048

$(ROOT_CRT): $(ROOT_KEY)
	openssl req -x509 -new -nodes -key $(ROOT_KEY) -days $(DAYS_VALID) -out $(ROOT_CRT) -subj "/CN=cern.ch"

$(SERVER_KEY):
	openssl genrsa -out $(SERVER_KEY) 2048

$(SERVER_CSR): $(ROOT_KEY)
	openssl req -new -key $(SERVER_KEY) -out $(SERVER_CSR) -subj "/C=CH/CN=$(FQDN)"

$(SERVER_CRT): $(SERVER_KEY) $(SERVER_CSR) $(ROOT_KEY) $(ROOT_CRT)
	openssl x509 -req -in $(SERVER_CSR) -CA $(ROOT_CRT) -CAkey $(ROOT_KEY) -CAcreateserial -out $(SERVER_CRT) -days $(DAYS_VALID)

clean:
	rm -v *.srl *.crt *.key *.csr
