#!/bin/bash
RED="\e[31m"
GREEN="\e[32m"
YELLOW="\e[33m"
BLUE="\e[34m"
NC="\e[0m"  # No Color

info(){
    printf "${BLUE}$(date -Iseconds) [INFO]$1${NC}\n"
}

warn(){
    printf "${YELLOW}$(date -Iseconds) [WARN]$1${NC}\n"
}

fatal(){
    printf "${RED}$(date -Iseconds) [ERROR]$1${NC}\n"
    exit 1
}

runExitOnError() {
  "$@"  # executes the command exactly as passed
  rc=$?

  if [ $rc -ne 0 ]; then
    fatal "Command failed (exit=$rc): $*"
    return $rc
  fi

  info "Command succeeded: $*"
}

flightCheck(){
    for cmd in uv pip3 python3 ollama; do
        if ! command -v "$cmd" >/dev/null 2>&1; then
            fatal "❌ $cmd is missing"
        else
            info "✅ $cmd is installed"
        fi
    done
}

init(){
    flightCheck
    runExitOnError uv venv
    runExitOnError source .venv/bin/activate
    runExitOnError uv pip install leann
    runExitOnError ollama pull llama3.2:1b
}

## RUN ###
init
