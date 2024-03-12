# LeetCode bot
Copyright © 2077, Boar D'White foundation. All rights reserved.

## Repo setup
```shell
$ curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.56.2
$ ln -sf $(pwd)/pre-commit .git/hooks/pre-commit
```

## Run bot
```shell
$ cp .env.example .env
$ docker-compose up --build -d
```
