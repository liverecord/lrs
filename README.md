# LiveRecord Server 

[![Go Report Card](https://goreportcard.com/badge/github.com/liverecord/lrs)](https://goreportcard.com/report/github.com/liverecord/lrs)
[![License](https://img.shields.io/github/license/liverecord/lrs.svg)](https://github.com/liverecord/lrs/blob/master/LICENSE)
[![Build Status](https://travis-ci.org/liverecord/lrs.svg?branch=master)](https://travis-ci.org/liverecord/lrs)
[![Github All Releases](https://img.shields.io/github/release/liverecord/lrs.svg)](https://github.com/liverecord/lrs/releases/latest)
[![Github All Releases](https://img.shields.io/github/downloads/liverecord/lrs/total.svg)](https://github.com/liverecord/lrs/releases/)


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
