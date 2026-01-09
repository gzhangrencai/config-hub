import crypto from 'crypto';
import http from 'http';
import https from 'https';
import { URL } from 'url';

/**
 * Configuration item
 */
export interface Config {
  name: string;
  namespace: string;
  environment: string;
  version: number;
  content: string;
}

/**
 * Client options
 */
export interface ClientOptions {
  /** ConfigHub server URL (required) */
  serverUrl: string;
  /** API access key (required) */
  accessKey: string;
  /** API secret key (required) */
  secretKey: string;
  /** Default namespace (default: "application") */
  namespace?: string;
  /** Default environment (default: "default") */
  environment?: string;
  /** Long-polling timeout in seconds (default: 30) */
  watchTimeout?: number;
  /** Callback for config changes */
  onChange?: (config: Config) => void;
  /** Callback for watch errors */
  onError?: (error: Error) => void;
}

/**
 * ConfigHub SDK errors
 */
export class ConfigHubError extends Error {
  constructor(message: string, public code: string) {
    super(message);
    this.name = 'ConfigHubError';
  }
}

export const Errors = {
  NOT_FOUND: 'NOT_FOUND',
  UNAUTHORIZED: 'UNAUTHORIZED',
  INVALID_CONFIG: 'INVALID_CONFIG',
  WATCH_TIMEOUT: 'WATCH_TIMEOUT',
  CLIENT_CLOSED: 'CLIENT_CLOSED',
};

/**
 * ConfigHub client
 */
export class ConfigHubClient {
  private opts: Required<ClientOptions>;
  private cache: Map<string, Config> = new Map();
  private watching: boolean = false;
  private watchAbortControllers: Map<string, AbortController> = new Map();

  constructor(options: ClientOptions) {
    if (!options.serverUrl) {
      throw new Error('serverUrl is required');
    }
    if (!options.accessKey || !options.secretKey) {
      throw new Error('accessKey and secretKey are required');
    }

    this.opts = {
      serverUrl: options.serverUrl.replace(/\/$/, ''),
      accessKey: options.accessKey,
      secretKey: options.secretKey,
      namespace: options.namespace || 'application',
      environment: options.environment || 'default',
      watchTimeout: options.watchTimeout || 30,
      onChange: options.onChange || (() => {}),
      onError: options.onError || (() => {}),
    };
  }


  /**
   * Get configuration by name
   */
  async get(name: string): Promise<Config> {
    return this.getWithOptions(name, this.opts.namespace, this.opts.environment);
  }

  /**
   * Get configuration with custom namespace and environment
   */
  async getWithOptions(name: string, namespace: string, environment: string): Promise<Config> {
    const cacheKey = this.cacheKey(name, namespace, environment);
    
    // Check cache first
    const cached = this.cache.get(cacheKey);
    if (cached) {
      return cached;
    }

    // Fetch from server
    const config = await this.fetchConfig(name, namespace, environment);
    this.cache.set(cacheKey, config);
    return config;
  }

  /**
   * Get configuration content as string
   */
  async getString(name: string): Promise<string> {
    const config = await this.get(name);
    return config.content;
  }

  /**
   * Get configuration content as parsed JSON
   */
  async getJSON<T = unknown>(name: string): Promise<T> {
    const config = await this.get(name);
    return JSON.parse(config.content);
  }

  /**
   * Force refresh configuration from server
   */
  async refresh(name: string): Promise<Config> {
    const cacheKey = this.cacheKey(name, this.opts.namespace, this.opts.environment);
    const config = await this.fetchConfig(name, this.opts.namespace, this.opts.environment);
    this.cache.set(cacheKey, config);
    return config;
  }

  /**
   * Fetch configuration from server
   */
  private async fetchConfig(name: string, namespace: string, environment: string): Promise<Config> {
    const url = new URL(`${this.opts.serverUrl}/api/v1/config`);
    url.searchParams.set('name', name);
    if (namespace) url.searchParams.set('namespace', namespace);
    if (environment) url.searchParams.set('env', environment);

    const response = await this.request('GET', url.toString());
    
    if (response.statusCode === 404) {
      throw new ConfigHubError('Config not found', Errors.NOT_FOUND);
    }
    if (response.statusCode === 401) {
      throw new ConfigHubError('Unauthorized', Errors.UNAUTHORIZED);
    }
    if (response.statusCode !== 200) {
      throw new ConfigHubError(`Server error: ${response.body}`, 'SERVER_ERROR');
    }

    return JSON.parse(response.body) as Config;
  }

  /**
   * Make HTTP request with authentication
   */
  private request(method: string, urlStr: string): Promise<{ statusCode: number; body: string }> {
    return new Promise((resolve, reject) => {
      const url = new URL(urlStr);
      const timestamp = Math.floor(Date.now() / 1000).toString();
      
      // Create signature
      const pathWithQuery = url.pathname + (url.search || '');
      const message = timestamp + method + pathWithQuery;
      const signature = crypto
        .createHmac('sha256', this.opts.secretKey)
        .update(message)
        .digest('hex');

      const options = {
        method,
        hostname: url.hostname,
        port: url.port || (url.protocol === 'https:' ? 443 : 80),
        path: pathWithQuery,
        headers: {
          'X-Access-Key': this.opts.accessKey,
          'X-Timestamp': timestamp,
          'X-Signature': signature,
        },
        timeout: (this.opts.watchTimeout + 10) * 1000,
      };

      const client = url.protocol === 'https:' ? https : http;
      const req = client.request(options, (res: http.IncomingMessage) => {
        let body = '';
        res.on('data', (chunk: Buffer | string) => { body += chunk; });
        res.on('end', () => {
          resolve({ statusCode: res.statusCode || 500, body });
        });
      });

      req.on('error', reject);
      req.on('timeout', () => {
        req.destroy();
        reject(new Error('Request timeout'));
      });
      req.end();
    });
  }

  /**
   * Generate cache key
   */
  private cacheKey(name: string, namespace: string, environment: string): string {
    return `${namespace}:${environment}:${name}`;
  }


  /**
   * Start watching for configuration changes
   */
  async watch(...names: string[]): Promise<void> {
    if (this.watching) {
      throw new Error('Already watching');
    }
    this.watching = true;

    // Initial fetch for all configs
    for (const name of names) {
      await this.get(name);
    }

    // Start watch loops
    for (const name of names) {
      this.watchConfig(name);
    }
  }

  /**
   * Watch a single configuration
   */
  private async watchConfig(name: string): Promise<void> {
    const namespace = this.opts.namespace;
    const environment = this.opts.environment;
    const cacheKey = this.cacheKey(name, namespace, environment);

    while (this.watching) {
      try {
        const currentVersion = this.cache.get(cacheKey)?.version || 0;
        const config = await this.watchOnce(name, namespace, environment, currentVersion);

        if (config && config.version > currentVersion) {
          this.cache.set(cacheKey, config);
          this.opts.onChange(config);
        }
      } catch (error) {
        if (error instanceof ConfigHubError && error.code === Errors.WATCH_TIMEOUT) {
          continue; // Normal timeout, retry
        }
        this.opts.onError(error as Error);
        await this.sleep(5000); // Backoff on error
      }
    }
  }

  /**
   * Perform a single long-poll request
   */
  private async watchOnce(
    name: string,
    namespace: string,
    environment: string,
    currentVersion: number
  ): Promise<Config | null> {
    const url = new URL(`${this.opts.serverUrl}/api/v1/config/watch`);
    url.searchParams.set('name', name);
    if (namespace) url.searchParams.set('namespace', namespace);
    if (environment) url.searchParams.set('env', environment);
    url.searchParams.set('version', currentVersion.toString());
    url.searchParams.set('timeout', this.opts.watchTimeout.toString());

    const response = await this.request('GET', url.toString());

    if (response.statusCode === 304) {
      throw new ConfigHubError('Watch timeout', Errors.WATCH_TIMEOUT);
    }
    if (response.statusCode === 401) {
      throw new ConfigHubError('Unauthorized', Errors.UNAUTHORIZED);
    }
    if (response.statusCode !== 200) {
      throw new ConfigHubError(`Watch error: ${response.body}`, 'WATCH_ERROR');
    }

    const result = JSON.parse(response.body);
    if (!result.changed) {
      throw new ConfigHubError('Watch timeout', Errors.WATCH_TIMEOUT);
    }

    return {
      name: result.name,
      namespace: result.namespace,
      environment: result.environment,
      version: result.version,
      content: result.content,
    };
  }

  /**
   * Stop watching for configuration changes
   */
  stopWatch(): void {
    this.watching = false;
    for (const controller of this.watchAbortControllers.values()) {
      controller.abort();
    }
    this.watchAbortControllers.clear();
  }

  /**
   * Close the client
   */
  close(): void {
    this.stopWatch();
  }

  /**
   * Clear the configuration cache
   */
  clearCache(): void {
    this.cache.clear();
  }

  /**
   * Get cached version of a configuration
   */
  getCachedVersion(name: string): number {
    const cacheKey = this.cacheKey(name, this.opts.namespace, this.opts.environment);
    return this.cache.get(cacheKey)?.version || 0;
  }

  /**
   * Sleep helper
   */
  private sleep(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }
}

// Default export
export default ConfigHubClient;
