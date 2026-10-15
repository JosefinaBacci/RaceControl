import { useState } from 'react';
import { StyleSheet, View } from 'react-native';

import { cancelEvent, type CalendarEvent } from '@/api/calendar';
import { AppText, Badge, Card, IconButton } from '@/components';
import { useFormSubmission } from '@/data/useFormSubmission';
import { colors, spacing } from '@/theme';

import { kindLabel } from './eventForm';

type EventAdminListProps = {
  events: CalendarEvent[];
  onEdit: (event: CalendarEvent) => void;
  onChanged: () => void;
};

export function EventAdminList({ events, onEdit, onChanged }: EventAdminListProps) {
  return (
    <Card flush>
      <TableHeader />
      {events.length === 0 ? (
        <AppText color="textMuted" style={styles.empty}>
          No hay eventos para mostrar.
        </AppText>
      ) : (
        events.map((event) => <EventRow key={event.id} event={event} onEdit={onEdit} onChanged={onChanged} />)
      )}
    </Card>
  );
}

function TableHeader() {
  return (
    <View style={[styles.row, styles.headerRow]}>
      <AppText variant="overline" color="textMuted" style={styles.eventColumn}>
        Evento
      </AppText>
      <AppText variant="overline" color="textMuted" style={styles.stateColumn}>
        Estado
      </AppText>
      <AppText variant="overline" color="textMuted" style={styles.actionsColumn}>
        Acciones
      </AppText>
    </View>
  );
}

type EventRowProps = {
  event: CalendarEvent;
  onEdit: (event: CalendarEvent) => void;
  onChanged: () => void;
};

function EventRow({ event, onEdit, onChanged }: EventRowProps) {
  return (
    <View style={styles.row}>
      <View style={styles.eventColumn}>
        <AppText variant="bodyStrong" numberOfLines={1}>
          {event.name}
        </AppText>
        <AppText variant="caption" color="textMuted" numberOfLines={1}>
          {subtitleOf(event)}
        </AppText>
      </View>
      <View style={styles.stateColumn}>
        <Badge
          label={event.status === 'finished' ? 'Finalizado' : 'Programado'}
          tone={event.status === 'finished' ? 'neutral' : 'accent'}
        />
      </View>
      <EventRowActions event={event} onEdit={onEdit} onChanged={onChanged} />
    </View>
  );
}

function EventRowActions({ event, onEdit, onChanged }: EventRowProps) {
  const [isConfirming, setIsConfirming] = useState(false);
  const { formError, isSubmitting, submit } = useFormSubmission([]);

  const cancel = async () => {
    const cancelled = await submit(() => cancelEvent(event.id));
    setIsConfirming(false);
    if (cancelled) {
      onChanged();
    }
  };

  if (isConfirming) {
    return (
      <View style={[styles.actionsColumn, styles.actionsColumnStacked]}>
        <AppText variant="caption" color="danger">
          ¿Cancelar?
        </AppText>
        <View style={styles.actions}>
          <IconButton icon="close" accessibilityLabel="Volver" onPress={() => setIsConfirming(false)} disabled={isSubmitting} />
          <IconButton
            icon="checkmark"
            tone="danger"
            accessibilityLabel={`Confirmar la cancelación de ${event.name}`}
            onPress={cancel}
            disabled={isSubmitting}
          />
        </View>
      </View>
    );
  }

  return (
    <View style={[styles.actionsColumn, styles.actionsColumnStacked]}>
      <View style={styles.actions}>
        <IconButton icon="create-outline" accessibilityLabel={`Editar ${event.name}`} onPress={() => onEdit(event)} />
        <IconButton
          icon="trash-outline"
          tone="danger"
          accessibilityLabel={`Cancelar ${event.name}`}
          onPress={() => setIsConfirming(true)}
        />
      </View>
      {formError ? (
        <AppText variant="caption" color="danger">
          {formError}
        </AppText>
      ) : null}
    </View>
  );
}

function subtitleOf(event: CalendarEvent): string {
  const place = event.circuitName ?? kindLabel(event.kind);
  return `${event.categoryName} · ${place} · ${event.startsAt.slice(0, 10)}`;
}

const styles = StyleSheet.create({
  row: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.md,
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.md,
    borderTopWidth: StyleSheet.hairlineWidth,
    borderTopColor: colors.border,
  },
  headerRow: { backgroundColor: colors.surfaceRaised, borderTopWidth: 0, paddingVertical: spacing.sm },
  eventColumn: { flex: 3 },
  stateColumn: { flex: 1.2 },
  actionsColumn: { flex: 1.6 },
  actionsColumnStacked: { alignItems: 'flex-end', gap: spacing.xs },
  actions: { flexDirection: 'row', gap: spacing.sm },
  empty: { padding: spacing.lg },
});
