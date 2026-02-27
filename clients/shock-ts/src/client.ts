import { ShockError, ShockLockedError, ShockNetworkError } from "./errors.js";
import type {
  DisplayAcl,
  AclType,
  DownloadOptions,
  NodeListQuery,
  NodeLocation,
  PaginatedResult,
  PollOptions,
  PreAuthResponse,
  ShockClientOptions,
  ShockFile,
  ShockNode,
  ShockPagedResponse,
  ShockResponse,
  ShockServerInfo,
  UploadOptions,
  UploadProgress,
} from "./types.js";
import { isValidNodeId } from "./types.js";

export class ShockClient {
  private baseUrl: string;
  private staticToken?: string;
  private getTokenFn?: () => string | undefined;

  constructor(options: ShockClientOptions) {
    // Strip trailing slash
    this.baseUrl = options.url.replace(/\/+$/, "");
    this.staticToken = options.token;
    this.getTokenFn = options.getToken;
  }

  /** Update the static auth token. */
  setToken(token?: string): void {
    this.staticToken = token;
  }

  // ─── Server Info ───────────────────────────────────────────────

  /** `GET /` — returns the bare server info struct (NOT envelope-wrapped). */
  async getServerInfo(): Promise<ShockServerInfo> {
    const response = await this.rawFetch("GET", "/");
    return response.json() as Promise<ShockServerInfo>;
  }

  // ─── Node CRUD ─────────────────────────────────────────────────

  /** `GET /node/{id}` */
  async getNode<TAttr = unknown>(id: string): Promise<ShockNode<TAttr>> {
    this.validateId(id);
    return this.request<ShockNode<TAttr>>("GET", `/node/${id}`);
  }

  /** `GET /node?...` */
  async listNodes<TAttr = unknown>(
    query?: NodeListQuery
  ): Promise<PaginatedResult<ShockNode<TAttr>>> {
    return this.pagedRequest<ShockNode<TAttr>>("/node", query);
  }

  /**
   * `POST /node` — create a new node.
   * All data is sent as FormData (never JSON) to avoid CORS preflight issues.
   */
  async createNode<TAttr = unknown>(
    options?: UploadOptions
  ): Promise<ShockNode<TAttr>> {
    const form = this.buildUploadForm(options);
    return this.formRequest<ShockNode<TAttr>>("POST", "/node", form);
  }

  /**
   * `PUT /node/{id}` — update a node.
   * All data is sent as FormData (never JSON) to avoid CORS preflight issues.
   */
  async updateNode<TAttr = unknown>(
    id: string,
    options: UploadOptions
  ): Promise<ShockNode<TAttr>> {
    this.validateId(id);
    const form = this.buildUploadForm(options);
    return this.formRequest<ShockNode<TAttr>>("PUT", `/node/${id}`, form);
  }

  /** `DELETE /node/{id}` */
  async deleteNode(id: string): Promise<void> {
    this.validateId(id);
    await this.request<null>("DELETE", `/node/${id}`);
  }

  // ─── Upload Parts ──────────────────────────────────────────────

  /**
   * `PUT /node/{id}` with a part number.
   * Uses XMLHttpRequest for upload progress reporting (browser only).
   * Part numbers are **1-indexed**. Sending 0 will crash the server.
   */
  async uploadPart<TAttr = unknown>(
    nodeId: string,
    partNumber: number,
    data: Blob,
    onProgress?: (progress: UploadProgress) => void
  ): Promise<ShockNode<TAttr>> {
    if (partNumber < 1 || !Number.isInteger(partNumber)) {
      throw new Error(
        `Part number must be a positive integer (got ${partNumber})`
      );
    }
    this.validateId(nodeId);

    const form = new FormData();
    form.append(String(partNumber), data);

    // Use XHR for progress in browser environments
    if (
      onProgress &&
      typeof XMLHttpRequest !== "undefined"
    ) {
      return this.xhrUpload<ShockNode<TAttr>>(
        "PUT",
        `/node/${nodeId}`,
        form,
        onProgress
      );
    }

    return this.formRequest<ShockNode<TAttr>>("PUT", `/node/${nodeId}`, form);
  }

  // ─── Download ──────────────────────────────────────────────────

  /** `GET /node/{id}?download` — returns the file as a Blob. */
  async downloadFile(id: string, options?: DownloadOptions): Promise<Blob> {
    this.validateId(id);
    const params = new URLSearchParams({ download: "" });
    if (options?.seek !== undefined) params.set("seek", String(options.seek));
    if (options?.length !== undefined)
      params.set("length", String(options.length));

    const response = await this.rawFetch(
      "GET",
      `/node/${id}?${params.toString()}`
    );

    if (!response.ok) {
      await this.throwFromResponse(response);
    }
    return response.blob();
  }

  /** `GET /node/{id}?download_url` — returns a pre-auth download response. */
  async getDownloadUrl(id: string): Promise<PreAuthResponse> {
    this.validateId(id);
    return this.request<PreAuthResponse>("GET", `/node/${id}?download_url`);
  }

  // ─── ACL ───────────────────────────────────────────────────────

  /** `GET /node/{id}/acl/` */
  async getAcl(nodeId: string): Promise<DisplayAcl> {
    this.validateId(nodeId);
    return this.request<DisplayAcl>("GET", `/node/${nodeId}/acl/`);
  }

  /** `PUT /node/{id}/acl/{type}?users=...` */
  async addAcl(
    nodeId: string,
    type: AclType,
    users: string[]
  ): Promise<DisplayAcl> {
    this.validateId(nodeId);
    const params = new URLSearchParams({ users: users.join(",") });
    return this.request<DisplayAcl>(
      "PUT",
      `/node/${nodeId}/acl/${type}?${params.toString()}`
    );
  }

  /** `DELETE /node/{id}/acl/{type}?users=...` */
  async removeAcl(
    nodeId: string,
    type: AclType,
    users: string[]
  ): Promise<DisplayAcl> {
    this.validateId(nodeId);
    const params = new URLSearchParams({ users: users.join(",") });
    return this.request<DisplayAcl>(
      "DELETE",
      `/node/${nodeId}/acl/${type}?${params.toString()}`
    );
  }

  // ─── Index ─────────────────────────────────────────────────────

  /** `PUT /node/{id}/index/{type}` — create or rebuild an index. */
  async createIndex(nodeId: string, indexType: string): Promise<void> {
    this.validateId(nodeId);
    await this.request<unknown>("PUT", `/node/${nodeId}/index/${indexType}`);
  }

  // ─── Locations ─────────────────────────────────────────────────

  /** `GET /node/{id}/locations/` */
  async getLocations(nodeId: string): Promise<NodeLocation[]> {
    this.validateId(nodeId);
    return this.request<NodeLocation[]>("GET", `/node/${nodeId}/locations/`);
  }

  // ─── Polling ───────────────────────────────────────────────────

  /**
   * Poll `GET /node/{id}` until `file.locked` is null (assembly/indexing complete).
   * Throws `ShockLockedError` if the lock contains an error or polling times out.
   */
  async pollUntilReady<TAttr = unknown>(
    nodeId: string,
    options?: PollOptions
  ): Promise<ShockNode<TAttr>> {
    const interval = options?.intervalMs ?? 2000;
    const maxAttempts = options?.maxAttempts ?? 150; // 5 min at 2s default

    for (let i = 0; i < maxAttempts; i++) {
      const node = await this.getNode<TAttr>(nodeId);
      if (node.file.locked === null) return node;
      if (node.file.locked.error) {
        throw new ShockLockedError(423, [
          `Assembly failed: ${node.file.locked.error}`,
        ]);
      }
      await new Promise((r) => setTimeout(r, interval));
    }
    throw new ShockLockedError(423, [
      "Timed out waiting for file assembly",
    ]);
  }

  // ─── Internal Helpers ──────────────────────────────────────────

  private resolveToken(): string | undefined {
    if (this.getTokenFn) return this.getTokenFn();
    return this.staticToken;
  }

  private authHeaders(): Record<string, string> {
    const token = this.resolveToken();
    if (token) {
      return { Authorization: `OAuth ${token}` };
    }
    return {};
  }

  private validateId(id: string): void {
    if (!isValidNodeId(id)) {
      throw new Error(`Invalid node ID: "${id}"`);
    }
  }

  /**
   * Raw fetch with auth headers. Does NOT parse the response.
   * Callers are responsible for checking response.ok and parsing.
   */
  private async rawFetch(
    method: string,
    path: string,
    init?: RequestInit
  ): Promise<Response> {
    const url = `${this.baseUrl}${path}`;
    try {
      return await fetch(url, {
        method,
        ...init,
        headers: {
          ...this.authHeaders(),
          ...init?.headers,
        },
      });
    } catch (err) {
      throw new ShockNetworkError(
        `Network error: ${method} ${path}`,
        err
      );
    }
  }

  /**
   * Fetch + unwrap the standard `{ status, data, error }` envelope.
   * Throws `ShockError` on non-2xx responses.
   */
  private async request<T>(method: string, path: string): Promise<T> {
    const response = await this.rawFetch(method, path);
    const json = (await response.json()) as ShockResponse<T>;

    if (!response.ok || (json.error && json.error.length > 0)) {
      throw ShockError.fromResponse(
        response.status,
        json.error
      );
    }
    return json.data as T;
  }

  /** Fetch + unwrap a paginated envelope. */
  private async pagedRequest<T>(
    path: string,
    query?: NodeListQuery
  ): Promise<PaginatedResult<T>> {
    const params = new URLSearchParams();
    if (query?.limit !== undefined) params.set("limit", String(query.limit));
    if (query?.offset !== undefined)
      params.set("offset", String(query.offset));
    if (query?.order) params.set("order", query.order);
    if (query?.direction) params.set("direction", query.direction);
    if (query?.querynode) params.set("querynode", "");
    if (query?.query) {
      for (const [k, v] of Object.entries(query.query)) {
        params.set(k, v);
      }
    }

    const qs = params.toString();
    const fullPath = qs ? `${path}?${qs}` : path;
    const response = await this.rawFetch("GET", fullPath);
    const json = (await response.json()) as ShockPagedResponse<T[]>;

    if (!response.ok || (json.error && json.error.length > 0)) {
      throw ShockError.fromResponse(response.status, json.error);
    }

    return {
      data: (json.data ?? []) as T[],
      limit: json.limit,
      offset: json.offset,
      totalCount: json.total_count,
    };
  }

  /**
   * Send a FormData request. NEVER manually sets Content-Type
   * (the browser/runtime sets it with the correct multipart boundary).
   */
  private async formRequest<T>(
    method: string,
    path: string,
    form: FormData
  ): Promise<T> {
    const response = await this.rawFetch(method, path, { body: form });
    const json = (await response.json()) as ShockResponse<T>;

    if (!response.ok || (json.error && json.error.length > 0)) {
      throw ShockError.fromResponse(response.status, json.error);
    }
    return json.data as T;
  }

  /**
   * Build a FormData payload from upload options.
   *
   * Key rules from shock-server/request/request.go:
   * - Form fields WITHOUT a filename → treated as params
   * - Form fields WITH a filename → treated as file uploads
   * - File upload field name must be "upload", "gzip", or "bzip2"
   * - Attributes: use "attributes_str" param (not file upload)
   */
  private buildUploadForm(options?: UploadOptions): FormData {
    const form = new FormData();
    if (!options) return form;

    if (options.file) {
      const fieldName =
        options.compression === "gzip"
          ? "gzip"
          : options.compression === "bzip2"
            ? "bzip2"
            : "upload";
      form.append(fieldName, options.file, options.fileName || "upload");
    } else if (options.fileName) {
      // Set file name without uploading a file
      form.append("file_name", options.fileName);
    }

    if (options.attributes !== undefined) {
      form.append("attributes_str", JSON.stringify(options.attributes));
    }

    if (options.expiration) {
      form.append("expiration", options.expiration);
    }

    if (options.parts !== undefined) {
      form.append("parts", String(options.parts));
    }

    return form;
  }

  /** Parse an error response body and throw a ShockError. */
  private async throwFromResponse(response: Response): Promise<never> {
    const text = await response.text();
    try {
      const json = JSON.parse(text) as ShockResponse<unknown>;
      throw ShockError.fromResponse(response.status, json.error);
    } catch (err) {
      if (err instanceof ShockError) throw err;
      throw new ShockError(response.status, [`HTTP ${response.status}`]);
    }
  }

  /**
   * Upload via XMLHttpRequest for progress reporting (browser only).
   * Falls back to fetch in environments without XHR.
   */
  private xhrUpload<T>(
    method: string,
    path: string,
    form: FormData,
    onProgress: (progress: UploadProgress) => void
  ): Promise<T> {
    return new Promise((resolve, reject) => {
      const xhr = new XMLHttpRequest();
      xhr.open(method, `${this.baseUrl}${path}`);

      const token = this.resolveToken();
      if (token) {
        xhr.setRequestHeader("Authorization", `OAuth ${token}`);
      }

      xhr.upload.addEventListener("progress", (e) => {
        if (e.lengthComputable) {
          onProgress({
            loaded: e.loaded,
            total: e.total,
            percent: Math.round((e.loaded / e.total) * 100),
          });
        }
      });

      xhr.addEventListener("load", () => {
        try {
          const json = JSON.parse(xhr.responseText) as ShockResponse<T>;
          if (xhr.status >= 200 && xhr.status < 300 && !(json.error && json.error.length > 0)) {
            resolve(json.data as T);
          } else {
            reject(
              ShockError.fromResponse(xhr.status, json.error)
            );
          }
        } catch {
          reject(new ShockError(xhr.status, [`HTTP ${xhr.status}`]));
        }
      });

      xhr.addEventListener("error", () => {
        reject(new ShockNetworkError(`XHR error: ${method} ${path}`));
      });

      xhr.addEventListener("abort", () => {
        reject(new ShockNetworkError(`XHR aborted: ${method} ${path}`));
      });

      xhr.send(form);
    });
  }
}
