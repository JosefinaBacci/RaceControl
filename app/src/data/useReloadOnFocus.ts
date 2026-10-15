import { useFocusEffect } from 'expo-router';
import { useCallback, useRef } from 'react';

// Coming back from the create or edit screen must show the change, but the first
// focus is already covered by the initial load.
export function useReloadOnFocus(reload: () => void) {
  const latestReload = useRef(reload);
  latestReload.current = reload;
  const isFirstFocus = useRef(true);

  useFocusEffect(
    useCallback(() => {
      if (isFirstFocus.current) {
        isFirstFocus.current = false;
        return;
      }
      latestReload.current();
    }, []),
  );
}
