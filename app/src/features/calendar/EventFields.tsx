import type { ReactNode } from 'react';
import { StyleSheet, View } from 'react-native';

import type { Circuit, Season } from '@/api/calendar';
import type { FieldErrors } from '@/api/formFailure';
import { AppText, ChipGroup, TextField } from '@/components';
import { categories } from '@/mocks/catalog';
import { eventKindOptions, kindLabel } from '@/features/eventKind';
import { spacing } from '@/theme';

import type { EventFormValues, EventFormField } from './eventForm';

type EventFieldsProps = {
  values: EventFormValues;
  onChange: (values: EventFormValues) => void;
  errors: FieldErrors<EventFormField>;
  seasons: Season[];
  circuits: Circuit[];
  isKindLocked?: boolean;
};

const categoryOptions = categories.map((category) => ({ value: category.value, label: category.label }));

export function EventFields({ values, onChange, errors, seasons, circuits, isKindLocked = false }: EventFieldsProps) {
  const update = (partial: Partial<EventFormValues>) => onChange({ ...values, ...partial });

  return (
    <View style={styles.fields}>
      <FieldGroup label="Tipo de evento" error={errors.kind}>
        {isKindLocked ? (
          <AppText color="textMuted">{kindLabel(values.kind)} · no se puede cambiar.</AppText>
        ) : (
          <ChipGroup
            options={eventKindOptions}
            selected={values.kind}
            onSelect={(kind) => update({ kind })}
            accessibilityLabel="Elegir tipo de evento"
          />
        )}
      </FieldGroup>

      <FieldGroup label="Categoría" error={errors.categoryCode}>
        <ChipGroup
          options={categoryOptions}
          selected={values.categoryCode}
          onSelect={(categoryCode) => update({ categoryCode })}
          accessibilityLabel="Elegir categoría"
        />
      </FieldGroup>

      <FieldGroup label="Temporada" error={errors.seasonId}>
        <ChipGroup
          options={seasons.map((season) => ({ value: String(season.id), label: String(season.year) }))}
          selected={String(values.seasonId)}
          onSelect={(seasonId) => update({ seasonId: Number(seasonId) })}
          accessibilityLabel="Elegir temporada"
        />
      </FieldGroup>

      {values.kind === 'technical_control' ? null : (
        <FieldGroup label="Circuito" error={errors.circuitId}>
          <ChipGroup
            options={circuits.map((circuit) => ({ value: String(circuit.id), label: circuit.name }))}
            selected={values.circuitId === null ? '' : String(values.circuitId)}
            onSelect={(circuitId) => update({ circuitId: Number(circuitId) })}
            accessibilityLabel="Elegir circuito"
          />
        </FieldGroup>
      )}

      <TextField
        label="Fecha (AAAA-MM-DD)"
        icon="calendar-outline"
        value={values.date}
        onChangeText={(date) => update({ date })}
        error={errors.startsAt}
        placeholder="2026-11-15"
        autoCapitalize="none"
        autoCorrect={false}
      />
      <TextField
        label="Hora de inicio (HH:MM)"
        icon="time-outline"
        value={values.startTime}
        onChangeText={(startTime) => update({ startTime })}
        placeholder="15:00"
        autoCapitalize="none"
        autoCorrect={false}
      />
      <TextField
        label="Hora de fin (HH:MM)"
        icon="time-outline"
        value={values.endTime}
        onChangeText={(endTime) => update({ endTime })}
        error={errors.endsAt}
        placeholder="17:00"
        autoCapitalize="none"
        autoCorrect={false}
      />

      {values.kind === 'race' ? (
        <>
          <TextField
            label="Número de ronda"
            icon="flag-outline"
            value={values.number}
            onChangeText={(number) => update({ number })}
            error={errors.number}
            keyboardType="number-pad"
            placeholder="19"
          />
          <TextField
            label="Nombre del Gran Premio"
            icon="trophy-outline"
            value={values.gpName}
            onChangeText={(gpName) => update({ gpName })}
            error={errors.gpName}
            placeholder="Gran Premio de Brasil"
          />
        </>
      ) : null}

      {values.kind === 'technical_control' ? (
        <TextField
          label="Tipo de control"
          icon="construct-outline"
          value={values.controlType}
          onChangeText={(controlType) => update({ controlType })}
          error={errors.controlType}
          placeholder="Verificación de pesos"
        />
      ) : null}

      <AppText variant="caption" color="textMuted">
        Las fechas se cargan en UTC, la misma zona con la que el sistema compone la etiqueta del evento.
      </AppText>
    </View>
  );
}

function FieldGroup({ label, error, children }: { label: string; error?: string; children: ReactNode }) {
  return (
    <View style={styles.group}>
      <AppText variant="bodyStrong">{label}</AppText>
      {children}
      {error ? (
        <AppText variant="caption" color="danger">
          {error}
        </AppText>
      ) : null}
    </View>
  );
}

const styles = StyleSheet.create({
  fields: { gap: spacing.md },
  group: { gap: spacing.sm },
});
