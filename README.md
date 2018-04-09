# LiveRecord Server 

[![Go Report Card](https://goreportcard.com/badge/github.com/liverecord/server)](https://goreportcard.com/report/github.com/liverecord/server)
[![Build Status](https://travis-ci.org/liverecord/server.svg?branch=master)](https://travis-ci.org/liverecord/server)

This is socket server for LiveRecord communication platform.

## Configure

Add to `.env`
```
MYSQL_DSN=root:123@tcp(127.0.0.1:3306)/liveRecord?charset=utf8&parseTime=True
DOCUMENT_ROOT=/Users/zoonman/Projects/www/liverecord/client/dist
LISTEN_ADDR=:8000
DEBUG=true
```

## Assemble and run

Install [make](https://www.gnu.org/software/make/manual/make.html) and execute:

```bash
make it work
```


Running inside Docker container

```bash



```