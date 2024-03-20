# LeetCode bot
Copyright © 2077, Boar D'White foundation. All rights reserved.

## Repo setup
```shell
$ curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.56.2
$ ln -sf $(pwd)/pre-commit .git/hooks/pre-commit
```

## Python scripts setup
```shell
# install pyenv
$ curl https://pyenv.run | bash
$ echo 'export PYENV_ROOT="$HOME/.pyenv"' >> ~/.zshrc
$ echo '[[ -d $PYENV_ROOT/bin ]] && export PATH="$PYENV_ROOT/bin:$PATH"' >> ~/.zshrc
$ echo 'eval "$(pyenv init -)"' >> ~/.zshrc
$ pyenv install $(cat .python-version)

# install uv
$ curl -LsSf https://astral.sh/uv/install.sh | sh
$ uv venv
$ . .venv/bin/activate
$ uv pip sync requirements.txt

# install latest chromedriver into ./chromedriver
# https://googlechromelabs.github.io/chrome-for-testing/
```

## Tests

# Unit tests
```
go test -race ./...
```

# E2E tests
```
go test --tags=e2e -race ./...
```

## Run bot
```shell
$ cp ./config/default_config.yaml config.yaml
$ # set tg.apiKey in config.yaml 
$ docker-compose up --build -d
```
