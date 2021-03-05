#!/bin/bash

VERSION=v0.1.4

if [[ ! -d ${HOME}/.hunched-dog/ ]]; then
  mkdir ~/.hunched-dog/
fi

if [[ ! -f ${HOME}/.hunched-dog/config.yml ]]; then

  sudo tee ~/.hunched-dog/config.yml >/dev/null <<EOT
target: ${HOME}/hunched-dog-cloud
multicast: 224.0.0.0:45046

EOT
fi

curl -L -o ./hunched-dog https://github.com/goforbroke1006/hunched-dog/releases/download/${VERSION}/hunched-dog__linux_amd64

sudo systemctl stop hunched-dog.service || true

sudo cp ./hunched-dog /usr/local/bin/hunched-dog

sudo tee /etc/systemd/system/hunched-dog.service >/dev/null <<EOT
[Unit]
Description=hunched dog service
After=network.target
StartLimitIntervalSec=0

[Service]
Type=simple
Restart=always
RestartSec=1
User=${USER}
ExecStart=/usr/bin/env /usr/local/bin/hunched-dog

[Install]
WantedBy=multi-user.target

EOT

sudo systemctl start hunched-dog.service
sudo systemctl status hunched-dog.service
