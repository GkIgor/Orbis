import { Observable } from "rxjs";
import { HttpRequest, HttpConfig } from "./request";
import { HttpResponse } from "./response";
import { HttpException } from "./exception";
import { HttpPipeline } from "./pipeline";
import { HttpInterceptor } from "./interceptor";
import { validateSecurityBounds } from "./security";

export class HttpClient {
  private pipeline = new HttpPipeline();
  private baseUrl = "";

  constructor(config: { baseUrl?: string } = {}) {
    if (config.baseUrl) {
      this.baseUrl = config.baseUrl;
    }
  }

  use(interceptor: HttpInterceptor): void {
    this.pipeline.use(interceptor);
  }

  get<T = any>(url: string, config?: HttpConfig): Observable<HttpResponse<T>> {
    return this.request<T>({ method: "GET", url, config });
  }

  post<T = any>(
    url: string,
    body?: any,
    config?: HttpConfig,
  ): Observable<HttpResponse<T>> {
    return this.request<T>({ method: "POST", url, body, config });
  }

  put<T = any>(
    url: string,
    body?: any,
    config?: HttpConfig,
  ): Observable<HttpResponse<T>> {
    return this.request<T>({ method: "PUT", url, body, config });
  }

  delete<T = any>(
    url: string,
    config?: HttpConfig,
  ): Observable<HttpResponse<T>> {
    return this.request<T>({ method: "DELETE", url, config });
  }

  /**
   * Resolves execution purely isolating RxJS observables wrapping native boundaries directly.
   * Completely decoupled from rendering loops ensuring manual execution explicitly required.
   */
  request<T>(req: HttpRequest): Observable<HttpResponse<T>> {
    return new Observable<HttpResponse<T>>((subscriber) => {
      const fullUrl = this._resolveUrl(req.url);
      const initialReq: HttpRequest = { ...req, url: fullUrl };

      // Phase 10: Security Validation Bound
      try {
        validateSecurityBounds(initialReq, !!this.baseUrl);
      } catch (err) {
        subscriber.error(err);
        return;
      }

      // Phase 10: Synchronous Interceptor Pass (Request)
      let finalReq: HttpRequest;
      try {
        finalReq = this.pipeline.executeRequestInterceptors(initialReq);
      } catch (err) {
        subscriber.error(err);
        return;
      }

      // Explicit Abort Configuration tightly coupling upstream logic seamlessly to native fetch boundaries
      const controller = new AbortController();
      let abortedLocally = false;

      if (finalReq.config?.signal) {
        finalReq.config.signal.addEventListener("abort", () => {
          abortedLocally = true;
          controller.abort();
          subscriber.error(
            new Error("Orbis HTTP: Request explicitly aborted."),
          );
        });
        if (finalReq.config.signal.aborted) {
          abortedLocally = true;
          controller.abort();
          subscriber.error(
            new Error("Orbis HTTP: Request explicitly aborted."),
          );
        }
      }

      const fetchOptions: RequestInit = {
        method: finalReq.method,
        headers: finalReq.config?.headers || {},
        signal: controller.signal,
      };

      if (finalReq.body) {
        fetchOptions.body = JSON.stringify(finalReq.body);
        const headers = fetchOptions.headers as Record<string, string>;
        if (!headers["Content-Type"] && !headers["content-type"]) {
          headers["Content-Type"] = "application/json";
        }
      }

      let timeoutId: any;
      if (finalReq.config?.timeout) {
        timeoutId = setTimeout(() => {
          abortedLocally = true;
          controller.abort();
          subscriber.error(
            new Error(
              `Orbis HTTP: Request timed out strictly after ${finalReq.config!.timeout}ms`,
            ),
          );
        }, finalReq.config.timeout);
      }

      // Deterministic Manual Retry Evaluation rigidly limiting sequences organically to valid TCP drops natively
      const executeFetch = (attemptsLeft: number) => {
        if (abortedLocally) return;

        fetch(finalReq.url, fetchOptions)
          .then(async (nativeResponse) => {
            if (timeoutId) clearTimeout(timeoutId);

            const headers: Record<string, string> = {};
            nativeResponse.headers.forEach((value, key) => {
              headers[key] = value; // Native format mapping securely isolating objects internally
            });

            // 4xx/5xx mapping into deterministic Exception structure identically transferring payload origin unconditionally
            if (!nativeResponse.ok) {
              let errorPayload: unknown;
              const text = await nativeResponse.text();
              try {
                errorPayload = JSON.parse(text);
              } catch {
                errorPayload = text;
              }
              const exception = new HttpException(
                nativeResponse.status,
                headers,
                errorPayload,
              );
              subscriber.error(exception);
              return;
            }

            // Explicit native payload resolution
            let body: any;
            if (nativeResponse.status !== 204) {
              const text = await nativeResponse.text();
              if (text) {
                try {
                  body = JSON.parse(text);
                } catch {
                  subscriber.error(
                    new Error(
                      `Orbis HTTP: Deterministic JSON compilation failed violently returning invalid data strings.`,
                    ),
                  );
                  return;
                }
              }
            }

            const initialRes: HttpResponse<T> = {
              status: nativeResponse.status,
              headers,
              body,
            };

            // Phase 10: Synchronous Interceptor Pass (Response)
            let finalRes: HttpResponse<T>;
            try {
              finalRes = this.pipeline.executeResponseInterceptors(initialRes);
            } catch (err) {
              subscriber.error(err);
              return;
            }

            subscriber.next(finalRes);
            subscriber.complete();
          })
          .catch((err) => {
            if (abortedLocally) {
              subscriber.error(err);
              return;
            }
            if (attemptsLeft > 0) {
              // Exact iteration mapping explicitly without recursive overheads or randomized timeouts
              executeFetch(attemptsLeft - 1);
            } else {
              subscriber.error(err);
            }
          });
      };

      const configuredRetries = finalReq.config?.retry ?? 0;
      executeFetch(configuredRetries);

      // Cleanup logic executed synchronously natively ending streams
      return () => {
        if (!controller.signal.aborted) {
          abortedLocally = true;
          controller.abort();
        }
        if (timeoutId) clearTimeout(timeoutId);
      };
    });
  }

  private _resolveUrl(path: string): string {
    if (path.includes("://")) return path;
    if (this.baseUrl) {
      const base = this.baseUrl.endsWith("/")
        ? this.baseUrl.slice(0, -1)
        : this.baseUrl;
      const p = path.startsWith("/") ? path : `/${path}`;
      return `${base}${p}`;
    }
    return path;
  }
}
