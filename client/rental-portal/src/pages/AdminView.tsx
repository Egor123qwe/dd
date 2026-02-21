import { useState, useEffect, useCallback, useMemo, useRef } from 'react'
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  Cell,
} from 'recharts'
import { api, type AdminRentHistoryItem, type TemplateInfo, type TemplatePort, type TemplateEnv } from '../api/client'
import styles from './AdminView.module.css'
import sharedStyles from './ClientView.module.css'

const RANGE_OPTIONS = [7, 14, 30] as const
type RangeDays = (typeof RANGE_OPTIONS)[number]

const BUSINESS_FEE_PERCENT = 10

/** Минимальные допустимые значения требований к ресурсам поставщика (обязательные поля). */
const MIN_REQ_CPU = 1
const MIN_REQ_RAM_GB = 0.5
const MIN_REQ_STORAGE_GB = 1
/** Минимальный объём тома (ГБ). */
const MIN_VOLUME_STORAGE_GB = 0.5

function toLocalDateKey(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

function getItemDateKey(r: AdminRentHistoryItem): string | null {
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

type UserOption = { id: string; name: string }

/** Уникальные пользователи из истории (покупатели и поставщики). */
function uniqueUsersFromHistory(history: AdminRentHistoryItem[]): UserOption[] {
  const byId = new Map<string, string>()
  for (const r of history) {
    if (r.client_id && !byId.has(r.client_id)) {
      byId.set(r.client_id, r.client_name || r.client_id)
    }
    if (r.merchant_id && !byId.has(r.merchant_id)) {
      byId.set(r.merchant_id, r.merchant_name || r.merchant_id)
    }
  }
  return Array.from(byId.entries(), ([id, name]) => ({ id, name })).sort(
    (a, b) => a.name.localeCompare(b.name) || a.id.localeCompare(b.id)
  )
}

export default function AdminView() {
  const [history, setHistory] = useState<AdminRentHistoryItem[]>([])
  const [loadingHistory, setLoadingHistory] = useState(true)
  const [historyError, setHistoryError] = useState<string | null>(null)
  const [historyRangeDays, setHistoryRangeDays] = useState<RangeDays>(30)
  const [historyWindowEnd, setHistoryWindowEnd] = useState<Date>(() => startOfDay(new Date()))
  const [selectedHistoryDay, setSelectedHistoryDay] = useState<string | null>(null)
  const [selectedUserIds, setSelectedUserIds] = useState<string[]>([])
  const [dropdownOpen, setDropdownOpen] = useState(false)
  const [dropdownSearch, setDropdownSearch] = useState('')
  const dropdownRef = useRef<HTMLDivElement | null>(null)
  const [historyTemplatePreview, setHistoryTemplatePreview] = useState<{
    item: AdminRentHistoryItem
    templateInfo: TemplateInfo | null
  } | null>(null)
  const [historyTemplatePreviewLoading, setHistoryTemplatePreviewLoading] = useState(false)
  const [systemTemplates, setSystemTemplates] = useState<TemplateInfo[]>([])
  const [systemTemplatesLoading, setSystemTemplatesLoading] = useState(false)
  const [templateEditModal, setTemplateEditModal] = useState<{
    open: boolean
    mode: 'edit' | 'create'
    templateId: string | null
  }>({ open: false, mode: 'create', templateId: null })
  const [formTitle, setFormTitle] = useState('')
  const [formDescription, setFormDescription] = useState('')
  const [formShortDescription, setFormShortDescription] = useState('')
  const [formContainerImageName, setFormContainerImageName] = useState('')
  const [formContainerImageTag, setFormContainerImageTag] = useState('')
  const [formUseGpu, setFormUseGpu] = useState(false)
  const [formPorts, setFormPorts] = useState<TemplatePort[]>([])
  const [formEnvs, setFormEnvs] = useState<TemplateEnv[]>([])
  const [formVolumes, setFormVolumes] = useState<string[]>([])
  const [formMinCpu, setFormMinCpu] = useState<number>(0)
  const [formMinRamGb, setFormMinRamGb] = useState<number>(0)
  const [formMinStorageGb, setFormMinStorageGb] = useState<number>(0)
  const [formMinVolumeStorageGb, setFormMinVolumeStorageGb] = useState<number[]>([])
  const [formSubmitError, setFormSubmitError] = useState<string | null>(null)
  const [formSubmitting, setFormSubmitting] = useState(false)
  const [templateEditLoading, setTemplateEditLoading] = useState(false)
  const templatesCarouselRef = useRef<HTMLDivElement | null>(null)
  const formImageFileInputRef = useRef<HTMLInputElement | null>(null)

  const loadHistory = useCallback(async () => {
    setLoadingHistory(true)
    setHistoryError(null)
    try {
      const data = await api.admin.getHistoryRents()
      setHistory(Array.isArray(data) ? data : [])
    } catch (e) {
      setHistory([])
      setHistoryError(e instanceof Error ? e.message : 'Не удалось загрузить историю')
    } finally {
      setLoadingHistory(false)
    }
  }, [])

  useEffect(() => {
    loadHistory()
  }, [loadHistory])

  const loadSystemTemplates = useCallback(async () => {
    setSystemTemplatesLoading(true)
    try {
      const data = await api.clientRent.getTemplates()
      setSystemTemplates(Array.isArray(data) ? data : [])
    } catch {
      setSystemTemplates([])
    } finally {
      setSystemTemplatesLoading(false)
    }
  }, [])

  useEffect(() => {
    loadSystemTemplates()
  }, [loadSystemTemplates])

  const scrollCarousel = useCallback((direction: 'left' | 'right') => {
    const el = templatesCarouselRef.current
    if (!el) return
    const step = 320
    el.scrollBy({ left: direction === 'left' ? -step : step, behavior: 'smooth' })
  }, [])

  const openEditTemplate = useCallback((templateId: string) => {
    setTemplateEditModal({ open: true, mode: 'edit', templateId })
  }, [])

  const openCreateTemplate = useCallback(() => {
    setTemplateEditModal({ open: true, mode: 'create', templateId: null })
  }, [])

  // Заполнить форму при открытии модалки: при редактировании — загрузить шаблон по ID (чтобы получить ports, envs, volumes)
  useEffect(() => {
    if (!templateEditModal.open) return
    setFormSubmitError(null)
    if (templateEditModal.mode !== 'edit' || !templateEditModal.templateId) {
      setFormTitle('')
      setFormDescription('')
      setFormShortDescription('')
      setFormContainerImageName('')
      setFormContainerImageTag('')
      setFormUseGpu(false)
      setFormPorts([])
      setFormEnvs([])
      setFormVolumes([])
      setFormMinCpu(MIN_REQ_CPU)
      setFormMinRamGb(MIN_REQ_RAM_GB)
      setFormMinStorageGb(MIN_REQ_STORAGE_GB)
      setFormMinVolumeStorageGb([])
      setTemplateEditLoading(false)
      return
    }
    let cancelled = false
    setTemplateEditLoading(true)
    api.admin
      .getTemplate(templateEditModal.templateId)
      .then((t) => {
        if (cancelled) return
        setFormTitle(t.title ?? '')
        setFormDescription(t.description ?? '')
        setFormShortDescription(t.short_description ?? '')
        setFormContainerImageName(t.container_image_name ?? '')
        setFormContainerImageTag(t.container_image_tag ?? '')
        setFormUseGpu(!!t.use_gpu)
        setFormPorts(Array.isArray(t.ports) ? t.ports.map((p) => ({ auth: p.auth, port: p.port, type: 'http', title: p.title })) : [])
        setFormEnvs(Array.isArray(t.envs) ? t.envs.map((e) => ({ key: e.key, value: e.value, type: e.type })) : [])
        setFormVolumes(Array.isArray(t.volumes) ? [...t.volumes] : [])
        setFormMinCpu(Math.max(MIN_REQ_CPU, t.min_cpu ?? 0))
        setFormMinRamGb(Math.max(MIN_REQ_RAM_GB, (t.min_ram_bytes ?? 0) / 1024 ** 3))
        setFormMinStorageGb(Math.max(MIN_REQ_STORAGE_GB, (t.min_storage_bytes ?? 0) / 1024 ** 3))
        const vols = Array.isArray(t.volumes) ? t.volumes : []
        const volGb = Array.isArray(t.min_volume_storage_bytes)
          ? t.min_volume_storage_bytes.map((b) => Math.max(MIN_VOLUME_STORAGE_GB, b / 1024 ** 3))
          : []
        const padded = [...volGb]
        while (padded.length < vols.length) padded.push(MIN_VOLUME_STORAGE_GB)
        setFormMinVolumeStorageGb(padded.slice(0, vols.length))
      })
      .catch(() => {
        if (!cancelled) setFormSubmitError('Не удалось загрузить шаблон')
      })
      .finally(() => {
        if (!cancelled) setTemplateEditLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [templateEditModal.open, templateEditModal.mode, templateEditModal.templateId])

  const handleFormImageFileChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file || !file.type.startsWith('image/')) return
    const reader = new FileReader()
    reader.onload = () => setFormShortDescription(String(reader.result))
    reader.readAsDataURL(file)
    e.target.value = ''
  }, [])

  const handleTemplateFormSubmit = useCallback(
    async (e: React.FormEvent) => {
      e.preventDefault()
      setFormSubmitError(null)

      const cpuOk = formMinCpu >= MIN_REQ_CPU
      const ramOk = formMinRamGb >= MIN_REQ_RAM_GB
      const storageOk = formMinStorageGb >= MIN_REQ_STORAGE_GB
      if (!cpuOk || !ramOk || !storageOk) {
        setFormSubmitError(
          `Укажите требования к ресурсам: CPU не менее ${MIN_REQ_CPU} ядра, RAM не менее ${MIN_REQ_RAM_GB} ГБ, Storage не менее ${MIN_REQ_STORAGE_GB} ГБ.`
        )
        return
      }

      setFormSubmitting(true)
      try {
        const portsFiltered = formPorts.filter((p) => p.port != null)
        const envsFiltered = formEnvs.filter((e) => (e.key ?? '').trim() !== '')
        const volumesFiltered = formVolumes.filter((v) => v.trim() !== '')
        const minVolumeBytes =
          formMinVolumeStorageGb.length > 0
            ? formMinVolumeStorageGb.slice(0, volumesFiltered.length).map((gb) => Math.round(gb * 1024 ** 3))
            : undefined
        const payload = {
          title: formTitle.trim() || undefined,
          description: formDescription.trim() || undefined,
          short_description: formShortDescription.trim() || undefined,
          container_image_name: formContainerImageName.trim() || undefined,
          container_image_tag: formContainerImageTag.trim() || undefined,
          use_gpu: formUseGpu,
          ports: portsFiltered.length ? portsFiltered.map((p) => ({ ...p, type: 'http' })) : undefined,
          envs: envsFiltered.length ? envsFiltered : undefined,
          volumes: volumesFiltered.length ? volumesFiltered : undefined,
          min_cpu: formMinCpu,
          min_ram_bytes: Math.round(formMinRamGb * 1024 ** 3),
          min_storage_bytes: Math.round(formMinStorageGb * 1024 ** 3),
          min_volume_storage_bytes: minVolumeBytes?.some((b) => b > 0) ? minVolumeBytes : undefined,
        }
        if (templateEditModal.mode === 'edit' && templateEditModal.templateId) {
          await api.admin.updateTemplate(templateEditModal.templateId, payload)
        } else {
          await api.admin.createTemplate(payload)
        }
        setTemplateEditModal((prev) => ({ ...prev, open: false }))
        loadSystemTemplates()
      } catch (err) {
        setFormSubmitError(err instanceof Error ? err.message : 'Ошибка сохранения')
      } finally {
        setFormSubmitting(false)
      }
    },
    [
      formTitle,
      formDescription,
      formShortDescription,
      formContainerImageName,
      formContainerImageTag,
      formUseGpu,
      formPorts,
      formEnvs,
      formVolumes,
      formMinCpu,
      formMinRamGb,
      formMinStorageGb,
      formMinVolumeStorageGb,
      templateEditModal.mode,
      templateEditModal.templateId,
      loadSystemTemplates,
    ]
  )

  const uniqueUsers = useMemo(() => uniqueUsersFromHistory(history), [history])

  const filteredHistory = useMemo(() => {
    if (selectedUserIds.length === 0) return history
    const set = new Set(selectedUserIds)
    return history.filter(
      (r) => set.has(String(r.client_id)) || set.has(String(r.merchant_id))
    )
  }, [history, selectedUserIds])

  const dropdownOptions = useMemo(() => {
    const search = dropdownSearch.trim().toLowerCase()
    return uniqueUsers.filter((u) => {
      if (selectedUserIds.includes(u.id)) return false
      if (!search) return true
      return (
        u.id.toLowerCase().includes(search) ||
        u.name.toLowerCase().includes(search)
      )
    })
  }, [uniqueUsers, selectedUserIds, dropdownSearch])

  const addToFilter = useCallback((id: string) => {
    setSelectedUserIds((prev) => (prev.includes(id) ? prev : [...prev, id]))
    setDropdownSearch('')
  }, [])

  const removeFromFilter = useCallback((id: string) => {
    setSelectedUserIds((prev) => prev.filter((x) => x !== id))
  }, [])

  const openHistoryTemplatePreview = useCallback((item: AdminRentHistoryItem) => {
    const tid = item.template_id
    setHistoryTemplatePreview({ item, templateInfo: null })
    setHistoryTemplatePreviewLoading(true)
    api.clientRent
      .getTemplates()
      .then((list) => {
        const t = Array.isArray(list)
          ? list.find((x) => x.template_id === tid) ?? null
          : null
        setHistoryTemplatePreview((prev) =>
          prev ? { ...prev, templateInfo: t } : null
        )
      })
      .catch(() =>
        setHistoryTemplatePreview((prev) =>
          prev ? { ...prev, templateInfo: null } : null
        )
      )
      .finally(() => setHistoryTemplatePreviewLoading(false))
  }, [])

  useEffect(() => {
    if (!dropdownOpen) return
    const handleClickOutside = (e: MouseEvent) => {
      const el = e.target as Node
      if (dropdownRef.current && !dropdownRef.current.contains(el))
        setDropdownOpen(false)
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [dropdownOpen])

  const today = useMemo(() => startOfDay(new Date()), [])

  const historyByDay = useMemo(() => {
    const map = new Map<
      string,
      { items: AdminRentHistoryItem[]; totalCost: number; totalBusinessRevenue: number }
    >()
    for (const r of filteredHistory) {
      const key = getItemDateKey(r)
      if (!key) continue
      const businessRevenue = (r.cost * BUSINESS_FEE_PERCENT) / 100
      const existing = map.get(key)
      if (existing) {
        existing.items.push(r)
        existing.totalCost += r.cost
        existing.totalBusinessRevenue += businessRevenue
      } else {
        map.set(key, {
          items: [r],
          totalCost: r.cost,
          totalBusinessRevenue: businessRevenue,
        })
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
  }, [filteredHistory])

  const historyWindowStart = useMemo(
    () => addDays(historyWindowEnd, -(historyRangeDays - 1)),
    [historyWindowEnd, historyRangeDays]
  )

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

  const userById = useMemo(() => new Map(uniqueUsers.map((u) => [u.id, u])), [uniqueUsers])

  return (
    <div className={styles.wrap}>
      <p className={styles.pageIntro}>
        Здесь отображается сводка по всем арендам: покупатели, поставщики, суммы и доход платформы. В разделе шаблонов можно управлять программами, которые доступны покупателям при оформлении аренды.
      </p>

      <section className={sharedStyles.section} aria-labelledby="admin-templates-heading">
        <div className={styles.templatesSectionHeader}>
          <h3 id="admin-templates-heading">Шаблоны в системе</h3>
          <button
            type="button"
            className={sharedStyles.primaryBtn}
            onClick={openCreateTemplate}
          >
            Добавить новый
          </button>
        </div>
        {systemTemplatesLoading && (
          <p className={sharedStyles.muted}>Загрузка шаблонов…</p>
        )}
        {!systemTemplatesLoading && (
          <div className={styles.templatesCarouselWrap}>
            <button
              type="button"
              className={styles.carouselBtn}
              onClick={() => scrollCarousel('left')}
              aria-label="Листать влево"
            >
              ←
            </button>
            <div
              className={styles.templatesCarousel}
              ref={templatesCarouselRef}
              role="list"
            >
              {systemTemplates.map((t) => (
                <div key={t.template_id} className={styles.templateSlideCard} role="listitem">
                  {t.short_description ? (
                    <div className={styles.templateSlideImage}>
                      <img
                        src={t.short_description}
                        alt={t.title || t.template_id}
                      />
                    </div>
                  ) : (
                    <div className={styles.templateSlidePlaceholder}>
                      {t.title || t.template_id}
                    </div>
                  )}
                  <div className={styles.templateSlideContent}>
                    <span className={styles.templateSlideTitle}>
                      {t.title || t.template_id}
                    </span>
                    <button
                      type="button"
                      className={sharedStyles.secondaryBtn}
                      onClick={() => openEditTemplate(t.template_id)}
                    >
                      Редактировать
                    </button>
                  </div>
                </div>
              ))}
            </div>
            <button
              type="button"
              className={styles.carouselBtn}
              onClick={() => scrollCarousel('right')}
              aria-label="Листать вправо"
            >
              →
            </button>
          </div>
        )}
      </section>

      <section className={sharedStyles.section} aria-labelledby="admin-history-heading">
        <h3 id="admin-history-heading">Аналитика аренд</h3>

        <div className={styles.filterBlock}>
          <span className={styles.filterLabel}>Фильтр по пользователям:</span>
          <div className={styles.chipsRow}>
            {selectedUserIds.map((id) => {
              const u = userById.get(id)
              return (
                <span key={id} className={styles.chip}>
                  {u ? `${u.name} (id: ${u.id})` : `id: ${id}`}
                  <button
                    type="button"
                    className={styles.chipRemove}
                    onClick={() => removeFromFilter(id)}
                    aria-label={`Убрать из фильтра`}
                  >
                    ×
                  </button>
                </span>
              )
            })}
          </div>
          <div className={styles.dropdownWrap} ref={dropdownRef}>
            <input
              type="text"
              value={dropdownSearch}
              onChange={(e) => setDropdownSearch(e.target.value)}
              onFocus={() => setDropdownOpen(true)}
              onBlur={() => {
                // Закрыть при клике мимо (потеря фокуса); задержка чтобы клик по пункту списка успел сработать
                setTimeout(() => setDropdownOpen(false), 150)
              }}
              onKeyDown={(e) => {
                if (e.key === 'Escape') setDropdownOpen(false)
              }}
              placeholder="Найти и добавить пользователя…"
              className={styles.dropdownInput}
              aria-expanded={dropdownOpen}
              aria-haspopup="listbox"
              aria-controls="admin-user-listbox"
              id="admin-user-filter-input"
            />
            {dropdownOpen && (
              <ul
                id="admin-user-listbox"
                className={styles.dropdownList}
                role="listbox"
                aria-labelledby="admin-user-filter-input"
              >
                {dropdownOptions.length === 0 ? (
                  <li className={styles.dropdownEmpty}>
                    {dropdownSearch.trim()
                      ? 'Никого не найдено'
                      : 'Все пользователи уже добавлены'}
                  </li>
                ) : (
                  dropdownOptions.map((u) => (
                    <li
                      key={u.id}
                      role="option"
                      className={styles.dropdownItem}
                      onClick={() => addToFilter(u.id)}
                      onKeyDown={(e) => {
                        if (e.key === 'Enter' || e.key === ' ') {
                          e.preventDefault()
                          addToFilter(u.id)
                        }
                      }}
                      tabIndex={0}
                    >
                      {u.name} (id: {u.id})
                    </li>
                  ))
                )}
              </ul>
            )}
          </div>
        </div>

        {loadingHistory && <p className={sharedStyles.muted}>Загрузка…</p>}
        {historyError && (
          <div className={sharedStyles.errorBlock}>{historyError}</div>
        )}
        {!loadingHistory && filteredHistory.length === 0 && (
          <p className={sharedStyles.muted}>
            {history.length === 0
              ? 'Нет записей в истории.'
              : 'Нет записей по выбранному фильтру.'}
          </p>
        )}
        {!loadingHistory && filteredHistory.length > 0 && (
          <div className={sharedStyles.historyChartWrap}>
            <div className={sharedStyles.historyChartControls}>
              <span className={sharedStyles.historyChartLabel}>Период:</span>
              <div className={sharedStyles.historyRangeButtons}>
                {RANGE_OPTIONS.map((n) => (
                  <button
                    key={n}
                    type="button"
                    className={
                      historyRangeDays === n
                        ? sharedStyles.primaryBtn
                        : sharedStyles.secondaryBtn
                    }
                    onClick={() => {
                      setHistoryRangeDays(n)
                      setSelectedHistoryDay(null)
                    }}
                  >
                    {n} дн.
                  </button>
                ))}
              </div>
              <div className={sharedStyles.historyNavButtons}>
                <button
                  type="button"
                  className={sharedStyles.secondaryBtn}
                  onClick={() => {
                    setHistoryWindowEnd((prev) =>
                      addDays(prev, -historyRangeDays)
                    )
                    setSelectedHistoryDay(null)
                  }}
                  title="Более ранний период"
                >
                  ← Ранее
                </button>
                <button
                  type="button"
                  className={sharedStyles.secondaryBtn}
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
            <div className={sharedStyles.historyChartContainer}>
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
                    label={{
                      value: 'Сеансов',
                      angle: -90,
                      position: 'insideLeft',
                      fill: 'var(--muted)',
                      fontSize: 11,
                    }}
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
                    formatter={(value: number) => [
                      `${value} сеанс(ов)`,
                      'Количество',
                    ]}
                    labelFormatter={(
                      _label: string,
                      payload: { payload?: { dateKey?: string } }[]
                    ) =>
                      payload[0]?.payload?.dateKey
                        ? new Date(
                            payload[0].payload.dateKey
                          ).toLocaleDateString('ru-RU', {
                            day: '2-digit',
                            month: 'long',
                            year: 'numeric',
                          })
                        : _label
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
                      if (key)
                        setSelectedHistoryDay((prev) =>
                          prev === key ? null : key
                        )
                    }}
                  >
                    {historyChartData.map((entry) => (
                      <Cell
                        key={entry.dateKey}
                        fill="var(--accent)"
                        opacity={
                          selectedHistoryDay === entry.dateKey ? 1 : 0.7
                        }
                        stroke={
                          selectedHistoryDay === entry.dateKey
                            ? 'var(--text)'
                            : 'transparent'
                        }
                        strokeWidth={2}
                      />
                    ))}
                  </Bar>
                </BarChart>
              </ResponsiveContainer>
            </div>
            {selectedHistoryDay && selectedDayInfo && (
              <div className={sharedStyles.historyDayDetail}>
                <h4 className={sharedStyles.historyDayDetailTitle}>
                  {new Date(
                    selectedHistoryDay
                  ).toLocaleDateString('ru-RU', {
                    weekday: 'long',
                    day: 'numeric',
                    month: 'long',
                    year: 'numeric',
                  })}
                </h4>
                <div className={sharedStyles.historyDayDetailStats}>
                  <span title="Сумма за день (потрачено покупателями)">
                    Потрачено всего:{' '}
                    <strong>
                      {selectedDayInfo.totalCost.toFixed(2)} BYN
                    </strong>
                  </span>
                  <span title="Доход бизнеса 10%">
                    Доход бизнеса (10%):{' '}
                    <strong>
                      {selectedDayInfo.totalBusinessRevenue.toFixed(2)} BYN
                    </strong>
                  </span>
                  <span title="Количество сеансов">
                    Сеансов: <strong>{selectedDayInfo.items.length}</strong>
                  </span>
                </div>
                <h5 className={sharedStyles.historyDayDetailListTitle}>
                  Сеансы за этот день
                </h5>
                <div className={styles.adminTableWrap}>
                  <div className={styles.adminTableHeader}>
                    <span>Время</span>
                    <span>Покупатель</span>
                    <span>Поставщик</span>
                    <span>Потрачено (BYN)</span>
                    <span>Доход 10% (BYN)</span>
                    <span>Шаблон</span>
                    <span>Длит.</span>
                  </div>
                  <ul className={styles.adminTableList} aria-label="Сеансы за день">
                    {selectedDayInfo.items.map((r) => {
                      const businessRevenue =
                        (r.cost * BUSINESS_FEE_PERCENT) / 100
                      return (
                        <li key={r.id} className={styles.adminTableRow}>
                          <span className={sharedStyles.historyDayDetailTime}>
                            {r.started_at
                              ? new Date(
                                  r.started_at
                                ).toLocaleTimeString('ru-RU', {
                                  hour: '2-digit',
                                  minute: '2-digit',
                                })
                              : '—'}
                          </span>
                          <span title={`id: ${r.client_id}`}>
                            {r.client_name || '—'} (id: {r.client_id})
                          </span>
                          <span title={`id: ${r.merchant_id}`}>
                            {r.merchant_name || '—'} (id: {r.merchant_id})
                          </span>
                          <span
                            className={sharedStyles.historyDayDetailCost}
                          >
                            {r.cost.toFixed(2)}
                          </span>
                          <span className={styles.businessRevenue}>
                            {businessRevenue.toFixed(2)}
                          </span>
                          <span>
                            <button
                              type="button"
                              className={sharedStyles.templateNameLink}
                              onClick={() => openHistoryTemplatePreview(r)}
                            >
                              {r.template_title || r.template_id || '—'}
                            </button>
                          </span>
                          <span className={sharedStyles.muted}>
                            {r.duration} мин
                          </span>
                        </li>
                      )
                    })}
                  </ul>
                </div>
              </div>
            )}
          </div>
        )}
      </section>

      {templateEditModal.open && (
        <div
          className={sharedStyles.modalBackdrop}
          onClick={() =>
            setTemplateEditModal((prev) => ({ ...prev, open: false }))
          }
        >
          <div
            className={sharedStyles.modal}
            onClick={(e) => e.stopPropagation()}
          >
            <div className={sharedStyles.modalHeader}>
              <h3 className={sharedStyles.modalTitle}>
                {templateEditModal.mode === 'edit'
                  ? 'Редактирование шаблона'
                  : 'Новый шаблон'}
              </h3>
              <button
                type="button"
                className={sharedStyles.modalClose}
                onClick={() =>
                  setTemplateEditModal((prev) => ({ ...prev, open: false }))
                }
                aria-label="Закрыть"
              >
                ×
              </button>
            </div>
            <div className={sharedStyles.modalBody}>
              {templateEditLoading ? (
                <p className={sharedStyles.muted}>Загрузка…</p>
              ) : (
              <form
                className={styles.templateForm}
                onSubmit={handleTemplateFormSubmit}
                noValidate
              >
                {templateEditModal.mode === 'edit' && templateEditModal.templateId && (
                  <div className={styles.formSection}>
                    <div className={styles.formRow}>
                      <label className={styles.formLabel}>ID шаблона</label>
                      <span className={styles.formId}>{templateEditModal.templateId}</span>
                    </div>
                  </div>
                )}

                <div className={styles.formSection}>
                  <h4 className={styles.formSectionTitle}>Основное</h4>
                  <div className={styles.formRow}>
                    <label htmlFor="template-form-title" className={styles.formLabel}>
                      Название
                    </label>
                    <input
                      id="template-form-title"
                      type="text"
                      className={styles.formInput}
                      value={formTitle}
                      onChange={(e) => setFormTitle(e.target.value)}
                      placeholder="Название шаблона"
                    />
                  </div>
                  <div className={styles.formRow}>
                    <label htmlFor="template-form-description" className={styles.formLabel}>
                      Описание
                    </label>
                    <textarea
                      id="template-form-description"
                      className={styles.formTextarea}
                      value={formDescription}
                      onChange={(e) => setFormDescription(e.target.value)}
                      placeholder="Текстовое описание"
                      rows={3}
                    />
                  </div>
                </div>

                <div className={styles.formSection}>
                  <h4 className={styles.formSectionTitle}>Docker-образ</h4>
                  <div className={styles.formGrid}>
                    <div className={styles.formRow}>
                      <label htmlFor="template-form-image-name" className={styles.formLabel}>
                        Имя образа
                      </label>
                      <input
                        id="template-form-image-name"
                        type="text"
                        className={styles.formInput}
                        value={formContainerImageName}
                        onChange={(e) => setFormContainerImageName(e.target.value)}
                        placeholder="registry.io/my-image"
                      />
                    </div>
                    <div className={styles.formRow}>
                      <label htmlFor="template-form-image-tag" className={styles.formLabel}>
                        Тег образа
                      </label>
                      <input
                        id="template-form-image-tag"
                        type="text"
                        className={styles.formInput}
                        value={formContainerImageTag}
                        onChange={(e) => setFormContainerImageTag(e.target.value)}
                        placeholder="latest"
                      />
                    </div>
                  </div>
                </div>

                <div className={styles.formSection}>
                  <h4 className={styles.formSectionTitle}>Технические настройки</h4>
                  <div className={styles.formListBlock}>
                    <span className={styles.formLabel}>Порты</span>
                    {formPorts.map((p, i) => (
                      <div key={i} className={styles.formListRow}>
                        <input
                          type="number"
                          className={styles.formInputSmall}
                          placeholder="Порт"
                          value={p.port ?? ''}
                          onChange={(e) => {
                            const v = e.target.value ? Number(e.target.value) : undefined
                            setFormPorts((prev) => prev.map((x, j) => (j === i ? { ...x, port: v } : x)))
                          }}
                        />
                        <span className={styles.formPortTypeFixed} title="Тип порта всегда http">http</span>
                        <input
                          type="text"
                          className={styles.formInputFlex}
                          placeholder="Название"
                          value={p.title ?? ''}
                          onChange={(e) => setFormPorts((prev) => prev.map((x, j) => (j === i ? { ...x, title: e.target.value } : x)))}
                        />
                        <label className={styles.formCheckboxInline}>
                          <input
                            type="checkbox"
                            checked={!!p.auth}
                            onChange={(e) => setFormPorts((prev) => prev.map((x, j) => (j === i ? { ...x, auth: e.target.checked } : x)))}
                          />
                          auth
                        </label>
                        <button type="button" className={styles.formListRemove} onClick={() => setFormPorts((prev) => prev.filter((_, j) => j !== i))} aria-label="Удалить порт">
                          ×
                        </button>
                      </div>
                    ))}
                    <button type="button" className={styles.formListAdd} onClick={() => setFormPorts((prev) => [...prev, { port: undefined, type: 'http', title: '', auth: false }])}>
                      + Добавить порт
                    </button>
                  </div>
                  <div className={styles.formListBlock}>
                    <span className={styles.formLabel}>Переменные окружения</span>
                    {formEnvs.map((e, i) => (
                      <div key={i} className={styles.formListRow}>
                        <input
                          type="text"
                          className={styles.formInputFlex}
                          placeholder="Ключ"
                          value={e.key ?? ''}
                          onChange={(ev) => setFormEnvs((prev) => prev.map((x, j) => (j === i ? { ...x, key: ev.target.value } : x)))}
                        />
                        <input
                          type="text"
                          className={styles.formInputFlex}
                          placeholder="Значение"
                          value={e.value ?? ''}
                          onChange={(ev) => setFormEnvs((prev) => prev.map((x, j) => (j === i ? { ...x, value: ev.target.value } : x)))}
                        />
                        <button type="button" className={styles.formListRemove} onClick={() => setFormEnvs((prev) => prev.filter((_, j) => j !== i))} aria-label="Удалить переменную">
                          ×
                        </button>
                      </div>
                    ))}
                    <button type="button" className={styles.formListAdd} onClick={() => setFormEnvs((prev) => [...prev, { key: '', value: '' }])}>
                      + Добавить переменную
                    </button>
                  </div>
                  <div className={styles.formListBlock}>
                    <span className={styles.formLabel}>Тома (пути и минимальный объём)</span>
                    {formVolumes.length > 0 && (
                      <div className={styles.formVolumesHeader}>
                        <span className={styles.formVolumesHeaderPath}>Путь в контейнере</span>
                        <span className={styles.formVolumesHeaderSize}>Мин. объём (ГБ)</span>
                        <span className={styles.formVolumesHeaderAction} aria-hidden> </span>
                      </div>
                    )}
                    {formVolumes.map((v, i) => (
                      <div key={i} className={styles.formListRow}>
                        <input
                          type="text"
                          className={styles.formInputFlex}
                          placeholder="/path/inside/container"
                          value={v}
                          onChange={(e) => setFormVolumes((prev) => prev.map((x, j) => (j === i ? e.target.value : x)))}
                          aria-label={`Путь тома ${i + 1}`}
                        />
                        <label className={styles.formVolumeSizeLabel}>
                          <input
                            type="number"
                            min={MIN_VOLUME_STORAGE_GB}
                            step={0.5}
                            className={styles.formInputNum}
                            placeholder={String(MIN_VOLUME_STORAGE_GB)}
                            value={formMinVolumeStorageGb[i] ?? ''}
                            onChange={(e) => {
                              const val = parseFloat(e.target.value)
                              setFormMinVolumeStorageGb((prev) => {
                                const next = [...(prev ?? [])]
                                while (next.length <= i) next.push(MIN_VOLUME_STORAGE_GB)
                                next[i] = Number.isFinite(val) ? Math.max(MIN_VOLUME_STORAGE_GB, val) : MIN_VOLUME_STORAGE_GB
                                return next
                              })
                            }}
                          />
                          <span className={styles.formVolumeSizeHint}>ГБ</span>
                        </label>
                        <button type="button" className={styles.formListRemove} onClick={() => { setFormVolumes((prev) => prev.filter((_, j) => j !== i)); setFormMinVolumeStorageGb((prev) => prev.filter((_, j) => j !== i)) }} aria-label="Удалить том">
                          ×
                        </button>
                      </div>
                    ))}
                    <button type="button" className={styles.formListAdd} onClick={() => { setFormVolumes((prev) => [...prev, '']); setFormMinVolumeStorageGb((prev) => [...prev, MIN_VOLUME_STORAGE_GB]) }}>
                      + Добавить том
                    </button>
                  </div>

                  <div className={styles.formSection}>
                    <h4 className={styles.formSectionTitle}>Минимальные требования шаблона</h4>
                    <div className={styles.formMinReqsGrid}>
                      <label className={styles.formMinReqItem}>
                        <span className={styles.formMinReqLabel}>CPU (ядер)</span>
                        <input
                          type="number"
                          min={MIN_REQ_CPU}
                          step={1}
                          required
                          value={formMinCpu || ''}
                          onChange={(e) => setFormMinCpu(Math.max(MIN_REQ_CPU, parseInt(e.target.value, 10) || MIN_REQ_CPU))}
                          placeholder={String(MIN_REQ_CPU)}
                        />
                      </label>
                      <label className={styles.formMinReqItem}>
                        <span className={styles.formMinReqLabel}>RAM (ГБ)</span>
                        <input
                          type="number"
                          min={MIN_REQ_RAM_GB}
                          step={0.5}
                          required
                          value={formMinRamGb || ''}
                          onChange={(e) => setFormMinRamGb(Math.max(MIN_REQ_RAM_GB, parseFloat(e.target.value) || MIN_REQ_RAM_GB))}
                          placeholder={String(MIN_REQ_RAM_GB)}
                        />
                      </label>
                      <label className={styles.formMinReqItem}>
                        <span className={styles.formMinReqLabel}>Storage приложения (ГБ)</span>
                        <input
                          type="number"
                          min={MIN_REQ_STORAGE_GB}
                          step={0.5}
                          required
                          value={formMinStorageGb || ''}
                          onChange={(e) => setFormMinStorageGb(Math.max(MIN_REQ_STORAGE_GB, parseFloat(e.target.value) || MIN_REQ_STORAGE_GB))}
                          placeholder={String(MIN_REQ_STORAGE_GB)}
                        />
                      </label>
                    </div>
                    <div className={styles.formSwitchRow}>
                      <span className={styles.formLabel}>Использует GPU</span>
                      <button
                        type="button"
                        role="switch"
                        aria-checked={formUseGpu}
                        className={styles.formSwitch + (formUseGpu ? ' ' + styles.formSwitchOn : '')}
                        onClick={() => setFormUseGpu(!formUseGpu)}
                      >
                        <span className={styles.formSwitchThumb} />
                      </button>
                    </div>
                  </div>
                </div>

                <div className={styles.formSection}>
                  <h4 className={styles.formSectionTitle}>Изображение (превью)</h4>
                  <div className={styles.formImageBlock}>
                    <div className={styles.formImageActions}>
                      <input
                        ref={formImageFileInputRef}
                        type="file"
                        accept="image/*"
                        className={styles.formFileInput}
                        onChange={handleFormImageFileChange}
                        aria-label="Загрузить изображение"
                      />
                      <button
                        type="button"
                        className={sharedStyles.secondaryBtn}
                        onClick={() => formImageFileInputRef.current?.click()}
                      >
                        Выбрать файл
                      </button>
                      {formShortDescription && (
                        <button
                          type="button"
                          className={sharedStyles.secondaryBtn}
                          onClick={() => setFormShortDescription('')}
                        >
                          Удалить изображение
                        </button>
                      )}
                    </div>
                    {formShortDescription && (
                      <div className={styles.formImagePreview}>
                        <img
                          src={formShortDescription}
                          alt=""
                          onError={(e) => {
                            e.currentTarget.style.display = 'none'
                          }}
                        />
                      </div>
                    )}
                  </div>
                </div>

                {formSubmitError && (
                  <div className={sharedStyles.errorBlock}>{formSubmitError}</div>
                )}
                <div className={styles.formActions}>
                  <button
                    type="button"
                    className={sharedStyles.secondaryBtn}
                    onClick={() =>
                      setTemplateEditModal((prev) => ({ ...prev, open: false }))
                    }
                    disabled={formSubmitting}
                  >
                    Отмена
                  </button>
                  <button
                    type="submit"
                    className={sharedStyles.primaryBtn}
                    disabled={formSubmitting}
                  >
                    {formSubmitting ? 'Сохранение…' : 'Сохранить'}
                  </button>
                </div>
              </form>
              )}
            </div>
          </div>
        </div>
      )}

      {historyTemplatePreview && (
        <div
          className={sharedStyles.modalBackdrop}
          onClick={() => setHistoryTemplatePreview(null)}
        >
          <div
            className={sharedStyles.modal}
            onClick={(e) => e.stopPropagation()}
          >
            <div className={sharedStyles.modalHeader}>
              <h3 className={sharedStyles.modalTitle}>Шаблон</h3>
              <button
                type="button"
                className={sharedStyles.modalClose}
                onClick={() => setHistoryTemplatePreview(null)}
                aria-label="Закрыть"
              >
                ×
              </button>
            </div>
            <div className={sharedStyles.modalBody}>
              {historyTemplatePreviewLoading && (
                <p className={sharedStyles.muted}>Загрузка…</p>
              )}
              {!historyTemplatePreviewLoading &&
                historyTemplatePreview.templateInfo && (
                  <div className={sharedStyles.templatePreviewCard}>
                    {historyTemplatePreview.templateInfo.short_description && (
                      <div className={sharedStyles.templateCardImage}>
                        <img
                          src={
                            historyTemplatePreview.templateInfo.short_description
                          }
                          alt={
                            historyTemplatePreview.templateInfo.title ||
                            historyTemplatePreview.templateInfo.template_id
                          }
                        />
                      </div>
                    )}
                    <div className={sharedStyles.templateCardContent}>
                      <h4 className={sharedStyles.templateCardTitle}>
                        {historyTemplatePreview.templateInfo.title ||
                          historyTemplatePreview.templateInfo.template_id}
                      </h4>
                      {historyTemplatePreview.templateInfo.description && (
                        <p className={sharedStyles.templateCardDescription}>
                          {
                            historyTemplatePreview.templateInfo.description
                          }
                        </p>
                      )}
                    </div>
                  </div>
                )}
              {!historyTemplatePreviewLoading &&
                !historyTemplatePreview.templateInfo && (
                  <p className={sharedStyles.muted}>
                    {historyTemplatePreview.item.template_title ||
                      historyTemplatePreview.item.template_id ||
                      '—'}. Описание шаблона недоступно.
                  </p>
                )}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
