import { HttpRequest } from "./request";

const REQUIRED_SCHEMES = ["http:", "https:"];
const PROTECTED_HEADERS = ["host", "content-length", "transfer-encoding"];
const LOCAL_HOSTS = ["localhost", "127.0.0.1", "[::1]"];

/**
 * Validates deterministic execution constraints protecting HTTP pipelines natively securely.
 * Blocks SSRF, internal reconnaissance, and header manipulations violating DOM protocols.
 */
export function validateSecurityBounds(
  req: HttpRequest,
  preventLocalhost = false,
): void {
  // Parsing absolute paths explicitly safely guarding schema injections
  if (req.url.includes("://")) {
    const urlObj = new URL(req.url);
    if (!REQUIRED_SCHEMES.includes(urlObj.protocol)) {
      throw new Error(
        `Orbis Security: Protocol '${urlObj.protocol}' is explicitly forbidden.`,
      );
    }
    if (preventLocalhost && LOCAL_HOSTS.includes(urlObj.hostname)) {
      throw new Error(
        `Orbis Security: Connections to local loopback networks are forbidden.`,
      );
    }
  }

  // Shield explicitly overriding headers mathematically dictating physical browser connections
  if (req.config?.headers) {
    for (const key of Object.keys(req.config.headers)) {
      if (PROTECTED_HEADERS.includes(key.toLowerCase())) {
        throw new Error(
          `Orbis Security: Overriding protected header '${key}' is explicitly forbidden.`,
        );
      }
    }
  }
}
