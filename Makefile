# Setup name variables for the package/tool
NAME := pkg
PKG := github.com/stevejuma/$(NAME)

CGO_ENABLED := 0

# Set any default go build tags.
BUILDTAGS :=

# Set our directory to build.
GO_CMD := ./...

include basic.mk

.PHONY: prebuild
prebuild: