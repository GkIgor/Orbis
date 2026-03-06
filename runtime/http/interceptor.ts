import { HttpRequest } from "./request";
import { HttpResponse } from "./response";

export interface HttpInterceptor {
  request?: (req: HttpRequest) => HttpRequest;
  response?: <T>(res: HttpResponse<T>) => HttpResponse<T>;
}
