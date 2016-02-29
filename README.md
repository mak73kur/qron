# Qron

Qron is a simple scheduler for message queues.

# Install

Download the latest release binary.

Or using go tool:

```
$ go get github.com/mak73kur/qron/cmd/qron
```

This will install the **qron** binary to your $GOPATH/bin directory.

## Usage

Start the program:

```
$ ./qron path_to_config.ext
```

Config has two main sections:

- **loader** which tells qron where should it look for a job schedule called *qrontab*.
- **writer** that will decide where qron will publish messages.

Specific config properties will depend on the chosen loader and writer types.

Thanks to spf13/viper config file supports different formats: JSON, TOML, YAML.

Program can be started without a config argument, then qron will look at the default location at /etc/qron.yml.

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

## Qrontab loaders

Tab can be loaded from one of the following sources.

### Inline source

Store the tab directly in the same config.

```
loader:
    type: inline
    tab: |
        * * * * * every minute
        */2 * * * * every two minutes
```

### File source

The file will be read once on program start. Sending SIGHUP will trigger file reread
without a restart (this is also true about other loaders except for inline).

```
loader:
    type: file
    path: /tmp/qrontab
```

### Redis source

Tab is stored as a single string value in Redis DB.

Separate program thread will call ```GET <key>``` every minute
to update the tab on any changes.

Note: if the value gets cleared for some reason - qron will assume the tab is empty
and will continue to run with nothing to publish.

```
loader:
    type: redis
    url: localhost:6379
    db: 0
    key: qrontab
```

## Writers

If a job matches current time then it will be published by qron writer.
At this moment supported writers include following.

### Log writer

For debug purposes. Writes message directly to the program output.

```
writer.type: log
```

### AMQP writer

AMQP writes messages to RabbitMQ or any other protocol implementations.

```
writer:
    type: amqp
    url: amqp://localhost:5672
    exchange: ""
    routing_key: qron
```

## TODO

- [ ] verbose mode
- [ ] proper godoc
- [ ] tests for the base package
- [ ] custom msg properties e.g. *ttl*
