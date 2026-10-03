import { router } from 'expo-router';
import { useState } from 'react';
import { StyleSheet, View } from 'react-native';

import { changeOwnPassword, type SessionUser } from '@/api/auth';
import { roleHome, roleLabel } from '@/auth/roles';
import { useSession } from '@/auth/SessionProvider';
import { AppText, Button, Card, Notice, Screen, TextField } from '@/components';
import { useFormSubmission } from '@/data/useFormSubmission';
import { passwordConfirmationError } from '@/features/accounts/passwordConfirmation';
import { teamName } from '@/mocks/catalog';
import { RoleGate } from '@/navigation/RoleGate';
import { spacing } from '@/theme';

const fields = ['currentPassword', 'password'] as const;

export default function AccountScreen() {
  return (
    <RoleGate roles={['fia_admin', 'team_admin']}>
      <AccountContent />
    </RoleGate>
  );
}

function AccountContent() {
  const { user } = useSession();
  if (!user) {
    return null;
  }
  const goBack = () => (router.canGoBack() ? router.back() : router.replace(roleHome[user.role]));

  return (
    <Screen title="Mi cuenta" headerAction={<Button label="Volver" icon="arrow-back" variant="ghost" compact onPress={goBack} />}>
      <AccountDetails user={user} />
      <ChangePasswordForm />
    </Screen>
  );
}

function AccountDetails({ user }: { user: SessionUser }) {
  const rows = [
    { label: 'Usuario', value: user.username },
    { label: 'Rol', value: roleLabel[user.role] },
    ...(user.teamId === null ? [] : [{ label: 'Escudería', value: teamName(user.teamId) }]),
  ];

  return (
    <Card title="Datos de la cuenta">
      {rows.map((row) => (
        <View key={row.label} style={styles.detail}>
          <AppText variant="overline" color="textMuted">
            {row.label}
          </AppText>
          <AppText>{row.value}</AppText>
        </View>
      ))}
    </Card>
  );
}

function ChangePasswordForm() {
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmation, setConfirmation] = useState('');
  const [isSaved, setIsSaved] = useState(false);
  const { fieldErrors, formError, isSubmitting, submit } = useFormSubmission(fields);
  const confirmationError = passwordConfirmationError(newPassword, confirmation);
  const canSubmit = currentPassword !== '' && newPassword !== '' && confirmation === newPassword;

  const save = async () => {
    setIsSaved(false);
    const result = await submit(async () => {
      await changeOwnPassword(currentPassword, newPassword);
      return true;
    });
    if (result) {
      setCurrentPassword('');
      setNewPassword('');
      setConfirmation('');
      setIsSaved(true);
    }
  };

  return (
    <Card title="Cambiar contraseña">
      {formError ? <Notice tone="danger" message={formError} /> : null}
      {isSaved ? <Notice tone="success" message="Contraseña actualizada. Se cerraron tus sesiones en otros dispositivos." /> : null}
      <TextField
        label="Contraseña actual"
        icon="lock-closed-outline"
        value={currentPassword}
        onChangeText={setCurrentPassword}
        error={fieldErrors.currentPassword}
        secureTextEntry
        autoCapitalize="none"
        autoComplete="current-password"
      />
      <TextField
        label="Nueva contraseña (12 caracteres o más)"
        icon="key-outline"
        value={newPassword}
        onChangeText={setNewPassword}
        error={fieldErrors.password}
        secureTextEntry
        autoCapitalize="none"
        autoComplete="new-password"
      />
      <TextField
        label="Repetir la nueva contraseña"
        icon="key-outline"
        value={confirmation}
        onChangeText={setConfirmation}
        error={confirmationError}
        secureTextEntry
        autoCapitalize="none"
        autoComplete="new-password"
      />
      <View style={styles.actions}>
        <Button label="Cambiar contraseña" icon="save-outline" onPress={save} loading={isSubmitting} disabled={!canSubmit} />
      </View>
    </Card>
  );
}

const styles = StyleSheet.create({
  detail: { gap: 2 },
  actions: { flexDirection: 'row', justifyContent: 'flex-end' },
});
