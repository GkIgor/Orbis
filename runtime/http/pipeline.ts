import { HttpRequest } from "./request";
import { HttpResponse } from "./response";
import { HttpInterceptor } from "./interceptor";

export class HttpPipeline {
  private interceptors: HttpInterceptor[] = [];

  use(interceptor: HttpInterceptor): void {
    this.interceptors.push(interceptor);
  }

  /**
   * Evaluates completely synchronous logic sequentially altering Request topologies rigidly preceding network boundaries natively
   */
  executeRequestInterceptors(req: HttpRequest): HttpRequest {
    let currentReq = req;
    for (const interceptor of this.interceptors) {
      if (interceptor.request) {
        currentReq = interceptor.request(currentReq);
      }
    }
    return currentReq;
  }

  /**
   * Evaluates securely immediately following physical TCP responses synchronously preventing hidden reactivity queues explicitly.
   */
  executeResponseInterceptors<T>(res: HttpResponse<T>): HttpResponse<T> {
    let currentRes = res;
    for (const interceptor of this.interceptors) {
      if (interceptor.response) {
        currentRes = interceptor.response(currentRes);
      }
    }
    return currentRes;
  }
}
