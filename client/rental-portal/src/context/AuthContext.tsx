import React, { createContext, useCallback, useContext, useEffect, useState } from 'react';
import { api } from '../api/client';

type User = {
  id: number;
  username: string;
  email: string;
  name: string;
  zip_code?: string;
  phone?: string;
};

/** Имя пермишена в ответе API (snake_case). */
export const PERMISSION_SYSTEM_SETTINGS_WRITE = 'system_settings_write';

type AuthContextType = {
  user: User | null;
  userId: number | null;
  isLoading: boolean;
  /** Баланс пользователя в BYN (с бэкенда). */
  balance: number;
  /** Обновить баланс с бэкенда (опционально по id пользователя). */
  refreshBalance: (userId?: number) => Promise<void>;
  /** Пополнить счёт на указанную сумму (BYN). */
  topUp: (amount: number) => Promise<void>;
  /** Вывести сумму с баланса (BYN). Возвращает false, если недостаточно средств. */
  withdraw: (amount: number) => Promise<boolean>;
  /** Список пермишенов текущего пользователя (загружается после входа). */
  permissions: string[];
  /** Есть ли пермишен на изменение системных настроек (доступ в панель администратора). */
  hasSystemSettingsWrite: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (body: { email: string; password: string; username: string; name: string; zip_code?: string; phone?: string }) => Promise<void>;
  logout: () => void;
  refreshUser: (userId: number) => Promise<void>;
  updateProfile: (userId: number, body: { name?: string; zip_code?: string; phone?: string }) => Promise<void>;
};

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [userId, setUserId] = useState<number | null>(() => {
    const id = localStorage.getItem('user_id');
    return id ? parseInt(id, 10) : null;
  });
  const [permissions, setPermissions] = useState<string[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [balance, setBalance] = useState<number>(0);

  const refreshBalance = useCallback(async (id?: number) => {
    const idToUse = id ?? userId;
    if (idToUse == null) return;
    try {
      const data = await api.user.getBalance(idToUse);
      setBalance(typeof data.balance === 'number' ? data.balance : 0);
    } catch {
      setBalance(0);
    }
  }, [userId]);

  const topUp = useCallback(
    async (amount: number) => {
      if (userId == null || amount <= 0) return;
      const data = await api.user.topUp(userId, amount);
      setBalance(typeof data.balance === 'number' ? data.balance : 0);
    },
    [userId]
  );

  const withdraw = useCallback(
    async (amount: number): Promise<boolean> => {
      if (userId == null || amount <= 0) return false;
      try {
        const data = await api.user.withdraw(userId, amount);
        setBalance(typeof data.balance === 'number' ? data.balance : 0);
        return true;
      } catch {
        return false;
      }
    },
    [userId]
  );

  const refreshUser = useCallback(async (id: number) => {
    try {
      const data = await api.user.get(id);
      setUser(data);
    } catch {
      setUser(null);
    }
  }, []);

  const loadPermissions = useCallback(async () => {
    try {
      const list = await api.permissions.getMy();
      setPermissions(Array.isArray(list) ? list : []);
    } catch {
      setPermissions([]);
    }
  }, []);

  useEffect(() => {
    const token = localStorage.getItem('access_token');
    const id = localStorage.getItem('user_id');
    if (!token || !id) {
      setPermissions([]);
      setBalance(0);
      setIsLoading(false);
      return;
    }
    const idNum = parseInt(id, 10);
    setUserId(idNum);
    Promise.all([refreshUser(idNum), loadPermissions(), refreshBalance(idNum)]).finally(() => setIsLoading(false));
  }, [refreshUser, loadPermissions, refreshBalance]);

  const login = useCallback(
    async (email: string, password: string) => {
      const data = await api.auth.signIn(email, password);
      localStorage.setItem('access_token', data.access_token);
      localStorage.setItem('refresh_token', data.refresh_token);
      localStorage.setItem('user_id', String(data.user_id));
      setUserId(data.user_id);
      await Promise.all([refreshUser(data.user_id), loadPermissions(), refreshBalance(data.user_id)]);
    },
    [refreshUser, loadPermissions, refreshBalance]
  );

  const register = useCallback(
    async (body: { email: string; password: string; username: string; name: string; zip_code?: string; phone?: string }) => {
      const data = await api.auth.register(body);
      localStorage.setItem('access_token', data.access_token);
      localStorage.setItem('refresh_token', data.refresh_token);
      localStorage.setItem('user_id', String(data.user_id));
      setUserId(data.user_id);
      await Promise.all([refreshUser(data.user_id), loadPermissions(), refreshBalance(data.user_id)]);
    },
    [refreshUser, loadPermissions, refreshBalance]
  );

  const logout = useCallback(() => {
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    localStorage.removeItem('user_id');
    setUser(null);
    setUserId(null);
    setPermissions([]);
    setBalance(0);
  }, []);

  const updateProfile = useCallback(
    async (id: number, body: { name?: string; zip_code?: string; phone?: string }) => {
      await api.user.updateProfile(id, body);
      await refreshUser(id);
    },
    [refreshUser]
  );

  const hasSystemSettingsWrite = permissions.includes(PERMISSION_SYSTEM_SETTINGS_WRITE);

  return (
    <AuthContext.Provider
      value={{
        user,
        userId,
        isLoading,
        balance,
        refreshBalance,
        topUp,
        withdraw,
        permissions,
        hasSystemSettingsWrite,
        login,
        register,
        logout,
        refreshUser,
        updateProfile,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}
