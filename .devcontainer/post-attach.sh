#!/usr/bin/env bash

# Install Go tools
./deps.sh

# Install direnv hook
grep -qxF 'include "direnv"' /home/vscode/.bashrc || echo 'eval "$(direnv hook bash)"' >> /home/vscode/.bashrc