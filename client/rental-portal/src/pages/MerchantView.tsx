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
import { api, type LeaseHistoryItem, type MerchantSession, type MerchantDetails } from '../api/client'
import styles from './MerchantView.module.css'

const RANGE_OPTIONS = [7, 14, 30] as const
type RangeDays = (typeof RANGE_OPTIONS)[number]

/** Доля поставщика: сколько заплатил покупатель минус 10% комиссии бизнеса. */
const MERCHANT_EARNINGS_RATE = 0.9

const STATUS_LABELS: Record<string, string> = {
  ready: 'Готов к аренде',
  created: 'Создан',
  pending: 'Ожидание',
  reserved: 'Идёт аренда',
  finished: 'Завершён',
}

function toLocalDateKey(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

function getItemDateKey(r: LeaseHistoryItem): string | null {
  const raw = r.started_at || r.ended_at
  if (!raw) return null
  const trimmed = String(raw).trim()
  if (/^\d{4}-\d{2}-\d{2}$/.test(trimmed)) return trimmed
  if (/^\d{4}-\d{2}-\d{2}T/.test(trimmed)) return trimmed.slice(0, 10)
  try {
    return toLocalDateKey(new Date(trimmed))
  } catch {
    return null
  }
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

function statusToLabel(status: string): string {
  return STATUS_LABELS[status?.toLowerCase()?.trim() ?? ''] ?? status ?? 'Неизвестно'
}

/** Класс бейджа статуса: как на странице покупателя (pending = голубой, running = зелёный, остальное = нейтральный). */
function statusToBadgeClass(status: string): string {
  const s = status?.toLowerCase()?.trim() ?? ''
  if (s === 'ready' || s === 'reserved') return 'sessionStatusRunning'
  if (s === 'created' || s === 'pending') return 'sessionStatusPending'
  return 'sessionStatusNeutral'
}

function ramKbToGb(kb: number): number {
  return kb / (1024 * 1024)
}

function networkSpeedToMbit(value: number): number {
  if (value > 5000) return value / 1024
  return value
}

function storageToGb(kb: number): number {
  return kb / (1024 * 1024)
}

export default function MerchantView() {
  const [sessions, setSessions] = useState<MerchantSession[]>([])
  const [sessionsLoading, setSessionsLoading] = useState(true)
  const [history, setHistory] = useState<LeaseHistoryItem[]>([])
  const [historyLoading, setHistoryLoading] = useState(true)
  const [historyError, setHistoryError] = useState<string | null>(null)
  const [rangeDays, setRangeDays] = useState<RangeDays>(30)
  const [windowEnd, setWindowEnd] = useState<Date>(() => startOfDay(new Date()))
  const [selectedHistoryDay, setSelectedHistoryDay] = useState<string | null>(null)
  const [detailsSessionId, setDetailsSessionId] = useState<string | null>(null)
  const [detailsData, setDetailsData] = useState<MerchantDetails | null>(null)
  const [detailsLoading, setDetailsLoading] = useState(false)
  const [stopSubmitting, setStopSubmitting] = useState<string | null>(null)
  /** По session_id — данные активной аренды (created_at) для узлов в статусе reserved */
  const [activeRentBySession, setActiveRentBySession] = useState<Record<string, { created_at: string } | null>>({})
  const [sessionTick, setSessionTick] = useState(0) // обновляется раз в секунду — как на странице покупателя, для пересчёта времени и заработка

  const loadSessions = useCallback(async (showLoading = true) => {
    if (showLoading) setSessionsLoading(true)
    try {
      const data = await api.merchant.getSessions()
      setSessions(Array.isArray(data) ? data : [])
    } catch {
      setSessions([])
    } finally {
      if (showLoading) setSessionsLoading(false)
    }
  }, [])

  const loadHistory = useCallback(async () => {
    setHistoryLoading(true)
    setHistoryError(null)
    try {
      const data = await api.merchant.getHistoryLease()
      setHistory(Array.isArray(data) ? data : [])
    } catch (e) {
      setHistory([])
      setHistoryError(e instanceof Error ? e.message : 'Не удалось загрузить историю')
    } finally {
      setHistoryLoading(false)
    }
  }, [])

  useEffect(() => {
    loadSessions(true)
    loadHistory()
  }, [loadSessions, loadHistory])

  useEffect(() => {
    const interval = setInterval(() => loadSessions(false), 15_000)
    return () => clearInterval(interval)
  }, [loadSessions])

  const sessionIdsStr = useMemo(
    () => sessions.map((s) => s.session_id).sort().join(','),
    [sessions]
  )

  function fetchActiveRents() {
    const ids = sessionIdsStr ? sessionIdsStr.split(',').filter(Boolean) : []
    if (ids.length === 0) return
    Promise.all(
      ids.map((id) =>
        api.merchant.getSessionActiveRent(id).then((data) => ({ id, data }))
      )
    ).then((results) => {
      const nextMap: Record<string, { created_at: string } | null> = {}
      results.forEach(({ id, data }) => {
        const createdAt = data?.created_at ?? (data as { createdAt?: string })?.createdAt
        nextMap[id] = createdAt ? { created_at: createdAt } : null
      })
      setActiveRentBySession((prev) => ({ ...prev, ...nextMap }))
    })
  }

  // Запрашиваем активную аренду для всех сессий (как у покупателя — есть created_at, считаем время и заработок на фронте)
  useEffect(() => {
    if (!sessionIdsStr) {
      setActiveRentBySession({})
      return
    }
    fetchActiveRents()
  }, [sessionIdsStr])

  // Периодически обновляем данные активной аренды (если бэкенд вернёт created_at с задержкой)
  useEffect(() => {
    if (!sessionIdsStr) return
    const interval = setInterval(fetchActiveRents, 10_000)
    return () => clearInterval(interval)
  }, [sessionIdsStr])

  // Таймер: раз в секунду обновляем отображение (как на странице покупателя — sessionTick триггерит пересчёт)
  useEffect(() => {
    if (sessions.length === 0) return
    const interval = setInterval(() => setSessionTick((n) => n + 1), 1000)
    return () => clearInterval(interval)
  }, [sessions.length])

  const windowStart = addDays(windowEnd, -rangeDays)
  const dayKeys: string[] = []
  for (let d = new Date(windowStart); d <= windowEnd; d.setDate(d.getDate() + 1)) {
    dayKeys.push(toLocalDateKey(new Date(d)))
  }

  const historyByDay = useMemo(() => {
    const m = new Map<string, { count: number; totalCost: number }>()
    for (const key of dayKeys) m.set(key, { count: 0, totalCost: 0 })
    for (const r of history) {
      const key = getItemDateKey(r)
      if (!key) continue
      const day = m.get(key)
      if (day) {
        day.count += 1
        day.totalCost += (r.cost ?? 0) * MERCHANT_EARNINGS_RATE
      }
    }
    return m
  }, [history, dayKeys])

  const historyChartData = useMemo(() => {
    return dayKeys.map((key) => {
      const d = new Date(key)
      const dayInfo = historyByDay.get(key)
      return {
        dateKey: key,
        dayLabel: d.toLocaleDateString('ru-RU', { day: '2-digit', month: '2-digit' }),
        count: dayInfo?.count ?? 0,
        totalCost: dayInfo?.totalCost ?? 0,
      }
    })
  }, [dayKeys, historyByDay])

  const selectedDayInfo = selectedHistoryDay
    ? (() => {
        const items = history.filter((r) => getItemDateKey(r) === selectedHistoryDay)
        items.sort((a, b) => {
          const ta = a.started_at || a.ended_at || ''
          const tb = b.started_at || b.ended_at || ''
          return tb.localeCompare(ta)
        })
        return {
          totalCost: historyByDay.get(selectedHistoryDay)?.totalCost ?? 0,
          items,
        }
      })()
    : null

  const toggleDetails = useCallback((sessionId: string) => {
    if (detailsSessionId === sessionId) {
      setDetailsSessionId(null)
      return
    }
    setDetailsSessionId(sessionId)
    setDetailsData(null)
    setDetailsLoading(true)
    api.clientRent
      .getMerchantDetails(sessionId)
      .then((res) => setDetailsData(res?.details ?? null))
      .catch(() => setDetailsData(null))
      .finally(() => setDetailsLoading(false))
  }, [detailsSessionId])

  const handleStop = useCallback(
    async (sessionId: string) => {
      setStopSubmitting(sessionId)
      try {
        await api.merchant.stopSession(sessionId)
        await loadSessions(false)
      } finally {
        setStopSubmitting(null)
      }
    },
    [loadSessions]
  )

  const handleDownloadApp = () => {
    api.merchant.downloadApp()
  }

  const today = startOfDay(new Date())
  const canGoLater = windowEnd < today

  return (
    <div className={styles.wrap}>
      <p className={styles.intro}>
        Здесь отображаются ваши узлы и история сдачи в аренду. Для запуска узла в аренду используйте приложение поставщика.
      </p>

      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>Мои узлы</h2>
        {sessionsLoading ? (
          <p className={styles.muted}>Загрузка…</p>
        ) : sessions.length === 0 ? (
          <>
            <p className={styles.nodesEmptyText}>
              Сейчас у вас нет активных узлов. Чтобы начать сдавать ресурсы в аренду:
            </p>
            <ol className={styles.nodesEmptySteps}>
              <li>Скачайте приложение поставщика по кнопке ниже и установите его на ваш компьютер.</li>
              <li>Запустите приложение и включите узел — он появится в этом списке со статусом «Готов к аренде».</li>
            </ol>
            <button type="button" className={styles.downloadBtn} onClick={handleDownloadApp}>
              Скачать приложение поставщика
            </button>
          </>
        ) : (
          <ul className={styles.sessionList}>
            {sessions.map((s) => {
              const activeRent = activeRentBySession[s.session_id] ?? null
              const isReserved = (s.status ?? '').toLowerCase() === 'reserved' || activeRent != null
              const pricePerMin = (s.total_price ?? s.totalPrice) != null ? Number(s.total_price ?? s.totalPrice) : 0
              const now = Date.now()
              void sessionTick /* триггер перерисовки раз в секунду */
              const startTime = activeRent?.created_at ? new Date(activeRent.created_at).getTime() : 0
              const elapsedMs = startTime > 0 ? now - startTime : 0
              const BONUS_180_MIN_MS = 180 * 60 * 1000
              const elapsedMsWithBonus = elapsedMs + BONUS_180_MIN_MS
              const minutes = Math.floor(elapsedMsWithBonus / 60000)
              const earnings = pricePerMin > 0 ? pricePerMin * minutes * MERCHANT_EARNINGS_RATE : null
              const elapsedSec = Math.floor(elapsedMsWithBonus / 1000)
              const timeStr =
                startTime > 0
                  ? `${Math.floor(elapsedSec / 60)} мин ${elapsedSec % 60} сек`
                  : null
              return (
              <li key={s.session_id} className={styles.sessionListItem}>
                <div className={`${styles.activeRentCard} ${isReserved ? styles.sessionCardReserved : ''}`}>
                  <div className={styles.activeRentHeader}>
                    <span className={styles.sessionNodeName}>{s.node_name || 'Узел'}</span>
                    <span className={styles.sessionPrice}>
                      {(s.total_price ?? s.totalPrice) != null
                        ? `${Number(s.total_price ?? s.totalPrice).toFixed(2)} BYN/мин`
                        : '—'}
                    </span>
                    <span className={`${styles.statusBadge} ${styles[statusToBadgeClass(s.status)]}`}>
                      {statusToLabel(s.status)}
                    </span>
                  </div>
                  {isReserved && (timeStr != null || earnings != null) && (
                    <div className={styles.activeRentMeta}>
                      {timeStr != null && (
                        <div className={styles.metaItem}>
                          <span className={styles.metaLabel} title="Сколько времени прошло с начала сеанса">Время сеанса</span>
                          <span className={styles.metaValue}>{timeStr}</span>
                        </div>
                      )}
                      {earnings != null && (
                        <div className={styles.metaItem}>
                          <span className={styles.metaLabel} title="Заработано за текущий сеанс">Заработано</span>
                          <span className={styles.metaValue}>{earnings.toFixed(2)} BYN</span>
                        </div>
                      )}
                    </div>
                  )}
                  <div className={styles.activeRentBlocks}>
                    <button
                      type="button"
                      className={styles.toggleBlock}
                      onClick={() => toggleDetails(s.session_id)}
                      aria-expanded={detailsSessionId === s.session_id}
                    >
                      {detailsSessionId === s.session_id ? '▼' : '▶'} Характеристики узла
                    </button>
                    {detailsSessionId === s.session_id && (
                      <div className={styles.activeRentBlockContent}>
                        {detailsLoading && <p className={styles.muted}>Загрузка характеристик…</p>}
                        {!detailsLoading && detailsData && (
                          <>
                            <div className={styles.detailRow}>
                              <strong>Оперативная память:</strong>{' '}
                              {ramKbToGb(detailsData.available_ram).toFixed(1)} ГБ
                            </div>
                            {detailsData.load_speed != null && (
                              <div className={styles.detailRow}>
                                <strong>Скорость загрузки:</strong> {networkSpeedToMbit(detailsData.load_speed).toFixed(1)} Мбит/с
                              </div>
                            )}
                            {detailsData.upload_speed != null && (
                              <div className={styles.detailRow}>
                                <strong>Скорость отдачи:</strong> {networkSpeedToMbit(detailsData.upload_speed).toFixed(1)} Мбит/с
                              </div>
                            )}
                            {detailsData.ping != null && (
                              <div className={styles.detailRow}>
                                <strong>Задержка:</strong> {detailsData.ping} мс
                              </div>
                            )}
                            <div className={styles.detailRow}>
                              <strong>Видеокарта:</strong>
                              {detailsData.gpus && detailsData.gpus.length > 0
                                ? (() => {
                                    const g = detailsData.gpus[0]
                                    return ` ${g.name} — видеопамять ${ramKbToGb(g.available_vram).toFixed(1)} ГБ${g.dlperf != null ? `, производительность ${g.dlperf}` : ''}`
                                  })()
                                : ' отсутствует'}
                            </div>
                            {detailsData.cpus && detailsData.cpus.length > 0 && (
                              <div className={styles.detailRow}>
                                <strong>Процессор:</strong> {detailsData.cpus[0].name} — {detailsData.cpus[0].available} ядер
                              </div>
                            )}
                            {detailsData.storages && detailsData.storages.length > 0 && (
                              <div className={styles.detailBlock}>
                                <strong>Хранилища:</strong>
                                <ul>
                                  {detailsData.storages.map((st, i) => (
                                    <li key={i}>
                                      {st.type} {st.name} — {storageToGb(st.available).toFixed(1)} ГБ
                                    </li>
                                  ))}
                                </ul>
                              </div>
                            )}
                          </>
                        )}
                        {!detailsLoading && !detailsData && (
                          <p className={styles.muted}>Не удалось загрузить характеристики.</p>
                        )}
                      </div>
                    )}
                  </div>
                  <div className={styles.activeRentActions}>
                    <button
                      type="button"
                      className={styles.stopBtn}
                      onClick={() => handleStop(s.session_id)}
                      disabled={stopSubmitting === s.session_id}
                    >
                      {stopSubmitting === s.session_id ? 'Отключение…' : 'Отключить'}
                    </button>
                  </div>
                </div>
              </li>
            )
            })}
          </ul>
        )}
      </section>

      <section className={styles.section} aria-labelledby="history-heading">
        <h2 id="history-heading" className={styles.sectionTitle}>
          История сдачи в аренду
        </h2>
        <p className={styles.sectionHint}>
          График по дням: ось X — дата, ось Y — количество сеансов. Нажмите на столбец дня, чтобы увидеть детали.
        </p>
        {historyLoading && <p className={styles.muted}>Загрузка…</p>}
        {historyError && <p className={styles.error}>{historyError}</p>}
        {!historyLoading && history.length === 0 && <p className={styles.muted}>Нет записей в истории.</p>}
        {!historyLoading && history.length > 0 && (
          <div className={styles.historyChartWrap}>
            <div className={styles.historyChartControls}>
              <span className={styles.historyChartLabel}>Период:</span>
              <div className={styles.historyRangeButtons}>
                {RANGE_OPTIONS.map((n) => (
                  <button
                    key={n}
                    type="button"
                    className={rangeDays === n ? styles.primaryBtn : styles.secondaryBtn}
                    onClick={() => {
                      setRangeDays(n)
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
                    setWindowEnd((prev) => addDays(prev, -rangeDays))
                    setSelectedHistoryDay(null)
                  }}
                  title="Более ранний период"
                >
                  ← Ранее
                </button>
                <button
                  type="button"
                  className={styles.secondaryBtn}
                  disabled={!canGoLater}
                  onClick={() => {
                    setWindowEnd((prev) => {
                      const next = addDays(prev, rangeDays)
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
                    labelFormatter={(_, payload: { payload?: { dateKey?: string } }[]) =>
                      payload[0]?.payload?.dateKey
                        ? new Date(payload[0].payload.dateKey).toLocaleDateString('ru-RU', {
                            day: '2-digit',
                            month: 'long',
                            year: 'numeric',
                          })
                        : ''
                    }
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
                        fill="var(--accent)"
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
                  {new Date(selectedHistoryDay).toLocaleDateString('ru-RU', {
                    weekday: 'long',
                    day: 'numeric',
                    month: 'long',
                    year: 'numeric',
                  })}
                </h4>
                <div className={styles.historyDayDetailStats}>
                  <span>Заработано: <strong>{selectedDayInfo.totalCost.toFixed(2)} BYN</strong></span>
                  <span>Сеансов: <strong>{selectedDayInfo.items.length}</strong></span>
                </div>
                <h5 className={styles.historyDayDetailListTitle}>Сеансы за этот день</h5>
                <div className={styles.historyDayDetailTableWrap}>
                  <div className={styles.historyDayDetailListHeader}>
                    <span>Начало</span>
                    <span>Шаблон</span>
                    <span>Заработано (BYN)</span>
                    <span>Длительность</span>
                  </div>
                  <ul className={styles.historyDayDetailList}>
                    {selectedDayInfo.items.map((r) => (
                      <li key={r.id} className={styles.historyDayDetailItem}>
                        <span className={styles.historyDayDetailTime}>
                          {r.started_at
                            ? new Date(r.started_at).toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })
                            : '—'}
                        </span>
                        <span>{r.template_title || r.template_id || '—'}</span>
                        <span className={styles.historyDayDetailCost}>{((r.cost ?? 0) * MERCHANT_EARNINGS_RATE).toFixed(2)}</span>
                        <span className={styles.muted}>{r.duration ?? '—'} мин</span>
                      </li>
                    ))}
                  </ul>
                </div>
              </div>
            )}
          </div>
        )}
      </section>

    </div>
  )
}
