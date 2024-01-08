# Wikimedia Enterprise Logger

This package provides a wrapper around the underlying logging library that allows to log messages of different severity levels with additional fields.

## Usage

The package exports four logging functions for each severity level: `Debug()`, `Info()`, `Warn()`, and `Error()`. Each function takes a message string and zero or more fields of type Field, which can be used to provide additional context to the log message.

The package also provides a `Sync()` function that flushes any buffered log entries. Applications should take care to call `Sync()` before exiting.

The package exports four interfaces, `InfoLogger`, `WarnLogger`, `DebugLogger`, and `ErrorLogger`, each providing a method to log messages with the corresponding severity level. The package also exports a Logger interface that encompasses all the severity levels and the `Sync()` method.

## Constants

The package exports the following log level constants:

- `LevelInfo`: info severity level
- `LevelWarn`: warn severity level
- `LevelDebug`: debug severity level
- `LevelError`: error severity level

## Fields

The `Field` type that is used to pass additional values to the log functions. The package provides a `Any()` function that creates a new Field that associates a key with an arbitrary value. It takes a string `key` and an `interface{}` value, and returns a `Field`.

## Initialization

The package initializes the logger during package initialization by calling the `New()` function. The function reads the `LOG_LEVEL` environment variable and sets the log level accordingly. The default log level is `info`. The logger also reports the line number of the calling function.

## Example

This example logs the message `"Hello, world!"` with the additional context `foo=bar` at the info severity level.

```go
package main

import "my-package/submodules/log"

func main() {
  log.Info("Hello, world!", log.Any("foo", "bar"))
  log.Error(errors.New("Hello world!"), log.Tip("Bark!"))
}
```
