import { router, useLocalSearchParams } from 'expo-router';
import { useState } from 'react';
import { StyleSheet, View } from 'react-native';

import { cancelEvent, updateEvent, type CalendarEvent, type Circuit, type Season } from '@/api/calendar';
import { AppText, Badge, Button, Card, Notice, Screen } from '@/components';
import { useEventForm } from '@/data/calendar';
import { useFormSubmission } from '@/data/useFormSubmission';
import { EventFields } from '@/features/calendar/EventFields';
import { hasStarted } from '@/features/calendar/eventTiming';
import {
  eventFormFields,
  formValuesOf,
  hasChanges,
  timingOf,
  toEventChanges,
  type EventFormValues,
} from '@/features/calendar/eventForm';
import { RemoteContent } from '@/features/RemoteContent';
import { spacing } from '@/theme';

export default function EditEventScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const { state, reload, setData } = useEventForm(Number(id));

  return (
    <Screen
      title="Editar evento"
      headerAction={<Button label="Volver" icon="arrow-back" variant="ghost" compact onPress={() => router.back()} />}
    >
      <RemoteContent state={state} onRetry={reload}>
        {(loaded) =>
          loaded.event === null ? (
            <Notice tone="danger" message="No se encontró el evento." />
          ) : (
            <>
              <EventEditForm
                event={loaded.event}
                seasons={loaded.seasons}
                circuits={loaded.circuits}
                onSaved={(event) => setData({ ...loaded, event })}
              />
              <CancelCard event={loaded.event} />
            </>
          )
        }
      </RemoteContent>
    </Screen>
  );
}

type EventEditFormProps = {
  event: CalendarEvent;
  seasons: Season[];
  circuits: Circuit[];
  onSaved: (event: CalendarEvent) => void;
};

function EventEditForm({ event, seasons, circuits, onSaved }: EventEditFormProps) {
  const [values, setValues] = useState<EventFormValues>(() => formValuesOf(event));
  const [savedMessage, setSavedMessage] = useState<string | null>(null);
  const { fieldErrors, formError, isSubmitting, submit, setFormError } = useFormSubmission(eventFormFields);

  const save = async () => {
    setSavedMessage(null);
    const timing = timingOf(values);
    if (!timing) {
      setFormError('Revisá la fecha (AAAA-MM-DD) y las horas (HH:MM).');
      return;
    }
    const changes = toEventChanges(event, values, timing);
    if (!hasChanges(changes)) {
      setFormError('No hay cambios para guardar.');
      return;
    }
    const updated = await submit(() => updateEvent(event.id, changes));
    if (updated) {
      setValues(formValuesOf(updated));
      setSavedMessage('Cambios guardados.');
      onSaved(updated);
    }
  };

  return (
    <Card title={event.name}>
      {formError ? <Notice tone="danger" message={formError} /> : null}
      {savedMessage ? <Notice tone="success" message={savedMessage} /> : null}
      <EventFields values={values} onChange={setValues} errors={fieldErrors} seasons={seasons} circuits={circuits} isKindLocked />
      <View style={styles.actions}>
        <Button label="Guardar cambios" icon="save-outline" onPress={save} loading={isSubmitting} />
      </View>
    </Card>
  );
}

function CancelCard({ event }: { event: CalendarEvent }) {
  const [isConfirming, setIsConfirming] = useState(false);
  const { formError, isSubmitting, submit } = useFormSubmission([]);

  const cancel = async () => {
    const cancelled = await submit(() => cancelEvent(event.id));
    if (cancelled) {
      router.back();
    }
  };

  return (
    <Card
      title="Cancelación"
      action={<Badge label={event.status === 'finished' ? 'Finalizado' : 'Programado'} tone="neutral" />}
    >
      {formError ? <Notice tone="danger" message={formError} /> : null}
      {hasStarted(event) ? (
        <AppText color="textMuted">El evento ya empezó: forma parte de la temporada y no se puede cancelar.</AppText>
      ) : (
        <CancelControls
          event={event}
          isConfirming={isConfirming}
          isSubmitting={isSubmitting}
          onConfirming={setIsConfirming}
          onCancel={cancel}
        />
      )}
    </Card>
  );
}

type CancelControlsProps = {
  event: CalendarEvent;
  isConfirming: boolean;
  isSubmitting: boolean;
  onConfirming: (isConfirming: boolean) => void;
  onCancel: () => void;
};

function CancelControls({ event, isConfirming, isSubmitting, onConfirming, onCancel }: CancelControlsProps) {
  return (
    <>
      <AppText color="textMuted">
        {event.kind === 'race'
          ? 'Cancelar una carrera la saca del calendario junto con su ronda.'
          : 'Cancelar el evento lo saca del calendario oficial.'}
      </AppText>
      {isConfirming ? (
        <>
          <Notice tone="warning" message="El evento va a dejar de aparecer en el calendario. No se puede deshacer." />
          <View style={styles.actions}>
            <Button label="Volver" variant="ghost" onPress={() => onConfirming(false)} disabled={isSubmitting} />
            <Button label="Confirmar cancelación" icon="trash-outline" onPress={onCancel} loading={isSubmitting} />
          </View>
        </>
      ) : (
        <View style={styles.actions}>
          <Button label="Cancelar evento" icon="trash-outline" variant="secondary" onPress={() => onConfirming(true)} />
        </View>
      )}
    </>
  );
}

const styles = StyleSheet.create({
  actions: { flexDirection: 'row', justifyContent: 'flex-end', gap: spacing.md },
});
