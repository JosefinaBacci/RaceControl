import { router, useLocalSearchParams } from 'expo-router';
import { useState } from 'react';
import { StyleSheet, View } from 'react-native';

import { releaseDriver, updateDriver, type TeamDriver } from '@/api/drivers';
import { AppText, Badge, Button, Card, Notice, Screen } from '@/components';
import { useTeamDriver } from '@/data/drivers';
import { useFormSubmission } from '@/data/useFormSubmission';
import { DriverFields } from '@/features/drivers/DriverFields';
import { driverChanges, driverFields, driverNumberError, hasDriverChanges, initialDriverForm } from '@/features/drivers/driverForm';
import { RemoteContent } from '@/features/RemoteContent';
import { spacing } from '@/theme';

export default function EditDriverScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const { state, reload, setData } = useTeamDriver(Number(id));

  return (
    <Screen
      title="Editar piloto"
      headerAction={<Button label="Volver" icon="arrow-back" variant="ghost" compact onPress={() => router.back()} />}
    >
      <RemoteContent state={state} onRetry={reload}>
        {(driver) => (
          <>
            <DriverForm driver={driver} onSaved={setData} />
            <ReleaseCard driver={driver} />
          </>
        )}
      </RemoteContent>
    </Screen>
  );
}

function DriverForm({ driver, onSaved }: { driver: TeamDriver; onSaved: (driver: TeamDriver) => void }) {
  const [values, setValues] = useState(initialDriverForm(driver));
  const [savedMessage, setSavedMessage] = useState<string | null>(null);
  const { fieldErrors, formError, isSubmitting, submit, setFormError } = useFormSubmission(driverFields);

  const save = async () => {
    setSavedMessage(null);
    const numberError = driverNumberError(values.number, driver.number);
    if (numberError) {
      setFormError(numberError);
      return;
    }
    const changes = driverChanges(driver, values);
    if (!hasDriverChanges(changes)) {
      setFormError('No hay cambios para guardar.');
      return;
    }
    const updated = await submit(() => updateDriver(driver.id, changes));
    if (updated) {
      setValues(initialDriverForm(updated));
      setSavedMessage('Cambios guardados.');
      onSaved(updated);
    }
  };

  return (
    <Card
      title={`${driver.firstName} ${driver.lastName}`}
      action={<Badge label={driver.isStarter ? 'Titular' : 'Suplente'} tone={driver.isStarter ? 'accent' : 'neutral'} />}
    >
      {formError ? <Notice tone="danger" message={formError} /> : null}
      {savedMessage ? <Notice tone="success" message={savedMessage} /> : null}
      <DriverFields values={values} onChange={setValues} errors={fieldErrors} />
      <View style={styles.actions}>
        <Button label="Guardar cambios" icon="save-outline" onPress={save} loading={isSubmitting} />
      </View>
    </Card>
  );
}

function ReleaseCard({ driver }: { driver: TeamDriver }) {
  const [isConfirming, setIsConfirming] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isReleasing, setIsReleasing] = useState(false);

  const release = async () => {
    setError(null);
    setIsReleasing(true);
    try {
      await releaseDriver(driver.id);
      router.back();
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : 'No se pudo dar de baja al piloto.');
    } finally {
      setIsReleasing(false);
    }
  };

  return (
    <Card title="Baja del piloto">
      {error ? <Notice tone="danger" message={error} /> : null}
      {isConfirming ? (
        <>
          <Notice tone="warning" message="El piloto deja la alineación desde hoy. El historial de asignaciones se conserva." />
          <View style={styles.actions}>
            <Button label="Cancelar" variant="ghost" onPress={() => setIsConfirming(false)} disabled={isReleasing} />
            <Button label="Confirmar baja" icon="person-remove-outline" onPress={release} loading={isReleasing} />
          </View>
        </>
      ) : (
        <>
          <AppText color="textMuted">Un piloto dado de baja sale de la alineación, pero su historial queda registrado.</AppText>
          <View style={styles.actions}>
            <Button label="Dar de baja" icon="person-remove-outline" variant="secondary" onPress={() => setIsConfirming(true)} />
          </View>
        </>
      )}
    </Card>
  );
}

const styles = StyleSheet.create({
  actions: { flexDirection: 'row', justifyContent: 'flex-end', gap: spacing.md },
});
