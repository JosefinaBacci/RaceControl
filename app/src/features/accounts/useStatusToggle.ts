import { useState } from 'react';

import { deactivateAccount, reactivateAccount, type Account } from '@/api/users';
import { useFormSubmission } from '@/data/useFormSubmission';

export function useStatusToggle(account: Account, onChanged: (account: Account) => void) {
  const [isConfirming, setIsConfirming] = useState(false);
  const { formError, isSubmitting, submit } = useFormSubmission([]);

  const toggle = async () => {
    const action = account.isActive ? deactivateAccount : reactivateAccount;
    const updated = await submit(() => action(account.id));
    setIsConfirming(false);
    if (updated) {
      onChanged(updated);
    }
  };

  return {
    isConfirming,
    isSubmitting,
    error: formError,
    ask: () => setIsConfirming(true),
    cancel: () => setIsConfirming(false),
    toggle,
  };
}
