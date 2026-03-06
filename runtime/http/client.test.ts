import { describe, it, expect, beforeEach, vi } from "vitest";
import { HttpClient } from "./client.js";
import { HttpException } from "./exception.js";

describe("Orbis HTTP Client Phase 10", () => {
  let client: HttpClient;

  beforeEach(() => {
    client = new HttpClient();
    global.fetch = vi.fn();
  });

  describe("Basic GET and POST", () => {
    it("executes explicit GET fetching identically formatting Observables", async () => {
      const mockResponse = {
        ok: true,
        status: 200,
        headers: new Map([["content-type", "application/json"]]),
        text: vi.fn().mockResolvedValue('{"id": 1}'),
      };
      (global.fetch as any).mockResolvedValue(mockResponse);

      const result = await new Promise((resolve, reject) => {
        client.get("https://api.example.com/users").subscribe({
          next: resolve,
          error: reject,
        });
      });

      expect(global.fetch).toHaveBeenCalledWith(
        "https://api.example.com/users",
        expect.objectContaining({ method: "GET" }),
      );
      expect((result as any).body).toEqual({ id: 1 });
      expect((result as any).status).toEqual(200);
    });

    it("executes POST securely preserving JSON stringified formats natively", async () => {
      const mockResponse = {
        ok: true,
        status: 201,
        headers: new Map(),
        text: vi.fn().mockResolvedValue('{"success": true}'),
      };
      (global.fetch as any).mockResolvedValue(mockResponse);

      const result = await new Promise((resolve, reject) => {
        client
          .post("https://api.example.com/users", { name: "Igor" })
          .subscribe({
            next: resolve,
            error: reject,
          });
      });

      const fetchArgs = (global.fetch as any).mock.calls[0];
      expect(fetchArgs[0]).toBe("https://api.example.com/users");
      expect(fetchArgs[1].method).toBe("POST");
      expect(fetchArgs[1].body).toBe('{"name":"Igor"}');
      expect(fetchArgs[1].headers).toEqual({
        "Content-Type": "application/json",
      });
    });
  });

  describe("Exceptions & Retries", () => {
    it("preserves identical API error payloads cleanly securely", async () => {
      const mockResponse = {
        ok: false,
        status: 404,
        headers: new Map(),
        text: vi.fn().mockResolvedValue('{"error":"USER_NOT_FOUND"}'),
      };
      (global.fetch as any).mockResolvedValue(mockResponse);

      try {
        await new Promise((resolve, reject) => {
          client.get("https://api.example.com/missing").subscribe({
            next: resolve,
            error: reject,
          });
        });
      } catch (err: any) {
        expect(err).toBeInstanceOf(HttpException);
        expect(err.status).toBe(404);
        expect(err.error).toEqual({ error: "USER_NOT_FOUND" });
      }
    });

    it("evaluates retry limits deterministically on rigid network failures", async () => {
      const networkError = new TypeError("Failed to fetch");
      (global.fetch as any).mockRejectedValue(networkError);

      try {
        await new Promise((resolve, reject) => {
          client.get("https://api.example.com/flaky", { retry: 2 }).subscribe({
            next: resolve,
            error: reject,
          });
        });
      } catch (err: any) {
        // Initial attempt + 2 retries = 3 total fetch calls
        expect(global.fetch).toHaveBeenCalledTimes(3);
        expect(err).toBe(networkError);
      }
    });

    it("does NOT retry on 4xx/5xx HTTP logical exceptions natively", async () => {
      const mockResponse = {
        ok: false,
        status: 500,
        headers: new Map(),
        text: vi.fn().mockResolvedValue("{}"),
      };
      (global.fetch as any).mockResolvedValue(mockResponse);

      try {
        await new Promise((resolve, reject) => {
          client.get("https://api.example.com/500", { retry: 5 }).subscribe({
            next: resolve,
            error: reject,
          });
        });
      } catch (err: any) {
        expect(global.fetch).toHaveBeenCalledTimes(1); // logical execution halting synchronously
        expect(err).toBeInstanceOf(HttpException);
      }
    });
  });

  describe("Security Boundaries", () => {
    it("abruptly blocks forbidden protocol schemas natively synchronously before fetch invocation", () => {
      let threw = false;
      client.get("file:///etc/passwd").subscribe({
        error: (err) => {
          threw = true;
          expect(err.message).toMatch(
            /Protocol 'file:' is explicitly forbidden/,
          );
        },
      });
      expect(threw).toBe(true);
      expect(global.fetch).not.toHaveBeenCalled();
    });

    it("abruptly blocks restricted payload headers explicitly overriding standard DOM protections", () => {
      let threw = false;
      client
        .get("https://api.com", { headers: { Host: "evil.com" } })
        .subscribe({
          error: (err) => {
            threw = true;
            expect(err.message).toMatch(
              /Overriding protected header 'Host' is explicitly forbidden/,
            );
          },
        });
      expect(threw).toBe(true);
      expect(global.fetch).not.toHaveBeenCalled();
    });
  });

  describe("Synchronous Interceptors", () => {
    it("maps deterministic pipeline modifications natively isolating sequences", async () => {
      const mockResponse = {
        ok: true,
        status: 200,
        headers: new Map(),
        text: vi.fn().mockResolvedValue('"OK"'),
      };
      (global.fetch as any).mockResolvedValue(mockResponse);

      client.use({
        request: (req) => ({
          ...req,
          config: { ...req.config, headers: { Authorization: "Bearer TOKEN" } },
        }),
        response: (res) => ({ ...res, body: `INTERCEPTED_${res.body}` }),
      });

      const result = await new Promise((resolve, reject) => {
        client
          .get("https://api.com/secure")
          .subscribe({ next: resolve, error: reject });
      });

      expect((global.fetch as any).mock.calls[0][1].headers).toEqual({
        Authorization: "Bearer TOKEN",
      });
      expect((result as any).body).toBe("INTERCEPTED_OK");
    });
  });

  describe("Cancellation", () => {
    it("aborts network natively mirroring standard AbortController structures cleanly", () => {
      const controller = new AbortController();
      const networkPromise = new Promise(() => {}); // never resolves initially hanging pending
      (global.fetch as any).mockReturnValue(networkPromise);

      let closed = false;
      const sub = client
        .get("https://api.com/long", { signal: controller.signal })
        .subscribe({
          complete: () => (closed = true),
          error: () => (closed = true),
        });

      controller.abort();

      expect(sub.closed).toBe(true); // Native RxJS subscription structure torn down identically
    });
  });
});
