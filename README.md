# Echo Debug Monitor

[![test](https://github.com/kohkimakimoto/echo-debugmonitor/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/kohkimakimoto/echo-debugmonitor/actions/workflows/test.yml?query=branch%3Amain)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/kohkimakimoto/echo-debugmonitor/blob/main/LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/kohkimakimoto/echo-debugmonitor/v5.svg)](https://pkg.go.dev/github.com/kohkimakimoto/echo-debugmonitor/v5)

**This repository is still under active development. Documentation is incomplete and breaking changes may occur.**

A debugging and monitoring dashboard for Go applications using the [Echo](https://echo.labstack.com/) web framework.
Provides real-time visibility into application behavior through multiple specialized monitors.

This `v5` release line supports Echo v5. The major version of Echo Debug Monitor matches the supported Echo major version.
For Echo v4, use the [`v4`](https://github.com/kohkimakimoto/echo-debugmonitor/tree/v4) branch (`github.com/kohkimakimoto/echo-debugmonitor/v4`).

![](https://raw.githubusercontent.com/kohkimakimoto/echo-debugmonitor/main/images/screenshot.png)

> [!WARNING]
> This tool is intended for debugging and development purposes only. **Do not activate it in production environments** as it may expose sensitive information and impact performance.

## Getting Started

### Installation

```bash
go get github.com/kohkimakimoto/echo-debugmonitor/v5
```

### Basic Usage

Here's a simple example to get started with Echo Debug Monitor:

```go
package main

import (
    "github.com/labstack/echo/v5"
    debugmonitor "github.com/kohkimakimoto/echo-debugmonitor/v5"
    "github.com/kohkimakimoto/echo-debugmonitor/v5/monitors"
)

func main() {
    e := echo.New()

    // Create the debug monitor manager
    m := debugmonitor.New()

    // Add requests monitor
    requestsMonitor, requestsMonitorMiddleware := monitors.NewRequestsMonitor(&monitors.RequestsMonitorConfig{
        Skipper: func(c *echo.Context) bool {
            return c.Path() == "/monitor" // Skip monitoring the monitor endpoint itself
        },
    })
    e.Use(requestsMonitorMiddleware)
    m.AddMonitor(requestsMonitor)

    // Add logs monitor (wraps *slog.Logger)
    logsMonitor, wrappedLogger := monitors.NewLogsMonitor(monitors.LogsMonitorConfig{
        Logger: e.Logger,
    })
    e.Logger = wrappedLogger
    m.AddMonitor(logsMonitor)

    // Register the dashboard route
    e.GET("/monitor", m.Handler())

    // Your application routes
    e.GET("/", func(c *echo.Context) error {
        e.Logger.Info("Home page accessed")
        return c.String(200, "Hello, World!")
    })

    e.Start(":8080")
}
```

Then access the monitoring dashboard at `http://localhost:8080/monitor`.

## Monitors

Monitors are the core units in Echo Debug Monitor. Each monitor tracks a specific aspect of your application and displays it in the dashboard.

A monitor is an independent component that:
- Captures specific types of data (requests, logs, errors, etc.)
- Stores captured data in an in-memory buffer.
- Provides a real-time view in the web interface

Each monitor operates independently and can be added or removed.
You can also implement custom monitors for your specific needs.

## Built-in Monitors

Echo Debug Monitor includes several ready-to-use monitors in the `github.com/kohkimakimoto/echo-debugmonitor/v5/monitors` package:

- **Requests Monitor**: Tracks incoming HTTP requests, response statuses, latencies, etc.
- **Logs Monitor**: Captures application logs via `log/slog` and displays them in real-time.
- **Writer Monitor**: Monitors output written to `io.Writer` interfaces.
- **Errors Monitor**: Records application errors and stack traces.
- **Queries Monitor**: Tracks database queries.

## Implementing Custom Monitors

WIP

## License

The MIT License (MIT)

## Author

Kohki Makimoto <kohki.makimoto@gmail.com>
