ROOT_DIR    = $(shell pwd)
NAMESPACE   = "default"
DEPLOY_NAME = "star"
DOCKER_NAME = "star"

include ./hack/hack-cli.mk
include ./hack/hack.mk