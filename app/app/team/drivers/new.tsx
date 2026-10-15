import { router } from 'expo-router';
import { useState } from 'react';
import { StyleSheet, View } from 'react-native';

import { createDriver } from '@/api/drivers';
import { Button, Card, Notice, Screen } from '@/components';
import { useFormSubmission } from '@/data/useFormSubmission';
import { DriverFields } from '@/features/drivers/DriverFields';
import { driverFields, emptyDriverForm, newDriverInput } from '@/features/drivers/driverForm';
import { spacing } from '@/theme';

export default function NewDriverScreen() {
  const [values, setValues] = useState(emptyDriverForm());
  const { fieldErrors, formError, isSubmitting, submit } = useFormSubmission(driverFields);

  const save = async () => {
    const created = await submit(() => createDriver(newDriverInput(values)));
    if (created) {
      router.back();
    }
  };

  return (
    <Screen title="Nuevo piloto" subtitle="Cargá los datos del piloto y su número para la temporada.">
      <Card>
        {formError ? <Notice tone="danger" message={formError} /> : null}
        <DriverFields values={values} onChange={setValues} errors={fieldErrors} />
      </Card>
      <View style={styles.actions}>
        <Button label="Cancelar" variant="ghost" onPress={() => router.back()} disabled={isSubmitting} />
        <Button label="Guardar piloto" icon="save-outline" onPress={save} loading={isSubmitting} />
      </View>
    </Screen>
  );
}

const styles = StyleSheet.create({
  actions: { flexDirection: 'row', justifyContent: 'flex-end', gap: spacing.md },
});
