import { router } from 'expo-router';
import { useState } from 'react';
import { StyleSheet, View } from 'react-native';

import { createAccount } from '@/api/users';
import { Button, Card, Notice, Screen, TextField } from '@/components';
import { useFormSubmission } from '@/data/useFormSubmission';
import { AccountFields, type AccountFieldValues } from '@/features/accounts/AccountFields';
import { spacing } from '@/theme';

const fields = ['username', 'password', 'email', 'role', 'teamId'] as const;

export default function NewUserScreen() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [values, setValues] = useState<AccountFieldValues>({ email: '', role: 'team_admin', teamId: null });
  const { fieldErrors, formError, isSubmitting, submit } = useFormSubmission(fields);

  const save = async () => {
    const created = await submit(() => createAccount({ username, password, ...values }));
    if (created) {
      router.back();
    }
  };

  return (
    <Screen title="Nuevo usuario" subtitle="La contraseña inicial debe tener al menos 12 caracteres.">
      <Card>
        {formError ? <Notice tone="danger" message={formError} /> : null}
        <TextField
          label="Usuario"
          icon="person-outline"
          value={username}
          onChangeText={setUsername}
          error={fieldErrors.username}
          autoCapitalize="none"
          autoCorrect={false}
          maxLength={32}
        />
        <TextField
          label="Contraseña inicial"
          icon="lock-closed-outline"
          value={password}
          onChangeText={setPassword}
          error={fieldErrors.password}
          secureTextEntry
          autoCapitalize="none"
          autoComplete="new-password"
        />
        <AccountFields values={values} onChange={setValues} errors={fieldErrors} />
      </Card>
      <View style={styles.actions}>
        <Button label="Cancelar" variant="ghost" onPress={() => router.back()} disabled={isSubmitting} />
        <Button label="Crear usuario" icon="person-add-outline" onPress={save} loading={isSubmitting} />
      </View>
    </Screen>
  );
}

const styles = StyleSheet.create({
  actions: { flexDirection: 'row', justifyContent: 'flex-end', gap: spacing.md },
});
