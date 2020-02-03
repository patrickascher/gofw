<h1>go-logger</h1>

go-logger provides an alternative to the standard library `log`.

Each `LogLevels` can have its own writer. That means `INFO` can be logged in a file, `ERROR` is mailed and everything else will logged in the console.
The logger is easy to extend. At the moment a console and file writer comes out of the box

The logger is registered as a singleton.
When the logger package is imported, the console Logger is registered automatically.

## Install

```go
go get github.com/fullhouse-productions/go-logger
```

## Usage

On a Logger the following functions are available
`Unspecified`, `Trace`, `Debug`, `Info`, `Warning`, `Error` and `Critical`.

```go
import "github.com/fullhouse-productions/go-logger"

console,err := logger.Get(CONSOLE)
if err != nil{
	//...
}

console.Debug("The logger is available for the project %#v","Gopher")
```

# LogLevels

The higher the log level, the more critical it is.

```go 
const (
	UNSPECIFIED Level = iota+1
	TRACE
	DEBUG
	INFO
	WARNING
	ERROR
	CRITICAL
)
``` 

## Interface
To create your own writer, you have to implement the Writer Interface

```go
type Writer interface {
	Write(LogEntry)
}
```

## Register
Register is used to register the log writer. 
Its useful to Register the logger in its init() function. So its automatically registered when you import the logger.
But you can also register it anywhere in the code.

```go
// init registers a email writer by importing your package
func init() {
	LoggerConfig := Config{Writer: &Email, LogLevel: UNSPECIFIED}
	logger.Register("email", &LoggerConfig})
}

// or just call it anywhere in you code
LoggerConfig := Config{Writer: &Email, LogLevel: UNSPECIFIED}
logger.Register("email", &LoggerConfig)
```

**Logger Config**

To ignore some useless logs and just start writing the log at a specific log-lvl you have to register your writer with more config information:

The following options are available

| Config                 | Describtion                                                                      |
|------------------------|----------------------------------------------------------------------------------|
| Writer                 | Writer which implements the `Writer` interface                                   |
| LogLevel               | The level the writer starts to log                                               |
| LevelUnspecifiedWriter `optional` | if empty the default Writer will be used. Otherwise define a Writer here. |
| TraceWriter `optional`            | if empty the default Writer will be used. Otherwise define a Writer here. |
| DebugWriter  `optional`           | if empty the default Writer will be used. Otherwise define a Writer here. |
| InfoWriter  `optional`           | if empty the default Writer will be used. Otherwise define a Writer here. |
| WarningWriter  `optional`        | if empty the default Writer will be used. Otherwise define a Writer here. |
| ErrorWriter  `optional`          | if empty the default Writer will be used. Otherwise define a Writer here. |
| CriticalWriter `optional`        | if empty the default Writer will be used. Otherwise define a Writer here. |

**Example**

```go
LoggerConfig := Config{Writer: &Email, LogLevel: INFO, CriticalWriter: &SMS}
logger.Register("special", &LoggerConfig)
// This would mean, just log everything which is of the LogLevel INFO or higher. 
// Use for everything the Email Writer but if its a Critical Log, use the SMS Writer.
```

!> Register can be called multiple times on the same name to reconfigure the Writer!

## Get
To get a log writer you have to call `logger.Get(CONSOLE)`. `CONSOLE` in that case is the log writer.

```go
l,err := logger.Get("email")
l.Debug("msg %v %v","arg-1","arg-2")
```

## LogEntry
`LogEntry` is getting posted to the writer. 
The following information is available.

| LogEntry                 | Describtion                                                                      |
|------------------------|----------------------------------------------------------------------------------|
| Level         | the `Level` of the log entry |
| Filename         | Filename in which the log got called. |
| Line         | Filename - Line  in which the log got called. |
| Timestamp         | Timestamp in which the log got called |
| Message         | The actual Message |


## Format
It is the job of your writer to format the message as you need it. There is a helper function `DefaultLoggingFormat` available which could be used in your writer.

It Outputs the message like this:
```go
return fmt.Sprintf("%s %s %s:%d %s", ts, e.Level.String(), filename, e.Line, e.Message)
```

# Console Logger

This logger is registered by default. To reconfigure it or access you can use the constant `logger.CONSOLE`
The console logger can get configured.

```go
//it's registered by default already
writer := ConsoleLogger(&ConsoleOptions{Color: true})
DefaultConfig := Config{Writer: writer, LogLevel: UNSPECIFIED}
Register(CONSOLE, DefaultConfig)
```

# File Logger

The file logger is writing the log entry in his own `goroutine` for better performance.
It also locks the file when its writing and unlocks it again when its finished.

To register it or reconfigure it you can use the constant `logger.FILE`

```go
writer := FileLogger(&FileOptions{File: "path/to/file.log"})
Config := Config{Writer: writer, LogLevel: UNSPECIFIED}
Register(FILE, Config)
```

?> If the file does not exist, the file logger tries to create it!

!> File logger is writing the file through a goroutine. That meas if your program exits before the goroutine is finished, the log is not getting written because your main program is killing all go routines.

# Issues & Ideas

To report Issues or to improve this package, please use the github issue board or send a pull request.

https://github.com/fullhouse-productions/go-logger/issues
