#   kube-secrets
#   Copyright 2017 Jolene Engo <dev.toaster@gmail.com>

#   This program is free software: you can redistribute it and/or modify
#   it under the terms of the GNU General Public License as published by
#   the Free Software Foundation, either version 3 of the License, or
#   (at your option) any later version.

#   This program is distributed in the hope that it will be useful,
#   but WITHOUT ANY WARRANTY; without even the implied warranty of
#   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
#   GNU General Public License for more details.

#   You should have received a copy of the GNU General Public License
#   along with this program.  If not, see <http://www.gnu.org/licenses/>.


all: depends

build:
	mkdir -p /go/build/darwin/amd64
	mkdir -p /go/build/linux/386
	mkdir -p /go/build/linux/amd64
	GOOS=darwin GOARCH=amd64 go build -o /go/build/darwin/amd64/kube-secrets kube-secrets.go lib.go
	GOOS=linux GOARCH=386 go build -o /go/build/linux/386/kube-secrets kube-secrets.go lib.go
	GOOS=linux GOARCH=amd64 go build -o /go/build/linux/amd64/kube-secrets kube-secrets.go lib.go

clean:
	docker-compose down -v --rmi local

depends:
	docker-compose up --build -d --remove-orphans --force-recreate

shell:
	docker-compose exec kube-secrets /bash.sh

test:
	docker-compose exec -T kube-secrets make test_docker

test_docker:
	cp -r test_data/ /tmp
	go test -v -covermode=count -coverprofile=coverage.out | /go/bin/go-junit-report > test-report.xml
	/go/bin/goveralls -coverprofile=coverage.out -service=travis-ci -repotoken ${COVERALLS_TOKEN}
