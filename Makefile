SHELL:=/bin/bash

build:
	go build
	if [[ ! -d ${HOME}/.hunched-dog/ ]]; then mkdir ~/.hunched-dog/ ; fi
	if [[ ! -f ${HOME}/.hunched-dog/config.yml ]]; then cp ./debian/config.yml ~/.hunched-dog/config.yml ; fi

install:
	systemctl stop hunched-dog.service || true
	cp ./hunched-dog /usr/local/bin/hunched-dog
	cp ./debian/hunched-dog.service /etc/systemd/system/hunched-dog.service
	echo "Run 'systemctl start hunched-dog' to start service"
	systemctl start hunched-dog.service
