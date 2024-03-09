# LeetCode bot
Copyright © 2077, Boar D'White foundation. All rights reserved.

## Repo setup
```shell
$ go install honnef.co/go/tools/cmd/staticcheck@latest
$ go install golang.org/x/tools/cmd/goimports@latest
$ ln -sf $(pwd)/pre-commit .git/hooks/pre-commit
```

## Run bot
```shell
$ cp .env.example .env
$ docker-compose up --build -d
```
