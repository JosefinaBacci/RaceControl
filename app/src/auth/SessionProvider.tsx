import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from 'react';

import * as authApi from '@/api/auth';
import type { SessionUser } from '@/api/auth';
import { ApiError } from '@/api/client';

type SessionState =
  | { status: 'loading'; user: null }
  | { status: 'anonymous'; user: null }
  | { status: 'authenticated'; user: SessionUser };

type SessionContextValue = SessionState & {
  signIn: (username: string, password: string) => Promise<SessionUser>;
  signOut: () => Promise<void>;
};

const SessionContext = createContext<SessionContextValue | null>(null);

const anonymous: SessionState = { status: 'anonymous', user: null };

export function SessionProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<SessionState>({ status: 'loading', user: null });

  useEffect(() => {
    let isMounted = true;
    restoreSession().then((restored) => {
      if (isMounted) {
        setState(restored);
      }
    });
    return () => {
      isMounted = false;
    };
  }, []);

  const signIn = useCallback(async (username: string, password: string) => {
    const user = await authApi.login(username, password);
    setState({ status: 'authenticated', user });
    return user;
  }, []);

  const signOut = useCallback(async () => {
    try {
      await authApi.logout();
    } catch (error) {
      if (!(error instanceof ApiError) || error.code !== 'unauthenticated') {
        throw error;
      }
    } finally {
      setState(anonymous);
    }
  }, []);

  const value = useMemo(() => ({ ...state, signIn, signOut }), [state, signIn, signOut]);

  return <SessionContext.Provider value={value}>{children}</SessionContext.Provider>;
}

async function restoreSession(): Promise<SessionState> {
  try {
    return { status: 'authenticated', user: await authApi.fetchCurrentUser() };
  } catch {
    return anonymous;
  }
}

export function useSession(): SessionContextValue {
  const context = useContext(SessionContext);
  if (!context) {
    throw new Error('useSession must be used inside SessionProvider');
  }
  return context;
}
