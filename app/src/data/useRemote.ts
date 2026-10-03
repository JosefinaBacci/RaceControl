import { useCallback, useEffect, useRef, useState, type DependencyList } from 'react';

import { ApiError } from '@/api/client';

export type RemoteState<T> =
  | { status: 'loading' }
  | { status: 'ready'; data: T }
  | { status: 'error'; message: string };

export function useRemote<T>(load: () => Promise<T>, deps: DependencyList) {
  const [state, setState] = useState<RemoteState<T>>({ status: 'loading' });
  const latestRequest = useRef(0);
  const stableLoad = useCallback(load, deps);

  const reload = useCallback(async () => {
    const request = ++latestRequest.current;
    try {
      const data = await stableLoad();
      if (request === latestRequest.current) {
        setState({ status: 'ready', data });
      }
    } catch (error) {
      if (request === latestRequest.current) {
        setState({ status: 'error', message: errorMessage(error) });
      }
    }
  }, [stableLoad]);

  useEffect(() => {
    reload();
  }, [reload]);

  return { state, reload, setData: (data: T) => setState({ status: 'ready', data }) };
}

function errorMessage(error: unknown): string {
  return error instanceof ApiError ? error.message : 'No se pudieron cargar los datos.';
}
