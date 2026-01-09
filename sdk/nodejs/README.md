# ConfigHub Node.js SDK

Node.js SDK for ConfigHub configuration management platform.

## Installation

```bash
npm install @confighub/sdk
# or
yarn add @confighub/sdk
```

## Quick Start

```typescript
import { ConfigHubClient } from '@confighub/sdk';

// Create client
const client = new ConfigHubClient({
  serverUrl: 'http://localhost:8080',
  accessKey: 'your-access-key',
  secretKey: 'your-secret-key',
  namespace: 'application',
  environment: 'dev',
});

// Get configuration
const config = await client.get('app-config');
console.log(`Config: ${config.name} (v${config.version})`);
console.log(config.content);

// Close client when done
client.close();
```

## Features

### Get Configuration

```typescript
// Get as Config object
const config = await client.get('app-config');

// Get as string
const content = await client.getString('app-config');

// Get and parse JSON
interface AppConfig {
  database: { host: string; port: number };
  redis: { host: string };
}
const appConfig = await client.getJSON<AppConfig>('app-config');

// Get with custom namespace/environment
const config = await client.getWithOptions('app-config', 'custom-ns', 'prod');
```

### Watch for Changes

```typescript
const client = new ConfigHubClient({
  serverUrl: 'http://localhost:8080',
  accessKey: 'your-access-key',
  secretKey: 'your-secret-key',
  onChange: (config) => {
    console.log(`Config changed: ${config.name} v${config.version}`);
    // Reload your application config here
  },
  onError: (error) => {
    console.error('Watch error:', error);
  },
});

// Start watching
await client.watch('app-config', 'db-config');

// Stop watching when done
client.stopWatch();
```

### Cache Management

```typescript
// Force refresh from server
const config = await client.refresh('app-config');

// Clear all cached configs
client.clearCache();

// Get cached version
const version = client.getCachedVersion('app-config');
```

## Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| serverUrl | string | (required) | ConfigHub server URL |
| accessKey | string | (required) | API access key |
| secretKey | string | (required) | API secret key |
| namespace | string | "application" | Default namespace |
| environment | string | "default" | Default environment |
| watchTimeout | number | 30 | Long-polling timeout (seconds) |
| onChange | function | - | Callback for config changes |
| onError | function | - | Callback for watch errors |

## Error Handling

```typescript
import { ConfigHubClient, ConfigHubError, Errors } from '@confighub/sdk';

try {
  const config = await client.get('app-config');
} catch (error) {
  if (error instanceof ConfigHubError) {
    switch (error.code) {
      case Errors.NOT_FOUND:
        console.log('Config not found');
        break;
      case Errors.UNAUTHORIZED:
        console.log('Invalid credentials');
        break;
      default:
        console.log('Error:', error.message);
    }
  }
}
```

## TypeScript Support

This SDK is written in TypeScript and includes type definitions.

```typescript
import { ConfigHubClient, Config, ClientOptions } from '@confighub/sdk';

const options: ClientOptions = {
  serverUrl: 'http://localhost:8080',
  accessKey: 'your-access-key',
  secretKey: 'your-secret-key',
};

const client = new ConfigHubClient(options);
const config: Config = await client.get('app-config');
```

## License

MIT License
