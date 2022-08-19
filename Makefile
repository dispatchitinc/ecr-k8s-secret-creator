VERSION ?= 0.1.0
LDFLAGS = -ldflags "-X main.VERSION=$(VERSION) -X main.COMMIT=$(shell git rev-parse --short HEAD) -X main.BRANCH=$(shell git branch | grep \* | cut -d ' ' -f2)"
IMAGE_REPO = albanda/ecr-k8s-secret-creator

setup:
	pre-commit install

publish:
	KO_DOCKER_REPO=$(IMAGE_REPO) ko build --tags $(VERSION) --bare ./cmd/ecr-k8s-secret-creator/

test:
	go test -v ./...

coverage:
	go test -cpu=1 -v ./... -failfast -coverprofile=coverage.txt -covermode=count

lint:
	GO111MODULE=off go get -u golang.org/x/lint/golint
	golint -set_exit_status ./...
