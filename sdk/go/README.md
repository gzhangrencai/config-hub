# ConfigHub Go SDK

Go SDK for ConfigHub configuration management platform.

## Installation

```bash
go get github.com/confighub/sdk-go/confighub
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/confighub/sdk-go/confighub"
)

func main() {
    // Create client
    client, err := confighub.NewClient(&confighub.ClientOptions{
        ServerURL:   "http://localhost:8080",
        AccessKey:   "your-access-key",
        SecretKey:   "your-secret-key",
        Namespace:   "application",
        Environment: "dev",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Get configuration
    config, err := client.Get(context.Background(), "app-config")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Config: %s (v%d)\n", config.Name, config.Version)
    fmt.Println(config.Content)
}
```

## Features

### Get Configuration

```go
// Get as Config struct
config, err := client.Get(ctx, "app-config")

// Get as string
content, err := client.GetString(ctx, "app-config")

// Get and unmarshal JSON
var appConfig AppConfig
err := client.GetJSON(ctx, "app-config", &appConfig)

// Get with custom namespace/environment
config, err := client.GetWithOptions(ctx, "app-config", "custom-ns", "prod")
```

### Watch for Changes

```go
client, _ := confighub.NewClient(&confighub.ClientOptions{
    ServerURL:   "http://localhost:8080",
    AccessKey:   "your-access-key",
    SecretKey:   "your-secret-key",
    OnChange: func(config *confighub.Config) {
        fmt.Printf("Config changed: %s v%d\n", config.Name, config.Version)
        // Reload your application config here
    },
    OnError: func(err error) {
        log.Printf("Watch error: %v", err)
    },
})

// Start watching
err := client.Watch(ctx, "app-config", "db-config")
if err != nil {
    log.Fatal(err)
}

// Stop watching when done
defer client.StopWatch()
```

### Cache Management

```go
// Force refresh from server
config, err := client.Refresh(ctx, "app-config")

// Clear all cached configs
client.ClearCache()

// Get cached version
version := client.GetCachedVersion("app-config")
```

## Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| ServerURL | string | (required) | ConfigHub server URL |
| AccessKey | string | (required) | API access key |
| SecretKey | string | (required) | API secret key |
| Namespace | string | "application" | Default namespace |
| Environment | string | "default" | Default environment |
| WatchTimeout | int | 30 | Long-polling timeout (seconds) |
| HTTPClient | *http.Client | nil | Custom HTTP client |
| OnChange | func(*Config) | nil | Callback for config changes |
| OnError | func(error) | nil | Callback for watch errors |

## Error Handling

```go
config, err := client.Get(ctx, "app-config")
if err != nil {
    switch err {
    case confighub.ErrNotFound:
        log.Println("Config not found")
    case confighub.ErrUnauthorized:
        log.Println("Invalid credentials")
    default:
        log.Printf("Error: %v", err)
    }
}
```

## License

MIT License
