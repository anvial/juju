(hook-command-juju-log)=
# `juju-log`
## Summary
Writes a message to Juju logs.

## Usage
``` juju-log [options] <message>```

### Options
| Flag | Default | Usage |
| --- | --- | --- |
| `--debug` | false | log at debug level |
| `--format` |  | deprecated format flag |
| `-l`, `--log-level` | INFO | Send log message at the given level |

## Examples

    juju-log -l 'WARN' Something has transpired