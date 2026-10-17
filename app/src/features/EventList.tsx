import { router } from 'expo-router';
import type { ReactNode } from 'react';
import { StyleSheet, View, useWindowDimensions } from 'react-native';

import type { CalendarEvent } from '@/api/calendar';
import { AppText, Badge, Card, ListItem } from '@/components';
import { eventKindTones, kindLabel } from '@/features/eventKind';
import { CircuitMap, Flag, circuitKeyFor } from '@/illustrations';
import { colors, radius } from '@/theme';

const monthFormatter = new Intl.DateTimeFormat('es-AR', { month: 'short' });

function DateBlock({ isoDate }: { isoDate: string }) {
  const [year, month, day] = isoDate.slice(0, 10).split('-').map(Number);
  return (
    <View style={styles.dateBlock}>
      <AppText variant="heading">{day}</AppText>
      <AppText variant="overline" color="textMuted">
        {monthFormatter.format(new Date(year, month - 1, day)).replace('.', '')}
      </AppText>
    </View>
  );
}

type EventListProps = {
  title?: string;
  events: CalendarEvent[];
  action?: ReactNode;
  emptyMessage?: string;
};

const mapBreakpoint = 560;

export function EventList({ title, events, action, emptyMessage = 'No hay eventos para mostrar.' }: EventListProps) {
  const showsMap = useWindowDimensions().width >= mapBreakpoint;

  return (
    <Card title={title} action={action}>
      {events.length === 0 ? (
        <AppText color="textMuted">{emptyMessage}</AppText>
      ) : (
        events.map((event, index) => (
          <ListItem
            key={event.id}
            leading={
              <View style={styles.leading}>
                <DateBlock isoDate={event.startsAt} />
                <Flag code={event.countryCode} width={26} />
              </View>
            }
            title={event.name}
            onPress={() => router.push({ pathname: '/events/[id]', params: { id: String(event.id) } })}
            subtitle={`${event.circuitName ?? kindLabel(event.kind)} · ${event.categoryName}`}
            trailing={
              <View style={styles.trailing}>
                {showsMap ? (
                  <CircuitMap circuit={circuitKeyFor(event.circuitName ?? '')} width={64} strokeWidth={7} color={colors.textMuted} />
                ) : null}
                <Badge
                  label={event.status === 'finished' ? 'Finalizado' : kindLabel(event.kind)}
                  tone={event.status === 'finished' ? 'neutral' : eventKindTones[event.kind]}
                />
              </View>
            }
            isLast={index === events.length - 1}
          />
        ))
      )}
    </Card>
  );
}

const styles = StyleSheet.create({
  leading: { flexDirection: 'row', alignItems: 'center', gap: 12 },
  trailing: { flexDirection: 'row', alignItems: 'center', gap: 12 },
  dateBlock: {
    width: 52,
    paddingVertical: 6,
    borderRadius: radius.md,
    backgroundColor: colors.surfaceRaised,
    borderWidth: 1,
    borderColor: colors.border,
    alignItems: 'center',
  },
});
