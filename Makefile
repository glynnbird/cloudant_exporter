# This outputs:
# - Tag name if we're exactly on a tag
# - Raw git SHA if there's no tag in the history
# - If there's a tag in the history somewhere:
#    <tag>-<commits since tag>-SHA
VERSION=$(shell git describe --always)

build:
	go build -ldflags="-X 'main.Version=$(VERSION)'" ./cmd/cloudant_exporter/
