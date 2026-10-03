import { router, useLocalSearchParams } from 'expo-router';
import { useState } from 'react';
import { StyleSheet, View } from 'react-native';

import { updateAccount, type Account } from '@/api/users';
import { useSession } from '@/auth/SessionProvider';
import { AppText, Badge, Button, Card, Notice, Screen, TextField } from '@/components';
import { useAccount } from '@/data/accounts';
import { useFormSubmission } from '@/data/useFormSubmission';
import { AccountFields, type AccountFieldValues } from '@/features/accounts/AccountFields';
import { accountChanges, hasChanges, initialFieldValues } from '@/features/accounts/accountChanges';
import { useStatusToggle } from '@/features/accounts/useStatusToggle';
import { RemoteContent } from '@/features/RemoteContent';
import { spacing } from '@/theme';

const fields = ['password', 'email', 'role', 'teamId'] as const;

export default function EditUserScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const { state, reload, setData } = useAccount(Number(id));

  return (
    <Screen title="Editar usuario" headerAction={<Button label="Volver" icon="arrow-back" variant="ghost" compact onPress={() => router.back()} />}>
      <RemoteContent state={state} onRetry={reload}>
        {(account) => (
          <>
            <AccountForm account={account} onSaved={setData} />
            <StatusCard account={account} onChanged={setData} />
          </>
        )}
      </RemoteContent>
    </Screen>
  );
}

function AccountForm({ account, onSaved }: { account: Account; onSaved: (account: Account) => void }) {
  const isSelf = useSession().user?.id === account.id;
  const [values, setValues] = useState<AccountFieldValues>(initialFieldValues(account));
  const [password, setPassword] = useState('');
  const [savedMessage, setSavedMessage] = useState<string | null>(null);
  const { fieldErrors, formError, isSubmitting, submit, setFormError } = useFormSubmission(fields);

  const save = async () => {
    setSavedMessage(null);
    const changes = accountChanges(account, values, password);
    if (!hasChanges(changes)) {
      setFormError('No hay cambios para guardar.');
      return;
    }
    const updated = await submit(() => updateAccount(account.id, changes));
    if (updated) {
      setValues(initialFieldValues(updated));
      setPassword('');
      setSavedMessage('Cambios guardados.');
      onSaved(updated);
    }
  };

  return (
    <Card title={account.username}>
      {formError ? <Notice tone="danger" message={formError} /> : null}
      {savedMessage ? <Notice tone="success" message={savedMessage} /> : null}
      <AccountFields values={values} onChange={setValues} errors={fieldErrors} isRoleLocked={isSelf} />
      <TextField
        label="Nueva contraseña (dejar vacío para no cambiarla)"
        icon="lock-closed-outline"
        value={password}
        onChangeText={setPassword}
        error={fieldErrors.password}
        secureTextEntry
        autoCapitalize="none"
        autoComplete="new-password"
      />
      <AppText variant="caption" color="textMuted">
        Cambiar el rol, la escudería o la contraseña cierra las sesiones abiertas del usuario.
      </AppText>
      <View style={styles.actions}>
        <Button label="Guardar cambios" icon="save-outline" onPress={save} loading={isSubmitting} />
      </View>
    </Card>
  );
}

function StatusCard({ account, onChanged }: { account: Account; onChanged: (account: Account) => void }) {
  const isSelf = useSession().user?.id === account.id;
  const status = useStatusToggle(account, onChanged);

  return (
    <Card
      title="Estado de la cuenta"
      action={<Badge label={account.isActive ? 'Activa' : 'Desactivada'} tone={account.isActive ? 'success' : 'neutral'} />}
    >
      {status.error ? <Notice tone="danger" message={status.error} /> : null}
      {isSelf ? (
        <AppText color="textMuted">No podés desactivar tu propia cuenta.</AppText>
      ) : account.isActive ? (
        <DeactivateControls
          isConfirming={status.isConfirming}
          isSubmitting={status.isSubmitting}
          onAsk={status.ask}
          onCancel={status.cancel}
          onConfirm={status.toggle}
        />
      ) : (
        <View style={styles.actions}>
          <Button label="Reactivar cuenta" icon="refresh" variant="secondary" onPress={status.toggle} loading={status.isSubmitting} />
        </View>
      )}
    </Card>
  );
}

type DeactivateControlsProps = {
  isConfirming: boolean;
  isSubmitting: boolean;
  onAsk: () => void;
  onCancel: () => void;
  onConfirm: () => void;
};

function DeactivateControls({ isConfirming, isSubmitting, onAsk, onCancel, onConfirm }: DeactivateControlsProps) {
  if (!isConfirming) {
    return (
      <View style={styles.actions}>
        <Button label="Desactivar cuenta" icon="ban-outline" variant="secondary" onPress={onAsk} />
      </View>
    );
  }
  return (
    <>
      <Notice tone="warning" message="El usuario va a perder el acceso de inmediato y se cerrarán todas sus sesiones." />
      <View style={styles.actions}>
        <Button label="Cancelar" variant="ghost" onPress={onCancel} disabled={isSubmitting} />
        <Button label="Confirmar desactivación" icon="ban-outline" onPress={onConfirm} loading={isSubmitting} />
      </View>
    </>
  );
}

const styles = StyleSheet.create({
  actions: { flexDirection: 'row', justifyContent: 'flex-end', gap: spacing.md },
});
