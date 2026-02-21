const API_BASE = import.meta.env.VITE_API_URL || '';

function getToken(): string | null {
  return localStorage.getItem('access_token');
}

/** Из ответа с ошибкой вытаскивает сообщение: messages[] или error или statusText */
function getErrorMessage(json: unknown, statusText: string): string {
  if (json && typeof json === 'object') {
    const o = json as Record<string, unknown>;
    if (Array.isArray(o.messages) && o.messages.length > 0) {
      const first = o.messages[0];
      return typeof first === 'string' ? first : String(first);
    }
    if (typeof o.error === 'string' && o.error) return o.error;
    if (typeof o.message === 'string' && o.message) return o.message;
  }
  return statusText || 'Ошибка запроса';
}

export async function request<T>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  const token = getToken();
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string>),
  };
  if (token) {
    (headers as Record<string, string>)['Authorization'] = `Bearer ${token}`;
  }
  const res = await fetch(`${API_BASE}${path}`, { ...options, headers });
  const json = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new Error(getErrorMessage(json, res.statusText));
  }
  if (json == null) return json as T;
  const data = (json as { data?: T }).data !== undefined ? (json as { data: T }).data : json;
  return data as T;
}

/** Запрос, при 404 возвращает fallback вместо ошибки (для «нет данных»). */
async function requestOrNotFound<T>(path: string, fallback: T, options: RequestInit = {}): Promise<T> {
  const token = getToken();
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string>),
  };
  if (token) {
    (headers as Record<string, string>)['Authorization'] = `Bearer ${token}`;
  }
  const res = await fetch(`${API_BASE}${path}`, { ...options, headers });
  if (res.status === 404) return fallback;
  const json = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new Error(getErrorMessage(json, res.statusText));
  }
  if (json == null) return fallback;
  const data = (json as { data?: T }).data !== undefined ? (json as { data: T }).data : json;
  return data as T;
}

const WS_CONNECT_TIMEOUT_MS = 8000   // если за это время WS не открылся — считаем, что не подключается
const WS_RESPONSE_TIMEOUT_MS = 15000 // ожидание ответа на start-session после открытия соединения
// Чаще, чем TTL на бэкенде (1 мин), чтобы ключ client_user_id не истекал при задержках
const KEEPALIVE_INTERVAL_MS = 25 * 1000 // 25 s

function nextMessageId(): string {
  return typeof crypto !== 'undefined' && crypto.randomUUID
    ? crypto.randomUUID()
    : `${Date.now()}-${Math.random().toString(36).slice(2)}`
}

export type WsRentMessage = {
  type?: string
  meta?: { message_id?: string; status?: string; session_id?: string; err?: { message?: string } }
  content?: {
    status?: string
    status_msg?: string
    session_id?: string
    request_id?: string
    initiator?: string
    created_at?: string
    [key: string]: unknown
  }
  error?: string
  message?: string
}

/**
 * Постоянное WS-подключение для сессии аренды: приём сообщений (session-status-updated и др.)
 * и отправка keepalive каждые 25 с (и сразу после открытия). request_id нужен для продления TTL аренды на бэкенде.
 */
export function createRentSessionConnection(params: {
  wsBaseUrl: string
  accessToken: string
  /** request_id активной аренды — передаётся в keepalive, чтобы бэкенд продлевал TTL (без этого сессия оборвётся по таймауту) */
  requestId?: string
  onMessage: (data: WsRentMessage) => void
  onClose?: () => void
}): { close: () => void } {
  const base = params.wsBaseUrl.replace(/\/$/, '')
  const path = wsPathFor(base)
  const url = `${base}${path}?access_token=${encodeURIComponent(params.accessToken)}`
  const ws = new WebSocket(url)
  let keepaliveId: ReturnType<typeof setInterval> | null = null

  const sendKeepalive = () => {
    if (ws.readyState === WebSocket.OPEN) {
      try {
        const content: { request_id?: string } = {}
        if (params.requestId) content.request_id = params.requestId
        ws.send(JSON.stringify({ type: 'keepalive', meta: { message_id: nextMessageId() }, content }))
      } catch {
        close()
      }
    }
  }

  const close = () => {
    if (keepaliveId != null) {
      clearInterval(keepaliveId)
      keepaliveId = null
    }
    try {
      ws.close()
    } catch {
      // ignore
    }
    params.onClose?.()
  }

  ws.onopen = () => {
    sendKeepalive()
    keepaliveId = setInterval(sendKeepalive, KEEPALIVE_INTERVAL_MS)
  }

  ws.onmessage = (event) => {
    try {
      const data = (typeof event.data === 'string' ? JSON.parse(event.data) : event.data) as WsRentMessage
      params.onMessage(data)
    } catch {
      // ignore parse errors
    }
  }

  ws.onerror = () => close()
  ws.onclose = () => close()

  return { close }
}

/** Путь WebSocket: /api/ws при прокси через тот же origin (Vite), иначе /api/v1/ws (прямо к координатору). */
function wsPathFor(baseUrl: string): string {
  if (typeof window === 'undefined') return '/api/v1/ws'
  const origin = (window.location.protocol === 'https:' ? 'wss:' : 'ws:') + '//' + window.location.host
  const base = baseUrl.replace(/\/$/, '')
  return base === origin ? '/api/ws' : '/api/v1/ws'
}

/** Полезная нагрузка ответа session-status-updated после start-session (для немедленного отображения активной аренды). */
export type StartSessionResult = {
  request_id: string
  session_id: string
  status: string
  created_at?: string
}

/** Запуск сессии по WebSocket: подключение к координатору, отправка start-session, ожидание session-status-updated. Возвращает данные сессии для отображения без перезагрузки. */
async function startSessionViaWs(params: {
  wsBaseUrl: string
  accessToken: string
  sessionId: string
  templateId: string
}): Promise<StartSessionResult> {
  const base = params.wsBaseUrl.replace(/\/$/, '')
  const path = wsPathFor(base)
  const url = `${base}${path}?access_token=${encodeURIComponent(params.accessToken)}`
  return new Promise((resolve, reject) => {
    let settled = false
    let connectionTimeoutId: ReturnType<typeof setTimeout> | null = null
    let responseTimeoutId: ReturnType<typeof setTimeout> | null = null
    const clearTimeouts = () => {
      if (connectionTimeoutId != null) {
        window.clearTimeout(connectionTimeoutId)
        connectionTimeoutId = null
      }
      if (responseTimeoutId != null) {
        window.clearTimeout(responseTimeoutId)
        responseTimeoutId = null
      }
    }
    const finish = (err: Error | null, payload?: StartSessionResult) => {
      if (settled) return
      settled = true
      clearTimeouts()
      try {
        ws.close()
      } catch {
        // ignore
      }
      if (err) reject(err)
      else resolve(payload ?? { request_id: '', session_id: params.sessionId, status: 'pending' })
    }
    const ws = new WebSocket(url)
    connectionTimeoutId = window.setTimeout(() => {
      if (ws.readyState !== WebSocket.OPEN) {
        finish(new Error('WebSocket не подключается. Проверьте, что координатор доступен (порт 8090) и прокси /api/ws настроен.'))
      }
    }, WS_CONNECT_TIMEOUT_MS)

    ws.onopen = () => {
      clearTimeouts()
      const messageId = typeof crypto !== 'undefined' && crypto.randomUUID
        ? crypto.randomUUID()
        : `${Date.now()}-${Math.random().toString(36).slice(2)}`
      const msg = {
        type: 'start-session',
        meta: { message_id: messageId },
        content: {
          session_id: params.sessionId,
          settings: { mode: 'proxy', template_id: params.templateId },
        },
      }
      ws.send(JSON.stringify(msg))
      responseTimeoutId = window.setTimeout(
        () => finish(new Error('Таймаут ожидания ответа от сервера')),
        WS_RESPONSE_TIMEOUT_MS
      )
    }

    ws.onmessage = (event) => {
      try {
        const data = typeof event.data === 'string' ? JSON.parse(event.data) : event.data
        const t = data?.type
        if (t === 'error' || data?.error) {
          const errMsg = data?.message ?? data?.error ?? 'Ошибка запуска сессии'
          finish(new Error(typeof errMsg === 'string' ? errMsg : JSON.stringify(errMsg)))
          return
        }
        if (t === 'session-status-updated' || t === 'session_started') {
          const c = data?.content ?? {}
          const payload: StartSessionResult = {
            request_id: c.request_id ?? c.requestID ?? '',
            session_id: c.session_id ?? params.sessionId,
            status: c.status ?? 'pending',
            created_at: c.created_at != null ? new Date(c.created_at).toISOString() : undefined,
          }
          finish(null, payload)
          return
        }
      } catch {
        // ignore parse errors
      }
    }

    ws.onerror = () => finish(new Error('Ошибка соединения WebSocket'))

    ws.onclose = (event) => {
      if (!settled && event.code !== 1000 && !event.wasClean) {
        finish(new Error(event.reason || 'Соединение закрыто'))
      }
    }
  })
}

export const api = {
  auth: {
    signIn: (email: string, password: string) =>
      request<{ user_id: number; access_token: string; refresh_token: string }>('/api/auth/signin', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
      }),
    register: (body: {
      email: string;
      password: string;
      username: string;
      name: string;
      zip_code?: string;
      phone?: string;
    }) =>
      request<{ user_id: number; access_token: string; refresh_token: string }>('/api/auth/register', {
        method: 'POST',
        body: JSON.stringify(body),
      }),
    refresh: (refreshToken: string) =>
      request<{ access_token: string; refresh_token: string }>('/api/auth/refresh', {
        method: 'POST',
        body: JSON.stringify({ refresh_token: refreshToken }),
      }),
  },
  user: {
    get: (userId: number) =>
      request<{
        id: number;
        username: string;
        email: string;
        name: string;
        zip_code?: string;
        phone?: string;
      }>(`/api/users/${userId}`),
    updateProfile: (userId: number, body: { name?: string; zip_code?: string; phone?: string }) =>
      request<unknown>(`/api/users/${userId}/profile`, {
        method: 'PATCH',
        body: JSON.stringify(body),
      }),
    /** Баланс в BYN (тестовая заглушка на бэке). */
    getBalance: (userId: number) =>
      request<{ balance: number }>(`/api/users/${userId}/balance`),
    topUp: (userId: number, amount: number) =>
      request<{ balance: number }>(`/api/users/${userId}/balance/top-up`, {
        method: 'POST',
        body: JSON.stringify({ amount }),
      }),
    withdraw: (userId: number, amount: number) =>
      request<{ balance: number }>(`/api/users/${userId}/balance/withdraw`, {
        method: 'POST',
        body: JSON.stringify({ amount }),
      }),
  },
  /** Список имён пермишенов текущего пользователя (требует авторизации). */
  permissions: {
    getMy: () => request<string[]>('/api/permissions/me'),
  },
  /** Админ: все аренды с именами покупателя и поставщика (требует SystemSettingsWrite). */
  admin: {
    getHistoryRents: () => request<AdminRentHistoryItem[]>('/api/admin/history/rents'),
    /** Получить шаблон по ID со всеми полями (ports, envs, volumes). Требует SystemSettingsWrite. */
    getTemplate: (templateId: string) =>
      request<TemplateInfo>(`/api/admin/templates/${encodeURIComponent(templateId)}`),
    /** Создать шаблон (требует SystemSettingsWrite). */
    createTemplate: (body: AdminTemplatePayload) =>
      request<TemplateInfo>('/api/admin/templates', { method: 'POST', body: JSON.stringify(body) }),
    /** Обновить шаблон (требует SystemSettingsWrite). */
    updateTemplate: (templateId: string, body: AdminTemplatePayload) =>
      request<TemplateInfo>(`/api/admin/templates/${encodeURIComponent(templateId)}`, {
        method: 'PATCH',
        body: JSON.stringify(body),
      }),
  },
  clientRent: {
    /** При 404 (нет активной аренды) возвращает null, ошибку не бросает. */
    getActive: () =>
      requestOrNotFound<ActiveRent | null>('/api/client/rent/active', null),
    /** Список поставщиков, готовых к аренде. При 404 возвращает []. */
    getAvailableMerchants: () =>
      requestOrNotFound<AvailableMerchant[]>('/api/client/rent/merchants', []),
    /** Список шаблонов (образов окружения). При 404 возвращает []. */
    getTemplates: () =>
      requestOrNotFound<TemplateInfo[]>('/api/client/rent/templates', []),
    /** Запуск сессии через WebSocket; возвращает данные сессии для немедленного отображения. */
    startSessionWs: (params: { wsBaseUrl: string; accessToken: string; sessionId: string; templateId: string }) =>
      startSessionViaWs(params),
    stop: (rentId: string, body: { reason?: string }) =>
      request<unknown>(`/api/client/rent/${encodeURIComponent(rentId)}/stop`, {
        method: 'POST',
        body: JSON.stringify(body),
      }),
    /** При 404 (нет истории) возвращает [], ошибку не бросает. */
    getHistory: () =>
      requestOrNotFound<RentHistoryItem[]>('/api/client/rent/history', []),
    /** Настройки подключения для активной аренды (логин, пароль, ссылка). Только для своей аренды. */
    getSettings: (rentId: string) =>
      request<ClientRentSettings>(`/api/client/rent/${encodeURIComponent(rentId)}/settings`),
    /** Характеристики поставщика по session_id (для блока «Поставщик и характеристики» в активной аренде). При 404 возвращает null. */
    getMerchantDetails: (sessionId: string) =>
      requestOrNotFound<AvailableMerchant | null>(
        `/api/client/rent/merchant/${encodeURIComponent(sessionId)}`,
        null
      ),
  },
  /** Поставщик: мои узлы, история сдачи в аренду, отключение узла. */
  merchant: {
    getSessions: () =>
      request<MerchantSession[]>('/api/merchant/sessions'),
    getHistoryLease: () =>
      requestOrNotFound<LeaseHistoryItem[]>('/api/merchant/history/lease', []),
    /** Активная аренда по сессии (для узла в статусе reserved): created_at и т.д. 404 — аренды нет. */
    getSessionActiveRent: (sessionId: string) =>
      requestOrNotFound<{ created_at?: string; createdAt?: string }>(
        `/api/merchant/session/${encodeURIComponent(sessionId)}/active-rent`,
        null
      ),
    stopSession: (sessionId: string) =>
      request<{ ok: boolean }>(`/api/merchant/session/${encodeURIComponent(sessionId)}/stop`, {
        method: 'POST',
      }),
    /** Заглушка: запрос на скачивание приложения поставщика. */
    downloadApp: async () => {
      const token = getToken();
      const headers: HeadersInit = { ...(token ? { Authorization: `Bearer ${token}` } : {}) };
      await fetch(`${API_BASE}/api/merchant/app/download`, { method: 'GET', headers });
    },
  },
};

export type ClientRentSettings = {
  template?: {
    login?: string;
    password?: string;
    authentication?: { login?: string; password?: string };
  };
  network?: {
    piko?: { endpoints?: Array<{ title?: string; link?: string }> };
  };
};

export type TemplatePort = {
  auth?: boolean;
  port?: number;
  type?: string;
  title?: string;
};

export type TemplateEnv = {
  key?: string;
  value?: string;
  type?: string;
};

export type TemplateInfo = {
  template_id: string;
  title?: string;
  type?: string;
  description?: string;
  short_description?: string;
  version?: string;
  container_image_name?: string;
  container_image_tag?: string;
  ports?: TemplatePort[];
  envs?: TemplateEnv[];
  volumes?: string[];
  use_gpu?: boolean;
  min_cpu?: number;
  min_ram_bytes?: number;
  min_storage_bytes?: number;
  min_volume_storage_bytes?: number[];
};

/** Тело запроса создания/обновления шаблона (админ). Тип и версия на бэкенде всегда proxy / 1.0. */
export type AdminTemplatePayload = {
  title?: string;
  description?: string;
  short_description?: string;
  container_image_name?: string;
  container_image_tag?: string;
  ports?: TemplatePort[];
  envs?: TemplateEnv[];
  volumes?: string[];
  use_gpu?: boolean;
  min_cpu?: number;
  min_ram_bytes?: number;
  min_storage_bytes?: number;
  min_volume_storage_bytes?: number[];
};

export type MerchantDetails = {
  total_price: number;
  total_ram: number;
  available_ram: number;
  price_ram: number;
  load_speed: number;
  upload_speed: number;
  ping: number;
  price_internet: number;
  gpus?: { name: string; total_vram: number; available_vram: number; dlperf: number; price: number }[];
  cpus?: { name: string; total: number; available: number; price: number }[];
  storages?: { type: string; name: string; total: number; available: number; bandwidth: number; price: number }[];
  templates?: string[];
};

export type AvailableMerchant = {
  id: string;
  session_id: string;
  user_id?: string;
  name?: string;
  node_name?: string;
  template_id?: string;
  details?: MerchantDetails;
};

export type ActiveRent = {
  id: string;
  session_id: string;
  status: string;
  created_at: string;
  merchant_name?: string;
  settings?: {
    template_id?: string;
    network?: Record<string, unknown>;
    template?: {
      login?: string;
      password?: string;
      authentication?: { login?: string; password?: string };
      [key: string]: unknown;
    };
  };
};

export type RentHistoryItem = {
  id: string;
  cost: number;
  duration: number;
  rating?: number;
  started_at?: string;
  ended_at?: string;
  template_id?: string;
  template_title?: string;
};

/** Узел поставщика (сессия) для блока «Мои узлы». */
export type MerchantSession = {
  session_id: string;
  node_name: string;
  status: string;
  /** Цена за минуту (BYN). Может приходить как total_price или totalPrice. */
  total_price?: number;
  totalPrice?: number;
};

/** Элемент истории сдачи в аренду (тот же формат, что и RentHistoryItem). */
export type LeaseHistoryItem = {
  id: string;
  cost: number;
  duration: number;
  rating?: number;
  started_at?: string;
  ended_at?: string;
  template_id?: string;
  template_title?: string;
};

/** Элемент админ-истории: все аренды с покупателем и поставщиком. */
export type AdminRentHistoryItem = {
  id: string;
  client_id: string;
  merchant_id: string;
  client_name: string;
  merchant_name: string;
  cost: number;
  duration: number;
  started_at?: string;
  ended_at?: string;
  template_id?: string;
  template_title?: string;
};
