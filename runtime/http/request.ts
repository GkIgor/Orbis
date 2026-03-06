export interface HttpConfig {
  headers?: Record<string, string>;
  timeout?: number;
  retry?: number;
  signal?: AbortSignal;
}

export interface HttpRequest {
  method: string;
  url: string;
  body?: any;
  config?: HttpConfig;
}
