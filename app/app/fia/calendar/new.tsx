import { router } from 'expo-router';
import { useState } from 'react';
import { StyleSheet, View } from 'react-native';

import { createEvent, type Circuit, type Season } from '@/api/calendar';
import { Button, Card, Notice, Screen } from '@/components';
import { useEventForm } from '@/data/calendar';
import { useFormSubmission } from '@/data/useFormSubmission';
import { EventFields } from '@/features/calendar/EventFields';
import { eventFormFields, newFormValues, timingOf, toNewEvent, type EventFormValues } from '@/features/calendar/eventForm';
import { RemoteContent } from '@/features/RemoteContent';
import { categories } from '@/mocks/catalog';
import { spacing } from '@/theme';

export default function NewEventScreen() {
  const { state, reload } = useEventForm();

  return (
    <Screen
      title="Nuevo evento"
      subtitle="Cargá una carrera, una prueba de neumáticos o un control técnico."
      headerAction={<Button label="Volver" icon="arrow-back" variant="ghost" compact onPress={() => router.back()} />}
    >
      <RemoteContent state={state} onRetry={reload}>
        {({ seasons, circuits }) =>
          seasons.length === 0 ? (
            <Notice tone="warning" message="Todavía no hay una temporada cargada para asociar el evento." />
          ) : (
            <NewEventForm seasonId={seasons[0].id} seasons={seasons} circuits={circuits} />
          )
        }
      </RemoteContent>
    </Screen>
  );
}

type NewEventFormProps = {
  seasonId: number;
  seasons: Season[];
  circuits: Circuit[];
};

function NewEventForm({ seasonId, seasons, circuits }: NewEventFormProps) {
  const [values, setValues] = useState<EventFormValues>(() => newFormValues(seasonId, categories[0].value));
  const { fieldErrors, formError, isSubmitting, submit, setFormError } = useFormSubmission(eventFormFields);

  const save = async () => {
    const timing = timingOf(values);
    if (!timing) {
      setFormError('Completá la fecha y las dos horas (por ejemplo 2026-11-15, 15:00 y 17:00).');
      return;
    }
    if (new Date(timing.startsAt).getTime() <= Date.now()) {
      setFormError('El inicio debe ser posterior a la fecha y hora actuales.');
      return;
    }
    const created = await submit(() => createEvent(toNewEvent(values, timing)));
    if (created) {
      router.back();
    }
  };

  return (
    <Card title="Datos del evento">
      {formError ? <Notice tone="danger" message={formError} /> : null}
      <EventFields values={values} onChange={setValues} errors={fieldErrors} seasons={seasons} circuits={circuits} />
      <View style={styles.actions}>
        <Button label="Crear evento" icon="save-outline" onPress={save} loading={isSubmitting} />
      </View>
    </Card>
  );
}

const styles = StyleSheet.create({
  actions: { flexDirection: 'row', justifyContent: 'flex-end', gap: spacing.md },
});
