import { useState, useEffect, useCallback, useMemo } from 'react'
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  Cell,
} from 'recharts'
import { api, createRentSessionConnection, type ActiveRent, type ClientRentSettings, type RentHistoryItem, type AvailableMerchant, type MerchantDetails, type TemplateInfo } from '../api/client'
import styles from './ClientView.module.css'

const RANGE_OPTIONS = [7, 14, 30] as const
type RangeDays = (typeof RANGE_OPTIONS)[number]

/** Возвращает ключ даты в локальном времени (YYYY-MM-DD), без смещения по часовому поясу. */
function toLocalDateKey(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

/** Группируем сессии по дню. Бэкенд присылает started_at в UTC; берём дату из ISO-строки (YYYY-MM-DD), чтобы сессии не уезжали на следующий день в UTC+3. */
function getItemDateKey(r: RentHistoryItem): string | null {
  const raw = r.started_at || r.ended_at
  if (!raw) return null
  const trimmed = raw.trim()
  const dateOnlyMatch = /^(\d{4})-(\d{2})-(\d{2})$/.exec(trimmed)
  if (dateOnlyMatch) {
    const [, y, m, day] = dateOnlyMatch
    const year = parseInt(y!, 10)
    const month = parseInt(m!, 10) - 1
    const d = parseInt(day!, 10)
    const date = new Date(year, month, d)
    return toLocalDateKey(date)
  }
  if (/^\d{4}-\d{2}-\d{2}T/.test(trimmed)) {
    return trimmed.slice(0, 10)
  }
  const date = new Date(trimmed)
  return toLocalDateKey(date)
}

function addDays(d: Date, n: number): Date {
  const out = new Date(d)
  out.setDate(out.getDate() + n)
  return out
}

function startOfDay(d: Date): Date {
  const out = new Date(d)
  out.setHours(0, 0, 0, 0)
  return out
}

/** RAM: бэкенд отдаёт в KB. Приводим к ГБ. */
function ramKbToGb(kb: number): number {
  return kb / (1024 * 1024)
}

/** Сеть: значения могут приходить в условных единицах (делим на 1024 для Мбит/с). */
function networkSpeedToMbit(value: number): number {
  if (value > 5000) return value / 1024
  return value
}

/** Хранилище: из KB в ГБ. */
function storageToGb(kb: number): number {
  return kb / (1024 * 1024)
}

export default function ClientView() {
  const [activeRent, setActiveRent] = useState<ActiveRent | null | undefined>(undefined)
  const [activeError, setActiveError] = useState<string | null>(null)
  const [history, setHistory] = useState<RentHistoryItem[]>([])
  const [historyError, setHistoryError] = useState<string | null>(null)
  const [loadingActive, setLoadingActive] = useState(true)
  const [loadingHistory, setLoadingHistory] = useState(true)
  const [stopConfirm, setStopConfirm] = useState(false)
  const [stopReason, setStopReason] = useState('')
  const [stopSubmitting, setStopSubmitting] = useState(false)
  const [stopError, setStopError] = useState<string | null>(null)
  const [startModalOpen, setStartModalOpen] = useState(false)
  const [startStep, setStartStep] = useState<'merchants' | 'templates'>('merchants')
  const [selectedMerchant, setSelectedMerchant] = useState<AvailableMerchant | null>(null)
  const [selectedTemplateId, setSelectedTemplateId] = useState<string | null>(null)
  const [merchants, setMerchants] = useState<AvailableMerchant[]>([])
  const [merchantsLoading, setMerchantsLoading] = useState(false)
  const [merchantsError, setMerchantsError] = useState<string | null>(null)
  const [templates, setTemplates] = useState<TemplateInfo[]>([])
  const [templatesLoading, setTemplatesLoading] = useState(false)
  const [templatesError, setTemplatesError] = useState<string | null>(null)
  const [startSubmitting, setStartSubmitting] = useState(false)
  const [startError, setStartError] = useState<string | null>(null)
  const [expandedMerchantId, setExpandedMerchantId] = useState<string | null>(null)
  const [showActiveMerchant, setShowActiveMerchant] = useState(false)
  const [showActiveSettings, setShowActiveSettings] = useState(false)
  const [activeRentSettings, setActiveRentSettings] = useState<ClientRentSettings | null>(null)
  const [settingsLoading, setSettingsLoading] = useState(false)
  const [activeMerchantDetails, setActiveMerchantDetails] = useState<MerchantDetails | null>(null)
  const [activeMerchantPricePerMin, setActiveMerchantPricePerMin] = useState<number | null>(null)
  const [activeMerchantDetailsLoading, setActiveMerchantDetailsLoading] = useState(false)
  const [historyRangeDays, setHistoryRangeDays] = useState<RangeDays>(30)
  const [historyWindowEnd, setHistoryWindowEnd] = useState<Date>(() => startOfDay(new Date()))
  const [selectedHistoryDay, setSelectedHistoryDay] = useState<string | null>(null)
  const [historyTemplatePreview, setHistoryTemplatePreview] = useState<{ item: RentHistoryItem; templateInfo: TemplateInfo | null } | null>(null)
  const [historyTemplatePreviewLoading, setHistoryTemplatePreviewLoading] = useState(false)

  const WS_BASE_URL = (() => {
    if (import.meta.env.VITE_WS_URL) return import.meta.env.VITE_WS_URL
    if (typeof window === 'undefined') return 'ws://localhost:8090'
    const host = window.location.hostname
    const isLocalhost = host === 'localhost' || host === '127.0.0.1'
    if (isLocalhost) return 'ws://localhost:8090'
    return (window.location.protocol === 'https:' ? 'wss:' : 'ws:') + '//' + window.location.host
  })()
  const accessToken = typeof localStorage !== 'undefined' ? localStorage.getItem('access_token') : null

  const POLL_ACTIVE_MS = 4000

  /** Загрузить активную аренду по HTTP. silent: не показывать индикатор загрузки (для опроса). keepPreviousIfNull: не сбрасывать состояние при null (после старта). */
  const loadActive = useCallback(async (opts?: { keepPreviousIfNull?: boolean; silent?: boolean }) => {
    if (!opts?.silent) setLoadingActive(true)
    setActiveError(null)
    try {
      const data = await api.clientRent.getActive()
      setActiveRent((prev) => {
        if (data != null) return data
        if (opts?.keepPreviousIfNull && prev != null) return prev
        return null
      })
    } catch (e) {
      setActiveRent((prev) => (opts?.keepPreviousIfNull ? prev : null))
      setActiveError(e instanceof Error ? e.message : 'Не удалось загрузить данные об аренде')
    } finally {
      setLoadingActive(false)
    }
  }, [])

  const loadHistory = useCallback(async () => {
    setLoadingHistory(true)
    setHistoryError(null)
    try {
      const data = await api.clientRent.getHistory()
      setHistory(Array.isArray(data) ? data : [])
    } catch (e) {
      setHistory([])
      setHistoryError(e instanceof Error ? e.message : 'Не удалось загрузить историю')
    } finally {
      setLoadingHistory(false)
    }
  }, [])

  useEffect(() => {
    loadActive()
  }, [loadActive])

  useEffect(() => {
    loadHistory()
  }, [loadHistory])

  const today = useMemo(() => startOfDay(new Date()), [])

  const historyByDay = useMemo(() => {
    const map = new Map<string, { items: RentHistoryItem[]; totalCost: number }>()
    for (const r of history) {
      const key = getItemDateKey(r)
      if (!key) continue
      const existing = map.get(key)
      if (existing) {
        existing.items.push(r)
        existing.totalCost += r.cost
      } else {
        map.set(key, { items: [r], totalCost: r.cost })
      }
    }
    for (const entry of map.values()) {
      entry.items.sort((a, b) => {
        const ta = a.started_at || a.ended_at || ''
        const tb = b.started_at || b.ended_at || ''
        return tb.localeCompare(ta)
      })
    }
    return map
  }, [history])

  const historyWindowStart = useMemo(() => {
    return addDays(historyWindowEnd, -(historyRangeDays - 1))
  }, [historyWindowEnd, historyRangeDays])

  const historyChartData = useMemo(() => {
    const data: { dateKey: string; dayLabel: string; count: number }[] = []
    let d = new Date(historyWindowStart)
    const end = new Date(historyWindowEnd)
    while (d <= end) {
      const dateKey = toLocalDateKey(d)
      const dayInfo = historyByDay.get(dateKey)
      data.push({
        dateKey,
        dayLabel: d.toLocaleDateString('ru-RU', { day: '2-digit', month: '2-digit' }),
        count: dayInfo ? dayInfo.items.length : 0,
      })
      d = addDays(d, 1)
    }
    return data
  }, [historyWindowStart, historyWindowEnd, historyByDay])

  const selectedDayInfo = useMemo(() => {
    if (!selectedHistoryDay) return null
    return historyByDay.get(selectedHistoryDay) ?? null
  }, [selectedHistoryDay, historyByDay])

  const canHistoryGoLater = historyWindowEnd < today

  // Опрос активной аренды раз в несколько секунд (без WS)
  useEffect(() => {
    if (!accessToken) return
    const t = setInterval(() => loadActive({ silent: true }), POLL_ACTIVE_MS)
    return () => clearInterval(t)
  }, [accessToken, loadActive])

  const hasActive = activeRent != null && typeof activeRent === 'object'
  const isConfiguring = hasActive && (activeRent.status === 'pending')
  const isStarted = hasActive && (activeRent.status === 'started' || activeRent.status === 'running')

  // Постоянное WS-подключение для аренды: keepalive с request_id продлевает TTL на бэкенде (без этого сессия оборвётся через 1 мин)
  useEffect(() => {
    if (!hasActive || !activeRent?.id || !accessToken) return
    const conn = createRentSessionConnection({
      wsBaseUrl: WS_BASE_URL,
      accessToken,
      requestId: activeRent.id,
      onMessage: (data) => {
        if (data?.type === 'session-status-updated' && data?.content) {
          setActiveRent((prev) => prev ? { ...prev, status: data.content?.status ?? prev.status } : prev)
        }
      },
      onClose: () => { loadActive({ keepPreviousIfNull: true, silent: true }) },
    })
    return () => conn.close()
  }, [hasActive, activeRent?.id, accessToken, WS_BASE_URL, loadActive])

  // Обновление времени сессии раз в секунду для запущенной аренды
  const [sessionTick, setSessionTick] = useState(0)
  useEffect(() => {
    if (!isStarted || !activeRent?.created_at) return
    const interval = setInterval(() => setSessionTick((n) => n + 1), 1000)
    return () => clearInterval(interval)
  }, [isStarted, activeRent?.created_at])

  // Расчёт «Потрачено» — как у продавца: время с created_at + 180 мин, минуты = floor(elapsedMsWithBonus/60000), сумма = цена * минуты
  const currentSpent = useMemo(() => {
    const price = activeMerchantPricePerMin
    if (!isStarted || !activeRent?.created_at || price == null || !Number.isFinite(price)) return null
    const startTime = new Date(activeRent.created_at).getTime()
    const now = Date.now()
    const elapsedMs = now - startTime
    const BONUS_180_MIN_MS = 180 * 60 * 1000
    const elapsedMsWithBonus = elapsedMs + BONUS_180_MIN_MS
    const minutes = Math.floor(elapsedMsWithBonus / 60000)
    return price * minutes
  }, [isStarted, activeRent?.created_at, activeMerchantPricePerMin, sessionTick])

  /** Цена за минуту из мерчанта/ответа API (total_price или totalPrice в details или в корне), как у продавца. */
  const extractPriceFromMerchant = useCallback((obj: unknown): number | null => {
    if (obj == null || typeof obj !== 'object') return null
    const o = obj as Record<string, unknown>
    const details = (o.details != null && typeof o.details === 'object' ? o.details : o) as Record<string, unknown>
    const p =
      Number(details?.total_price) ||
      Number(details?.totalPrice) ||
      Number(o?.total_price) ||
      Number(o?.totalPrice)
    return Number.isFinite(p) && p >= 0 ? p : null
  }, [])

  // Загрузка настроек подключения (логин, пароль, ссылка) при открытии блока «Ссылка и доступ» для запущенной аренды
  useEffect(() => {
    if (!showActiveSettings || !isStarted || !activeRent?.id) return
    let cancelled = false
    setActiveRentSettings(null)
    setSettingsLoading(true)
    api.clientRent.getSettings(activeRent.id)
      .then((data) => { if (!cancelled) setActiveRentSettings(data) })
      .catch(() => { if (!cancelled) setActiveRentSettings(null) })
      .finally(() => { if (!cancelled) setSettingsLoading(false) })
    return () => { cancelled = true }
  }, [showActiveSettings, isStarted, activeRent?.id])

  // Загрузка характеристик и цены для «Потрачено» (как у продавца: цена за минуту из мерчанта/сессии)
  useEffect(() => {
    const needDetails = (showActiveMerchant || isStarted) && hasActive && activeRent?.session_id
    if (!needDetails) return
    const sessionId = activeRent.session_id
    let cancelled = false
    if (showActiveMerchant) setActiveMerchantDetails(null)
    setActiveMerchantDetailsLoading(true)

    function trySetPrice(data: unknown) {
      const price = extractPriceFromMerchant(data)
      if (price != null) setActiveMerchantPricePerMin(price)
    }

    api.clientRent.getMerchantDetails(sessionId)
      .then((data) => {
        if (cancelled) return
        if (data != null && typeof data === 'object' && 'details' in data && (data as { details?: unknown }).details) {
          setActiveMerchantDetails((data as { details: MerchantDetails }).details)
        }
        trySetPrice(data)
      })
      .catch(() => {
        if (!cancelled && showActiveMerchant) setActiveMerchantDetails(null)
      })
      .finally(() => {
        if (!cancelled) setActiveMerchantDetailsLoading(false)
      })

    if (isStarted) {
      api.clientRent.getAvailableMerchants()
        .then((list) => {
          if (cancelled) return
          const arr = Array.isArray(list) ? list : []
          const merchant = arr.find((m: { session_id?: string }) => m.session_id === sessionId)
          if (merchant != null) {
            trySetPrice(merchant.details ?? merchant)
          }
        })
        .catch(() => {})
    }

    return () => { cancelled = true }
  }, [showActiveMerchant, isStarted, hasActive, activeRent?.session_id, extractPriceFromMerchant])

  // Повторно запрашиваем цену, пока «Потрачено» не сможем посчитать (после перезагрузки страницы и т.п.)
  useEffect(() => {
    if (!isStarted || !activeRent?.session_id || activeMerchantPricePerMin != null) return
    const sessionId = activeRent.session_id
    const interval = setInterval(() => {
      api.clientRent.getMerchantDetails(sessionId)
        .then((data) => {
          const price = extractPriceFromMerchant(data)
          if (price != null) setActiveMerchantPricePerMin(price)
        })
        .catch(() => {})
      api.clientRent.getAvailableMerchants()
        .then((list) => {
          const arr = Array.isArray(list) ? list : []
          const m = arr.find((x: { session_id?: string }) => x.session_id === sessionId)
          if (m != null) {
            const price = extractPriceFromMerchant(m.details ?? m)
            if (price != null) setActiveMerchantPricePerMin(price)
          }
        })
        .catch(() => {})
    }, 8000)
    return () => clearInterval(interval)
  }, [isStarted, activeRent?.session_id, activeMerchantPricePerMin, extractPriceFromMerchant])

  const handleStop = async () => {
    if (!activeRent) return
    setStopError(null)
    setStopSubmitting(true)
    try {
      await api.clientRent.stop(activeRent.id, { reason: stopReason || undefined })
      setStopConfirm(false)
      setStopReason('')
      await loadActive()
      await loadHistory()
    } catch (e) {
      setStopError(e instanceof Error ? e.message : 'Ошибка остановки')
    } finally {
      setStopSubmitting(false)
    }
  }

  const openStartModal = () => {
    setStartError(null)
    setMerchantsError(null)
    setStartStep('merchants')
    setSelectedMerchant(null)
    setSelectedTemplateId(null)
    setStartModalOpen(true)
  }

  const goBackToMerchants = () => {
    setStartStep('merchants')
    setSelectedMerchant(null)
    setSelectedTemplateId(null)
    setStartError(null)
  }

  const openTemplateStep = (merchant: AvailableMerchant) => {
    setSelectedMerchant(merchant)
    setSelectedTemplateId(null)
    setStartError(null)
    setStartStep('templates')
  }

  const POLL_INTERVAL_MS = 4000

  useEffect(() => {
    if (!startModalOpen) return
    setMerchantsLoading(true)
    setMerchantsError(null)
    api.clientRent.getAvailableMerchants()
      .then((list) => setMerchants(Array.isArray(list) ? list : []))
      .catch((e) => {
        setMerchants([])
        setMerchantsError(e instanceof Error ? e.message : 'Не удалось загрузить список')
      })
      .finally(() => setMerchantsLoading(false))

    setTemplatesLoading(true)
    setTemplatesError(null)
    api.clientRent.getTemplates()
      .then((list) => setTemplates(Array.isArray(list) ? list : [] as TemplateInfo[]))
      .catch((e) => {
        setTemplates([])
        setTemplatesError(e instanceof Error ? e.message : 'Не удалось загрузить темплейты')
      })
      .finally(() => setTemplatesLoading(false))

    const intervalId = setInterval(() => {
      api.clientRent.getAvailableMerchants()
        .then((list) => setMerchants(Array.isArray(list) ? list : []))
        .catch(() => { /* не сбрасываем список при ошибке опроса */ })
    }, POLL_INTERVAL_MS)

    return () => clearInterval(intervalId)
  }, [startModalOpen])

  const handleStartSession = async () => {
    if (!selectedMerchant || !selectedTemplateId || !accessToken) return
    setStartError(null)
    setStartSubmitting(true)
    const sessionId = selectedMerchant.session_id
    try {
      const result = await api.clientRent.startSessionWs({
        wsBaseUrl: WS_BASE_URL,
        accessToken,
        sessionId,
        templateId: selectedTemplateId,
      })
      setStartModalOpen(false)
      setStartStep('merchants')
      // Цену за минуту — как у продавца: из выбранного мерчанта (details или корень, snake/camel)
      let pricePerMin = extractPriceFromMerchant(selectedMerchant)
      if (pricePerMin == null) {
        try {
          const details = await api.clientRent.getMerchantDetails(sessionId)
          pricePerMin = extractPriceFromMerchant(details)
        } catch {
          /* оставляем null, подтянем в эффекте */
        }
      }
      if (pricePerMin != null) setActiveMerchantPricePerMin(pricePerMin)
      setSelectedMerchant(null)
      setSelectedTemplateId(null)
      // Сразу показываем активную аренду
      setActiveRent({
        id: result.request_id || result.session_id,
        session_id: result.session_id,
        status: result.status,
        created_at: result.created_at ?? new Date().toISOString(),
        settings: { template_id: selectedTemplateId },
      })
      await loadActive({ keepPreviousIfNull: true, silent: true })
      await loadHistory()
    } catch (e) {
      setStartError(e instanceof Error ? e.message : 'Ошибка запуска аренды')
    } finally {
      setStartSubmitting(false)
    }
  }

  const openHistoryTemplatePreview = useCallback((item: RentHistoryItem) => {
    const tid = item.template_id
    setHistoryTemplatePreview({ item, templateInfo: null })
    setHistoryTemplatePreviewLoading(true)
    api.clientRent.getTemplates()
      .then((list) => {
        const t = Array.isArray(list) ? list.find((x) => x.template_id === tid) ?? null : null
        setHistoryTemplatePreview((prev) => (prev ? { ...prev, templateInfo: t } : null))
      })
      .catch(() => setHistoryTemplatePreview((prev) => (prev ? { ...prev, templateInfo: null } : null)))
      .finally(() => setHistoryTemplatePreviewLoading(false))
  }, [])

  return (
    <div className={styles.wrap}>
      <p className={styles.pageIntro}>
        Нажмите «Оформить аренду», чтобы выбрать поставщика и начать сеанс.
      </p>

      <section className={styles.section} aria-labelledby="active-rent-heading">
        <h3 id="active-rent-heading">Активная аренда</h3>
        <p className={styles.sectionHint}>Текущий сеанс аренды: статус, время и данные для входа.</p>
        {loadingActive && <p className={styles.muted}>Загрузка…</p>}
        {!loadingActive && activeError && (
          <div className={styles.errorBlock}>
            {activeError}
          </div>
        )}
        {!loadingActive && !activeError && hasActive && (
          <div className={styles.activeRentCard}>
            <div className={styles.activeRentHeader}>
              <span
                className={styles.statusBadge + (isConfiguring ? ' ' + styles.statusPending : ' ' + styles.statusRunning)}
                title={isConfiguring ? 'Поставщик подготавливает сеанс; данные для входа появятся после запуска' : 'Сеанс запущен, можно подключаться'}
              >
                {activeRent.status === 'pending'
                  ? 'Настройка'
                  : activeRent.status === 'started' || activeRent.status === 'running'
                    ? 'Запущена'
                    : activeRent.status}
              </span>
            </div>
            <div className={styles.activeRentMeta}>
              {activeRent.created_at && (
                <div className={styles.metaItem}>
                  <span className={styles.metaLabel} title="Дата и время начала сеанса">Начало</span>
                  <span className={styles.metaValue}>{new Date(activeRent.created_at).toLocaleString('ru-RU')}</span>
                </div>
              )}
              {isStarted && activeRent.created_at && (
                <div className={styles.metaItem}>
                  <span className={styles.metaLabel} title="Сколько времени прошло с начала текущего сеанса">Время сеанса</span>
                  <span className={styles.metaValue}>
                    {(() => {
                      const start = new Date(activeRent.created_at).getTime()
                      let elapsedSec = Math.floor((Date.now() - start) / 1000)
                      if (elapsedSec < 0) elapsedSec += 180 * 60
                      const min = Math.floor(elapsedSec / 60)
                      const sec = elapsedSec % 60
                      return `${min} мин ${sec} сек`
                    })()}
                  </span>
                </div>
              )}
              {isStarted && currentSpent != null && (
                <div className={styles.metaItem}>
                  <span className={styles.metaLabel} title="Сколько уже потрачено за текущий сеанс">Потрачено</span>
                  <span className={styles.metaValue}>{currentSpent.toFixed(2)} BYN</span>
                </div>
              )}
            </div>
            <div className={styles.activeRentBlocks}>
              <button
                type="button"
                className={styles.toggleBlock}
                onClick={() => setShowActiveMerchant((v) => !v)}
                aria-expanded={showActiveMerchant}
              >
                {showActiveMerchant ? '▼' : '▶'} Поставщик и характеристики
              </button>
            {showActiveMerchant && (
              <div className={styles.activeRentBlockContent}>
                {activeRent.merchant_name && (
                  <div className={styles.cardRow}>
                    <span className={styles.label} title="Имя пользователя поставщика">Поставщик:</span>
                    <span>{activeRent.merchant_name}</span>
                  </div>
                )}
                <div className={styles.cardRow}>
                  <span className={styles.label} title="Уникальный идентификатор сеанса поставщика">Сеанс поставщика:</span>
                  <span className={styles.mono}>{activeRent.session_id}</span>
                </div>
                {activeMerchantDetailsLoading && (
                  <p className={styles.muted}>Загрузка характеристик…</p>
                )}
                {!activeMerchantDetailsLoading && activeMerchantDetails && (
                  <div className={styles.merchantDetails}>
                    <div className={styles.detailRow}>
                      <strong title="Объём оперативной памяти">Оперативная память:</strong>{' '}
                      {ramKbToGb(activeMerchantDetails.total_ram).toFixed(1)} ГБ
                    </div>
                    {activeMerchantDetails.load_speed != null && (
                      <div className={styles.detailRow}>
                        <strong>Скорость загрузки:</strong> {networkSpeedToMbit(activeMerchantDetails.load_speed).toFixed(1)} Мбит/с
                      </div>
                    )}
                    {activeMerchantDetails.upload_speed != null && (
                      <div className={styles.detailRow}>
                        <strong>Скорость отдачи:</strong> {networkSpeedToMbit(activeMerchantDetails.upload_speed).toFixed(1)} Мбит/с
                      </div>
                    )}
                    {activeMerchantDetails.ping != null && (
                      <div className={styles.detailRow}>
                        <strong title="Сетевая задержка до машины">Задержка:</strong> {activeMerchantDetails.ping} мс
                      </div>
                    )}
                    <div className={styles.detailRow}>
                      <strong title="Видеокарта">Видеокарта:</strong>
                      {activeMerchantDetails.gpus && activeMerchantDetails.gpus.length > 0 ? (
                        (() => {
                          const g = activeMerchantDetails.gpus[0]
                          return `${g.name} — видеопамять ${ramKbToGb(g.available_vram).toFixed(1)} ГБ${g.dlperf != null ? `, производительность ${g.dlperf}` : ''}`
                        })()
                      ) : (
                        'отсутствует'
                      )}
                    </div>
                    {activeMerchantDetails.cpus && activeMerchantDetails.cpus.length > 0 && (
                      <div className={styles.detailRow}>
                        <strong title="Процессор">Процессор:</strong>{' '}
                        {activeMerchantDetails.cpus[0].name} — {activeMerchantDetails.cpus[0].available} ядер
                      </div>
                    )}
                    {activeMerchantDetails.storages && activeMerchantDetails.storages.length > 0 && (
                      <div className={styles.detailBlock}>
                        <strong>Хранилища:</strong>
                        <ul>
                          {activeMerchantDetails.storages.map((s, i) => (
                            <li key={i}>
                              {s.type} {s.name} — {storageToGb(s.available).toFixed(1)} ГБ
                            </li>
                          ))}
                        </ul>
                      </div>
                    )}
                  </div>
                )}
              </div>
            )}
            <button
              type="button"
              className={styles.toggleBlock}
              onClick={() => setShowActiveSettings((v) => !v)}
              aria-expanded={showActiveSettings}
            >
                {showActiveSettings ? '▼' : '▶'} Ссылка и данные для входа
            </button>
            {showActiveSettings && (
              <div className={styles.activeRentBlockContent}>
                {isConfiguring && (
                  <p className={styles.muted}>Данные для подключения появятся после того, как поставщик запустит сеанс.</p>
                )}
                {isStarted && (
                  <>
                    {settingsLoading && <p className={styles.muted}>Загрузка данных подключения…</p>}
                    {!settingsLoading && (() => {
                      const s = activeRent.settings ?? activeRentSettings
                      const login = s?.template?.login ?? s?.template?.authentication?.login
                      const password = s?.template?.password ?? s?.template?.authentication?.password
                      const firstEndpoint = s?.network?.piko?.endpoints?.[0]
                      const hasAny = login != null || password != null || (firstEndpoint?.link != null)
                      if (!hasAny) {
                        return <p className={styles.muted}>Данные для подключения пока не получены от поставщика.</p>
                      }
                      return (
                        <>
                          {(login != null || password != null) && (
                            <>
                              <div className={styles.cardRow}>
                                <span className={styles.label}>Логин:</span>
                                <span>{login ?? '—'}</span>
                              </div>
                              <div className={styles.cardRow}>
                                <span className={styles.label}>Пароль:</span>
                                <span>{password ?? '—'}</span>
                              </div>
                            </>
                          )}
                          {firstEndpoint?.link && (
                            <div className={styles.cardRow}>
                              <span className={styles.label}>Ссылка:</span>
                              <a href={firstEndpoint.link} target="_blank" rel="noopener noreferrer">
                                {firstEndpoint.link}
                              </a>
                            </div>
                          )}
                        </>
                      )
                    })()}
                  </>
                )}
              </div>
            )}
            <div className={styles.activeRentActions}>
            {!stopConfirm ? (
              <button
                type="button"
                className={styles.dangerBtn}
                onClick={() => setStopConfirm(true)}
                title="Завершить текущий сеанс аренды. Сеанс будет остановлен, списание прекратится."
              >
                Завершить сеанс
              </button>
            ) : (
              <div className={styles.stopConfirm}>
                <p className={styles.hint}>Сеанс будет завершён. Причину можно указать по желанию.</p>
                <label>
                  Причина завершения (необязательно):
                  <input
                    type="text"
                    value={stopReason}
                    onChange={(e) => setStopReason(e.target.value)}
                    placeholder="Причина остановки"
                    className={styles.input}
                  />
                </label>
                {stopError && <div className={styles.errorBlock}>{stopError}</div>}
                <div className={styles.buttons}>
                  <button
                    type="button"
                    className={styles.primaryBtn}
                    onClick={handleStop}
                    disabled={stopSubmitting}
                  >
                    {stopSubmitting ? 'Остановка…' : 'Подтвердить остановку'}
                  </button>
                  <button
                    type="button"
                    className={styles.secondaryBtn}
                    onClick={() => { setStopConfirm(false); setStopError(null); }}
                    disabled={stopSubmitting}
                  >
                    Отмена
                  </button>
                </div>
              </div>
            )}
            </div>
            </div>
          </div>
        )}
        {!loadingActive && !activeError && !hasActive && (
          <div className={styles.noActive}>
            <p>Сейчас у вас нет активной аренды.</p>
            <p className={styles.hint}>Выберите поставщика и шаблон, затем нажмите «Запуск».</p>
            <div>
              <button
                type="button"
                className={styles.primaryBtn}
                onClick={openStartModal}
                disabled={startSubmitting}
                title="Открыть выбор поставщика и шаблона для новой аренды"
              >
                Оформить аренду
              </button>
            </div>
          </div>
        )}
      </section>

      <section className={styles.section} aria-labelledby="history-heading">
        <h3 id="history-heading">История аренд</h3>
        <p className={styles.sectionHint}>
          График по дням: ось X — дата, ось Y — количество сеансов. Нажмите на столбец дня, чтобы увидеть детали.
        </p>
        {loadingHistory && <p className={styles.muted}>Загрузка…</p>}
        {historyError && <div className={styles.errorBlock}>{historyError}</div>}
        {!loadingHistory && history.length === 0 && (
          <p className={styles.muted}>Нет записей в истории.</p>
        )}
        {!loadingHistory && history.length > 0 && (
          <div className={styles.historyChartWrap}>
            <div className={styles.historyChartControls}>
              <span className={styles.historyChartLabel}>Период:</span>
              <div className={styles.historyRangeButtons}>
                {RANGE_OPTIONS.map((n) => (
                  <button
                    key={n}
                    type="button"
                    className={historyRangeDays === n ? styles.primaryBtn : styles.secondaryBtn}
                    onClick={() => {
                      setHistoryRangeDays(n)
                      setSelectedHistoryDay(null)
                    }}
                  >
                    {n} дн.
                  </button>
                ))}
              </div>
              <div className={styles.historyNavButtons}>
                <button
                  type="button"
                  className={styles.secondaryBtn}
                  onClick={() => {
                    setHistoryWindowEnd((prev) => addDays(prev, -historyRangeDays))
                    setSelectedHistoryDay(null)
                  }}
                  title="Более ранний период"
                >
                  ← Ранее
                </button>
                <button
                  type="button"
                  className={styles.secondaryBtn}
                  disabled={!canHistoryGoLater}
                  onClick={() => {
                    setHistoryWindowEnd((prev) => {
                      const next = addDays(prev, historyRangeDays)
                      return next > today ? today : next
                    })
                    setSelectedHistoryDay(null)
                  }}
                  title="Более поздний период"
                >
                  Позже →
                </button>
              </div>
            </div>
            <div className={styles.historyChartContainer}>
              <ResponsiveContainer width="100%" height={280}>
                <BarChart
                  data={historyChartData}
                  margin={{ top: 12, right: 16, left: 8, bottom: 8 }}
                >
                  <XAxis
                    dataKey="dayLabel"
                    tick={{ fill: 'var(--muted)', fontSize: 12 }}
                    tickLine={{ stroke: 'var(--border)' }}
                    axisLine={{ stroke: 'var(--border)' }}
                  />
                  <YAxis
                    allowDecimals={false}
                    tick={{ fill: 'var(--muted)', fontSize: 12 }}
                    tickLine={{ stroke: 'var(--border)' }}
                    axisLine={{ stroke: 'var(--border)' }}
                    label={{ value: 'Сеансов', angle: -90, position: 'insideLeft', fill: 'var(--muted)', fontSize: 11 }}
                  />
                  <Tooltip
                    contentStyle={{
                      background: 'var(--surface)',
                      border: '1px solid var(--border)',
                      borderRadius: 'var(--radius)',
                      color: 'var(--text)',
                    }}
                    labelStyle={{ color: 'var(--text)' }}
                    itemStyle={{ color: 'var(--text)' }}
                    formatter={(value: number) => [`${value} сеанс(ов)`, 'Количество']}
                    labelFormatter={(label: string, payload: { payload?: { dateKey?: string } }[]) => payload[0]?.payload?.dateKey ? new Date(payload[0].payload.dateKey).toLocaleDateString('ru-RU', { day: '2-digit', month: 'long', year: 'numeric' }) : label}
                  />
                  <Bar
                    dataKey="count"
                    name="Сеансов"
                    radius={[4, 4, 0, 0]}
                    cursor="pointer"
                    maxBarSize={48}
                    onClick={(data: { dateKey?: string }) => {
                      const key = data?.dateKey
                      if (key) setSelectedHistoryDay((prev) => (prev === key ? null : key))
                    }}
                  >
                    {historyChartData.map((entry) => (
                      <Cell
                        key={entry.dateKey}
                        fill={selectedHistoryDay === entry.dateKey ? 'var(--accent)' : 'var(--accent)'}
                        opacity={selectedHistoryDay === entry.dateKey ? 1 : 0.7}
                        stroke={selectedHistoryDay === entry.dateKey ? 'var(--text)' : 'transparent'}
                        strokeWidth={2}
                      />
                    ))}
                  </Bar>
                </BarChart>
              </ResponsiveContainer>
            </div>
            {selectedHistoryDay && selectedDayInfo && (
              <div className={styles.historyDayDetail}>
                <h4 className={styles.historyDayDetailTitle}>
                  {new Date(selectedHistoryDay).toLocaleDateString('ru-RU', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' })}
                </h4>
                <div className={styles.historyDayDetailStats}>
                  <span title="Сумма за день в белорусских рублях">Потрачено денег: <strong>{selectedDayInfo.totalCost.toFixed(2)} BYN</strong></span>
                  <span title="Количество сеансов">Сеансов: <strong>{selectedDayInfo.items.length}</strong></span>
                </div>
                <h5 className={styles.historyDayDetailListTitle}>Сеансы за этот день</h5>
                <div className={styles.historyDayDetailTableWrap}>
                  <div className={styles.historyDayDetailListHeader}>
                    <span>Время начала</span>
                    <span>Название шаблона</span>
                    <span>Потрачено денег (BYN)</span>
                    <span>Время аренды</span>
                  </div>
                  <ul className={styles.historyDayDetailList} aria-label="Список сеансов за день">
                    {selectedDayInfo.items.map((r) => (
                      <li key={r.id} className={styles.historyDayDetailItem}>
                        <span className={styles.historyDayDetailTime}>
                          {r.started_at
                            ? new Date(r.started_at).toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })
                            : '—'}
                        </span>
                        <span>
                          <button
                            type="button"
                            className={styles.templateNameLink}
                            onClick={() => openHistoryTemplatePreview(r)}
                          >
                            {r.template_title || r.template_id || '—'}
                          </button>
                        </span>
                        <span className={styles.historyDayDetailCost}>{r.cost} BYN</span>
                        <span className={styles.muted}>{r.duration} мин</span>
                      </li>
                    ))}
                  </ul>
                </div>
              </div>
            )}
          </div>
        )}
      </section>

      {startModalOpen && (
        <div className={styles.modalBackdrop} onClick={() => !startSubmitting && setStartModalOpen(false)}>
          <div className={styles.modal} onClick={(e) => e.stopPropagation()}>
            {startStep === 'merchants' ? (
              <>
                <div className={styles.modalHeader}>
                  <h3 className={styles.modalTitle}>Выберите поставщика</h3>
                  <button
                    type="button"
                    className={styles.modalClose}
                    onClick={() => !startSubmitting && setStartModalOpen(false)}
                    disabled={startSubmitting}
                    aria-label="Закрыть"
                  >
                    ×
                  </button>
                </div>
                <div className={styles.modalBody}>
                  <div className={styles.templateExplain}>
                    <p className={styles.templateExplainText}>
                      Поставщик — это человек или компания, которые отдают в аренду свой компьютер. Выберите поставщика, при необходимости раскройте «Характеристики», чтобы посмотреть параметры компьютера.
                    </p>
                  </div>
                  {merchantsLoading && <p className={styles.muted}>Загрузка списка…</p>}
                  {merchantsError && <div className={styles.errorBlock}>{merchantsError}</div>}
                  {!merchantsLoading && !merchantsError && merchants.length === 0 && (
                    <p className={styles.muted}>Сейчас нет поставщиков, готовых принять аренду. Попробуйте позже.</p>
                  )}
                  {!merchantsLoading && merchants.length > 0 && (
                    <ul className={styles.merchantList}>
                      {merchants.map((m) => (
                        <li key={m.session_id} className={styles.merchantCard}>
                          <div className={styles.merchantCardTop}>
                            <div className={styles.merchantCardInfo}>
                              <span className={styles.merchantName}>
                                {m.node_name || `Узел ${m.session_id.slice(0, 8)}…`} ({m.name || '—'})
                              </span>
                              {m.details != null && (
                                <span className={styles.merchantPriceLine}>
                                  <span className={styles.merchantPriceLabel}>Цена:</span>{' '}
                                  <span className={styles.merchantPrice}>{m.details.total_price} BYN/мин</span>
                                </span>
                              )}
                            </div>
                            <button
                              type="button"
                              className={styles.primaryBtn}
                              onClick={() => openTemplateStep(m)}
                              disabled={startSubmitting}
                            >
                              Выбрать
                            </button>
                          </div>
                          <button
                            type="button"
                            className={styles.toggleBlock}
                            onClick={() => setExpandedMerchantId((id) => (id === m.session_id ? null : m.session_id))}
                            aria-expanded={expandedMerchantId === m.session_id}
                          >
                            {expandedMerchantId === m.session_id ? '▼' : '▶'} Характеристики
                          </button>
                          {expandedMerchantId === m.session_id && m.details && (
                        <div className={styles.merchantDetails}>
                          <div className={styles.detailRow}>
                            <strong title="Объём оперативной памяти">Оперативная память:</strong>{' '}
                            {ramKbToGb(m.details.total_ram).toFixed(1)} ГБ
                          </div>
                          {m.details.load_speed != null && (
                            <div className={styles.detailRow}>
                              <strong>Скорость загрузки:</strong> {networkSpeedToMbit(m.details.load_speed).toFixed(1)} Мбит/с
                            </div>
                          )}
                          {m.details.upload_speed != null && (
                            <div className={styles.detailRow}>
                              <strong>Скорость отдачи:</strong> {networkSpeedToMbit(m.details.upload_speed).toFixed(1)} Мбит/с
                            </div>
                          )}
                          {m.details.ping != null && (
                            <div className={styles.detailRow}>
                              <strong title="Сетевая задержка">Задержка:</strong> {m.details.ping} мс
                            </div>
                          )}
                          <div className={styles.detailRow}>
                            <strong title="Видеокарта">Видеокарта:</strong>
                            {m.details.gpus && m.details.gpus.length > 0 ? (
                              (() => {
                                const g = m.details.gpus[0]
                                return `${g.name} — видеопамять ${ramKbToGb(g.available_vram).toFixed(1)} ГБ${g.dlperf != null ? `, производительность ${g.dlperf}` : ''}`
                              })()
                            ) : (
                              'отсутствует'
                            )}
                          </div>
                          {m.details.cpus && m.details.cpus.length > 0 && (
                            <div className={styles.detailRow}>
                              <strong title="Процессор">Процессор:</strong>{' '}
                              {m.details.cpus[0].name} — {m.details.cpus[0].available} ядер
                            </div>
                          )}
                          {m.details.storages && m.details.storages.length > 0 && (
                            <div className={styles.detailBlock}>
                              <strong>Хранилища:</strong>
                              <ul>
                                {m.details.storages.map((s, i) => (
                                  <li key={i}>
                                    {s.type} {s.name} — {storageToGb(s.available).toFixed(1)} ГБ
                                  </li>
                                ))}
                              </ul>
                            </div>
                          )}
                        </div>
                      )}
                    </li>
                  ))}
                </ul>
              )}
                  {startError && <div className={styles.errorBlock}>{startError}</div>}
                </div>
              </>
            ) : (
              <>
                <div className={styles.modalHeader}>
                  <h3>Выберите шаблон</h3>
                  <div className={styles.modalHeaderActions}>
                    <button
                      type="button"
                      className={styles.secondaryBtn}
                      onClick={goBackToMerchants}
                      disabled={startSubmitting}
                    >
                      Назад
                    </button>
                    <button
                      type="button"
                      className={styles.modalClose}
                      onClick={() => !startSubmitting && setStartModalOpen(false)}
                      disabled={startSubmitting}
                      aria-label="Закрыть"
                    >
                      ×
                    </button>
                  </div>
                </div>
                <div className={styles.modalBody}>
                  {selectedMerchant && (
                    <>
                      <div className={styles.templateExplain}>
                        <p className={styles.templateExplainText}>
                          Шаблон — это готовый рабочий стол с системой и программами. Выберите один вариант и нажмите «Запуск».
                        </p>
                      </div>
                      {templatesLoading && <p className={styles.muted}>Загрузка списка шаблонов…</p>}
                      {templatesError && <div className={styles.errorBlock}>{templatesError}</div>}
                      {!templatesLoading && !templatesError && templates.length === 0 && (
                        <p className={styles.muted}>Нет доступных шаблонов у этого поставщика.</p>
                      )}
                      {!templatesLoading && templates.length > 0 && (
                        <div className={styles.templateGrid}>
                          {templates.map((t) => (
                            <label
                              key={t.template_id}
                              className={`${styles.templateCard} ${selectedTemplateId === t.template_id ? styles.templateCardSelected : ''}`}
                            >
                              <input
                                type="radio"
                                name="template"
                                checked={selectedTemplateId === t.template_id}
                                onChange={() => setSelectedTemplateId(t.template_id)}
                                className={styles.templateCardRadio}
                              />
                              {t.short_description && (
                                <div className={styles.templateCardImage}>
                                  <img src={t.short_description} alt={t.title || t.template_id} />
                                </div>
                              )}
                              <div className={styles.templateCardContent}>
                                <h4 className={styles.templateCardTitle}>{t.title || t.template_id}</h4>
                                {t.description && (
                                  <p className={styles.templateCardDescription}>{t.description}</p>
                                )}
                              </div>
                            </label>
                          ))}
                        </div>
                      )}
                      <div className={styles.buttons} style={{ marginTop: '1rem' }}>
                        <button
                          type="button"
                          className={styles.primaryBtn}
                          onClick={handleStartSession}
                          disabled={startSubmitting || templates.length === 0 || !selectedTemplateId}
                        >
                          {startSubmitting ? 'Запуск…' : 'Запуск'}
                        </button>
                      </div>
                    </>
                  )}
                  {startError && <div className={styles.errorBlock}>{startError}</div>}
                </div>
              </>
            )}
          </div>
        </div>
      )}

      {historyTemplatePreview && (
        <div className={styles.modalBackdrop} onClick={() => setHistoryTemplatePreview(null)}>
          <div className={styles.modal} onClick={(e) => e.stopPropagation()}>
            <div className={styles.modalHeader}>
              <h3 className={styles.modalTitle}>Шаблон</h3>
              <button
                type="button"
                className={styles.modalClose}
                onClick={() => setHistoryTemplatePreview(null)}
                aria-label="Закрыть"
              >
                ×
              </button>
            </div>
            <div className={styles.modalBody}>
              {historyTemplatePreviewLoading && (
                <p className={styles.muted}>Загрузка…</p>
              )}
              {!historyTemplatePreviewLoading && historyTemplatePreview.templateInfo && (
                <div className={styles.templatePreviewCard}>
                  {historyTemplatePreview.templateInfo.short_description && (
                    <div className={styles.templateCardImage}>
                      <img
                        src={historyTemplatePreview.templateInfo.short_description}
                        alt={historyTemplatePreview.templateInfo.title || historyTemplatePreview.templateInfo.template_id}
                      />
                    </div>
                  )}
                  <div className={styles.templateCardContent}>
                    <h4 className={styles.templateCardTitle}>
                      {historyTemplatePreview.templateInfo.title || historyTemplatePreview.templateInfo.template_id}
                    </h4>
                    {historyTemplatePreview.templateInfo.description && (
                      <p className={styles.templateCardDescription}>
                        {historyTemplatePreview.templateInfo.description}
                      </p>
                    )}
                  </div>
                </div>
              )}
              {!historyTemplatePreviewLoading && !historyTemplatePreview.templateInfo && (
                <p className={styles.muted}>
                  {historyTemplatePreview.item.template_title || historyTemplatePreview.item.template_id || '—'}. Описание шаблона недоступно.
                </p>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
