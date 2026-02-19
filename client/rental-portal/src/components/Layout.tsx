import { useState, useCallback } from 'react'
import { Outlet, NavLink, useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import styles from './Layout.module.css'

function formatBalance(value: number): string {
  return (
    new Intl.NumberFormat('ru-RU', {
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    }).format(value) + ' BYN'
  )
}

/** Маска номера карты: только цифры, формат 0000 0000 0000 0000 */
function formatCardNumber(raw: string): string {
  const digits = raw.replace(/\D/g, '').slice(0, 16)
  const parts: string[] = []
  for (let i = 0; i < digits.length; i += 4) {
    parts.push(digits.slice(i, i + 4))
  }
  return parts.join(' ')
}

/** Парсинг суммы: число или 0, округление до сотых */
function parseAmount(s: string): number {
  const normalized = s.replace(/\s/g, '').replace(',', '.')
  const n = parseFloat(normalized)
  if (!Number.isFinite(n) || n < 0) return 0
  return Math.round(n * 100) / 100
}

/** Ограничить ввод суммы: только цифры, запятая/точка, максимум 2 знака после запятой */
function formatAmountInput(raw: string): string {
  const normalized = raw.replace(',', '.')
  const parts = normalized.split('.')
  if (parts.length > 2) {
    const intPart = parts[0].replace(/\D/g, '')
    const decimals = parts.slice(1).join('').replace(/\D/g, '').slice(0, 2)
    return intPart + (decimals ? '.' + decimals : '')
  }
  if (parts.length === 2) {
    const intPart = parts[0].replace(/\D/g, '')
    const decimals = parts[1].replace(/\D/g, '').slice(0, 2)
    const suffix = decimals.length ? '.' + decimals : (parts[1] === '' ? '.' : '')
    return intPart + suffix
  }
  return normalized.replace(/\D/g, '')
}

/** Маска срока карты: MM/YY, только цифры */
function formatExpiryInput(raw: string): string {
  const digits = raw.replace(/\D/g, '').slice(0, 4)
  if (digits.length <= 2) return digits
  return digits.slice(0, 2) + '/' + digits.slice(2)
}

type WalletTab = 'deposit' | 'withdraw'

export default function Layout() {
  const { user, logout, hasSystemSettingsWrite, balance, refreshBalance, topUp, withdraw } = useAuth()
  const navigate = useNavigate()

  const [walletOpen, setWalletOpen] = useState(false)
  const [walletTab, setWalletTab] = useState<WalletTab>('deposit')

  const [card, setCard] = useState('')
  const [expiry, setExpiry] = useState('')
  const [cvv, setCvv] = useState('')
  const [amount, setAmount] = useState('')
  const [cardName, setCardName] = useState('')
  const [formError, setFormError] = useState<string | null>(null)
  const [formSubmitting, setFormSubmitting] = useState(false)

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  const openWallet = useCallback(() => {
    setWalletTab('deposit')
    setCard('')
    setExpiry('')
    setCvv('')
    setAmount('')
    setCardName('')
    setFormError(null)
    setFormSubmitting(false)
    setWalletOpen(true)
  }, [])

  const closeWallet = useCallback(() => {
    setWalletOpen(false)
    refreshBalance()
  }, [refreshBalance])

  const handleDepositSubmit = useCallback(
    async (e: React.FormEvent) => {
      e.preventDefault()
      setFormError(null)
      if (!isExpiryValid(expiry)) {
        setFormError('Укажите срок действия карты в формате ММ/ГГ (месяц 01–12)')
        return
      }
      setFormSubmitting(true)
      const numAmount = parseAmount(amount)
      if (numAmount <= 0) {
        setFormError('Введите сумму для пополнения')
        setFormSubmitting(false)
        return
      }
      try {
        await topUp(numAmount)
        closeWallet()
      } catch {
        setFormError('Не удалось пополнить счёт')
      } finally {
        setFormSubmitting(false)
      }
    },
    [amount, expiry, topUp, closeWallet]
  )

  const handleWithdrawSubmit = useCallback(
    async (e: React.FormEvent) => {
      e.preventDefault()
      setFormError(null)
      if (!isExpiryValid(expiry)) {
        setFormError('Укажите срок действия карты в формате ММ/ГГ (месяц 01–12)')
        return
      }
      setFormSubmitting(true)
      const numAmount = parseAmount(amount)
      if (numAmount <= 0) {
        setFormError('Введите сумму для вывода')
        setFormSubmitting(false)
        return
      }
      if (numAmount > balance) {
        setFormError('Недостаточно средств на балансе')
        setFormSubmitting(false)
        return
      }
      try {
        const ok = await withdraw(numAmount)
        if (ok) {
          closeWallet()
        } else {
          setFormError('Недостаточно средств на балансе')
        }
      } catch {
        setFormError('Не удалось вывести средства')
      } finally {
        setFormSubmitting(false)
      }
    },
    [amount, balance, expiry, withdraw, closeWallet]
  )

  const handleCardChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setCard(formatCardNumber(e.target.value))
  }

  const handleAmountChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setAmount(formatAmountInput(e.target.value))
  }

  const handleExpiryChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setExpiry(formatExpiryInput(e.target.value))
  }

  const handleCvvChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setCvv(e.target.value.replace(/\D/g, '').slice(0, 3))
  }

  /** Проверка срока: MM/YY, месяц 01–12 */
  function isExpiryValid(exp: string): boolean {
    const digits = exp.replace(/\D/g, '')
    if (digits.length !== 4) return false
    const mm = parseInt(digits.slice(0, 2), 10)
    return mm >= 1 && mm <= 12
  }

  return (
    <div className={styles.layout}>
      <header className={styles.header}>
        <div className={styles.headerInner}>
          <nav className={styles.nav}>
            <NavLink to="/" end className={({ isActive }) => (isActive ? styles.active : '')} title="Главная страница: о сервисе">
              Главная
            </NavLink>
            <NavLink to="/client" className={({ isActive }) => (isActive ? styles.active : '')} title="Раздел для покупателя: аренда ресурсов у поставщиков">
              Покупатель
            </NavLink>
            <NavLink to="/merchant" className={({ isActive }) => (isActive ? styles.active : '')} title="Раздел для поставщика ресурсов">
              Поставщик
            </NavLink>
            {hasSystemSettingsWrite && (
              <NavLink to="/admin" className={({ isActive }) => (isActive ? styles.active : '')} title="Панель администратора">
                Панель администратора
              </NavLink>
            )}
          </nav>
          <div className={styles.userRow}>
            <button type="button" className={styles.balanceTrigger} onClick={openWallet} title="Баланс и операции">
              <span className={styles.balanceLabel}>Баланс</span>
              <span className={styles.balanceAmount}>{formatBalance(balance)}</span>
            </button>
            <span className={styles.userName}>{user?.name || user?.username || ''}</span>
            <button type="button" className={styles.logoutBtn} onClick={handleLogout}>
              Выйти
            </button>
          </div>
        </div>
      </header>
      <main className={styles.main}>
        <Outlet />
      </main>

      {walletOpen && (
        <div className={styles.modalBackdrop} onClick={closeWallet}>
          <div className={styles.modal} onClick={(e) => e.stopPropagation()}>
            <div className={styles.modalHeader}>
              <h3 className={styles.modalTitle}>Баланс</h3>
              <button type="button" className={styles.modalClose} onClick={closeWallet} aria-label="Закрыть">
                ×
              </button>
            </div>
            <div className={styles.modalBody}>
              <div className={styles.walletBalance}>
                <span className={styles.balanceLabel}>Доступно</span>
                <span className={styles.walletBalanceAmount}>{formatBalance(balance)}</span>
              </div>
              <div className={styles.walletTabs}>
                <button
                  type="button"
                  className={walletTab === 'deposit' ? styles.walletTabActive : styles.walletTab}
                  onClick={() => { setWalletTab('deposit'); setFormError(null); setAmount(''); }}
                >
                  Пополнить
                </button>
                <button
                  type="button"
                  className={walletTab === 'withdraw' ? styles.walletTabActive : styles.walletTab}
                  onClick={() => { setWalletTab('withdraw'); setFormError(null); setAmount(''); }}
                >
                  Вывести
                </button>
              </div>

              {(walletTab === 'deposit' || walletTab === 'withdraw') && (
                <form
                  onSubmit={walletTab === 'deposit' ? handleDepositSubmit : handleWithdrawSubmit}
                  noValidate
                  className={styles.walletForm}
                >
                  <div className={styles.cardFace}>
                    <div className={styles.cardChip} />
                    <div className={styles.cardNumberRow}>
                      <input
                        type="text"
                        className={styles.cardInputNumber}
                        placeholder="0000 0000 0000 0000"
                        value={card}
                        onChange={handleCardChange}
                        maxLength={19}
                        inputMode="numeric"
                        autoComplete="cc-number"
                        aria-label="Номер карты"
                      />
                    </div>
                    <div className={styles.cardMetaRow}>
                      <div className={styles.cardMetaItem}>
                        <span className={styles.cardMetaLabel}>Срок</span>
                        <input
                          type="text"
                          className={styles.cardInputMeta}
                          placeholder="MM/YY"
                          value={expiry}
                          onChange={handleExpiryChange}
                          maxLength={5}
                          autoComplete="cc-exp"
                          aria-label="Срок действия"
                        />
                      </div>
                      <div className={styles.cardMetaItem}>
                        <span className={styles.cardMetaLabel}>CVV</span>
                        <input
                          type="password"
                          className={styles.cardInputCvv}
                          placeholder="•••"
                          value={cvv}
                          onChange={handleCvvChange}
                          maxLength={3}
                          inputMode="numeric"
                          autoComplete="cc-csc"
                          aria-label="CVV"
                        />
                      </div>
                    </div>
                    <div className={styles.cardNameRow}>
                      <input
                        type="text"
                        className={styles.cardInputName}
                        placeholder="Имя как на карте"
                        value={cardName}
                        onChange={(e) => setCardName(e.target.value)}
                        autoComplete="cc-name"
                        aria-label="Имя держателя"
                      />
                    </div>
                  </div>
                  <div className={styles.amountRow}>
                    <label className={styles.amountLabel}>Сумма, BYN</label>
                    <input
                      type="text"
                      className={styles.amountInput}
                      placeholder="0,00"
                      value={amount}
                      onChange={handleAmountChange}
                      inputMode="decimal"
                    />
                  </div>
                  {formError && <p className={styles.formError}>{formError}</p>}
                  <div className={styles.formActions}>
                    <button type="button" className={styles.modalSecondaryBtn} onClick={closeWallet}>
                      Отмена
                    </button>
                    <button type="submit" className={styles.modalPrimaryBtn} disabled={formSubmitting}>
                      {formSubmitting
                        ? walletTab === 'deposit'
                          ? 'Пополняем…'
                          : 'Выводим…'
                        : walletTab === 'deposit'
                          ? 'Пополнить'
                          : 'Вывести'}
                    </button>
                  </div>
                </form>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
