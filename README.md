# Sample ws server-watcher for charts dir.

Test project.

## Description

This server watch for file directory and push events to clients via ws protocol.

## API

### /

Serve static.

### /ws

Route for ws subscribing.

## Local launch

### Requirements

You need to have vgo installed.

```bash
go get -u golang.org/x/vgo
```

### Launch

1. Clone the repository.
2. Install dependencies, create the config file.
3. Create static files directory, based on config file.
4. Launch the project.

```bash
git clone https://github.com/lillilli/graphex.git && cd graphex
make setup && make config
mkdir -p shared
make run
```

### Docker

1. Clone the repository.
2. Make image (need some time).
3. Launch image (will be available on localhost:8081).

```bash
git clone https://github.com/lillilli/http_file_server.git && cd http_file_server
make run:image
```
