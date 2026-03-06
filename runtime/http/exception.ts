export class HttpException extends Error {
  status: number;
  headers: Record<string, string>;
  error: unknown;

  constructor(status: number, headers: Record<string, string>, error: unknown) {
    super(`HttpException: ${status}`);
    this.name = "HttpException";
    this.status = status;
    this.headers = headers;
    this.error = error;

    // Explicitly restore prototype chain for native Error subclassing in TS targeting older ES configurations
    Object.setPrototypeOf(this, HttpException.prototype);
  }
}
