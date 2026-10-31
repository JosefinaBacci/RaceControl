import { router } from 'expo-router';

import { AppText, Badge, Button, Card, ListItem } from '@/components';
import { useGradableControls } from '@/data/technicalControls';
import { formatInstantDay } from '@/features/history/day';
import { RemoteContent } from '@/features/RemoteContent';

import { controlResultsHref } from './routes';

// The controls that already ended, for the FIA to grade (US21). A new control is
// scheduled in the calendar, which owns its days and type.
export function FiaControlsCard() {
  const { state, reload } = useGradableControls();

  return (
    <Card
      title="Controles técnicos"
      action={
        <Button
          label="Nuevo control técnico"
          icon="add"
          variant="secondary"
          compact
          onPress={() => router.push({ pathname: '/fia/calendar/new', params: { kind: 'technical_control' } })}
        />
      }
    >
      <RemoteContent state={state} onRetry={reload}>
        {(controls) =>
          controls.length === 0 ? (
            <AppText color="textMuted">Todavía no terminó ningún control técnico.</AppText>
          ) : (
            controls.map((control, index) => (
              <ListItem
                key={control.id}
                title={control.controlType ?? control.name}
                subtitle={`${control.categoryName} · ${formatInstantDay(control.startsAt)}`}
                trailing={<Badge label="Registrar" tone="neutral" />}
                onPress={() => router.push(controlResultsHref(control.id))}
                isLast={index === controls.length - 1}
              />
            ))
          )
        }
      </RemoteContent>
    </Card>
  );
}
