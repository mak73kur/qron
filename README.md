# Qron

Qron is a simple scheduler for message queues.

## Install

Download the latest release binary.

Or use go tool. This will install the qron binary to your $GOPATH/bin:

```Shell
$ go get github.com/mak73kur/qron/cmd/qron
```

## Usage

Run qron (in a verbose mode):

```Shell
$ qron -c /path/to/config.yml -v
```

Example config:

```YAML
reader:
    type: inline
    tab: |
        * * * * * [every minute]

writer:
    type: log
```

There are two main sections:

- **reader** which tells qron where should it look for a job schedule called *qrontab*.
- **writer** that decides where to qron will publish messages.

Specific properties depend on the chosen reader and writer types.

Thanks to spf13/viper, config file supports different formats: json, toml, yaml.

If path argument is empty, qron will try ./qron.yml by default.

## Qrontab

Each line is a new job.

Parameters should be separated by a single whitespace character.

First five are schedule parameters, sixth is a message body.

Message body can have whitespace or any other character, except newline.

```
┌───────────── min (0-59)
│ ┌────────────── hour (0-23)
│ │ ┌─────────────── day of month (1-31)
│ │ │ ┌──────────────── month (1-12)
│ │ │ │ ┌───────────────── day of week (0-6) (0 to 6 are Sunday to Saturday)
│ │ │ │ │
│ │ │ │ │
* * * * * message_body
```

Allowed parameter expressions:

- asterisk(*) — anything passes.
- comma(,) — ```0 5,17 * * 1``` — executes at 5AM and 5PM each Monday.
- slash(/) — ```0 */6 * * *``` — shortcut for ```0,6,12,18```.

Qron will skip any lines that are empty or start with the comment characters (# or //).

### Qrontab readers

Tab can be loaded from one of the following sources.

#### Inline source

Store the tab directly in the same config.

```YAML
reader:
    type: inline
    tab: |
        * * * * * every minute
        */2 * * * * every two minutes
```

#### File source

The file will be read once on program start. Sending SIGHUP will trigger file reread
without a restart (this is also true for redis reader).

```YAML
reader:
    type: file
    path: /tmp/qrontab
```

#### Redis source

Tab is stored as a single string value in Redis DB.

Separate program thread will call ```GET <key>``` every minute
to update the tab on any changes.

Note: if the value gets cleared for some reason - qron will assume the tab is empty
and will continue to run with nothing to publish.

```YAML
reader:
    type: redis
    url: localhost:6379
    key: qrontab
    # optional, 0 by default
    db: 0
    # optional, redis auth password
    auth: secret
```

### Writers

Every minute qron checks whether a job matches the current time.
If it does — qron writer will publish the job message.

#### Log writer

For debug purposes. Writes message directly to the program output.

```YAML
writer.type: log
```

#### AMQP writer

AMQP writes messages to RabbitMQ or any other protocol implementation.

```YAML
writer:
    type: amqp
    url: amqp://localhost:5672
    exchange: ""
    routing_key: qron
```

#### Redis writer

Push messages into the Redis list. Consumer side can use BRPOP or BLPOP to receive them.

```YAML
writer:
    type: redis
    url: localhost:6379
    key: qron
    # optional, 0 by default
    db: 1
    # optional, redis auth password
    auth: secret
    # optionaly use LPUSH, default is RPUSH
    left_push: true
```

## TODO

- [x] verbose mode
- [ ] proper godoc
- [ ] tests for the base package
- [ ] custom msg properties e.g. ttl, override key
