BEAT_NAME=snifferbeat
BEAT_PATH=github.com/gitaiqaq/snifferbeat
BEAT_GOPATH=$(firstword $(subst :, ,${GOPATH}))
BEAT_URL=https://${BEAT_PATH}
SYSTEM_TESTS=false
TEST_ENVIRONMENT=false
ES_BEATS?=./vendor/github.com/elastic/beats
GOPACKAGES=$(shell glide novendor)
PREFIX?=.
NOTICE_FILE=NOTICE

# Path to the libbeat Makefile
-include $(ES_BEATS)/libbeat/scripts/Makefile

# Initial beat setup
.PHONY: setup
setup: copy-vendor
	make update

# Copy beats into vendor directory
.PHONY: copy-vendor
copy-vendor:
	mkdir -p vendor/github.com/elastic/
	cp -R ${BEAT_GOPATH}/src/github.com/elastic/beats vendor/github.com/elastic/
	rm -rf vendor/github.com/elastic/beats/.git

.PHONY: git-init
git-init:
	git init
	git add README.md CONTRIBUTING.md
	git commit -m "Initial commit"
	git add LICENSE
	git commit -m "Add the LICENSE"
	git add .gitignore
	git commit -m "Add git settings"
	git add .
	git reset -- .travis.yml
	git commit -m "Add snifferbeat"
	git add .travis.yml
	git commit -m "Add Travis CI"

# This is called by the beats packer before building starts
.PHONY: before-build
before-build:

# Collects all dependencies and then calls update
.PHONY: collect
collect:

deploy:
	mkdir -p ${GOPATH}/releases
	
	GOOS=windows GOARCH=386 go build -ldflags "-s -w"
	tar -cvjf ${GOPATH}/releases/snifferbeat_windows_x86.tar.bzip2 snifferbeat.exe snifferbeat.yml

	go clean

	GOOS=windows GOARCH=amd64 go build -ldflags "-s -w"
	tar -cvjf ${GOPATH}/releases/snifferbeat_windows_x64.tar.bzip2 snifferbeat.exe snifferbeat.yml

	go clean

	GOOS=linux GOARCH=386 go build -ldflags "-s -w"
	tar -cvjf ${GOPATH}/releases/snifferbeat_linux_x86.tar.bzip2 snifferbeat snifferbeat.yml

	go clean

	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w"
	tar -cvjf ${GOPATH}/releases/snifferbeat_linux_x64.tar.bzip2 snifferbeat snifferbeat.yml

	go clean