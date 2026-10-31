import type { PublicDriver } from '@/api/drivers';
import type { PenaltyType } from '@/api/sanctions';
import type { TeamSummary } from '@/api/teams';
import { ChipGroup, FieldGroup, SelectField, TextField } from '@/components';
import { useChampionshipEvents } from '@/data/sanctions';
import { RemoteContent } from '@/features/RemoteContent';
import { TeamPicker } from '@/features/TeamPicker';

import { penaltyTypeOf, type SanctionFormErrors, type SanctionFormValues, type SubjectKind } from './sanctionForm';
import { unitLabel } from './sanctionText';

type FieldsProps = {
  values: SanctionFormValues;
  onChange: (values: SanctionFormValues) => void;
  errors: SanctionFormErrors;
};

const subjectOptions: readonly { value: SubjectKind; label: string }[] = [
  { value: 'driver', label: 'Piloto' },
  { value: 'team', label: 'Escudería' },
];

const noEvent = '';

export function SubjectFields({ values, onChange, errors, drivers }: FieldsProps & { drivers: readonly PublicDriver[] }) {
  const driverOptions = drivers.map((driver) => ({
    value: String(driver.id),
    label: `${driver.firstName} ${driver.lastName} · ${driver.teamName}`,
  }));

  return (
    <>
      <FieldGroup label="Destinatario" error={errors.target}>
        <ChipGroup
          options={subjectOptions}
          selected={values.subjectKind}
          onSelect={(subjectKind) => onChange({ ...values, subjectKind, eventId: null })}
          accessibilityLabel="Elegir a quién se sanciona"
        />
      </FieldGroup>
      {values.subjectKind === 'driver' ? (
        <FieldGroup label="Piloto" error={errors.driverId}>
          <SelectField
            options={driverOptions}
            selected={values.driverId === null ? '' : String(values.driverId)}
            onSelect={(driverId) => onChange({ ...values, driverId: Number(driverId), eventId: null })}
            accessibilityLabel="Elegir piloto"
          />
        </FieldGroup>
      ) : (
        <TeamPicker teamId={values.teamId} onChange={(teamId) => onChange({ ...values, teamId, eventId: null })} error={errors.teamId} />
      )}
    </>
  );
}

type TermsFieldsProps = FieldsProps & {
  penaltyTypes: readonly PenaltyType[];
  categoryCode: string | null;
};

export function TermsFields({ values, onChange, errors, penaltyTypes, categoryCode }: TermsFieldsProps) {
  const penaltyType = penaltyTypeOf(values, penaltyTypes);

  return (
    <>
      <FieldGroup label="Penalización" error={errors.penaltyType}>
        <SelectField
          options={penaltyTypes.map((type) => ({ value: type.code, label: type.name }))}
          selected={values.penaltyType}
          onSelect={(code) => onChange({ ...values, penaltyType: code, penaltyValue: '' })}
          accessibilityLabel="Elegir penalización"
        />
      </FieldGroup>
      {penaltyType?.unit ? (
        <TextField
          label={`Valor en ${unitLabel(penaltyType.unit)} (1 a ${penaltyType.maxValue})`}
          value={values.penaltyValue}
          onChangeText={(penaltyValue) => onChange({ ...values, penaltyValue })}
          error={errors.penaltyValue}
          inputMode="numeric"
        />
      ) : null}
      <TextField
        label="Motivo"
        value={values.reason}
        onChangeText={(reason) => onChange({ ...values, reason })}
        error={errors.reason}
        multiline
      />
      <EventField values={values} onChange={onChange} errors={errors} categoryCode={categoryCode} />
    </>
  );
}

function EventField({ values, onChange, errors, categoryCode }: FieldsProps & { categoryCode: string | null }) {
  const { state, reload } = useChampionshipEvents(categoryCode);

  return (
    <FieldGroup label="Evento (opcional)" error={errors.eventId}>
      <RemoteContent state={state} onRetry={reload}>
        {(events) => (
          <SelectField
            options={[{ value: noEvent, label: 'Sin evento asociado' }, ...events.map((event) => ({ value: String(event.id), label: event.name }))]}
            selected={values.eventId === null ? noEvent : String(values.eventId)}
            onSelect={(eventId) => onChange({ ...values, eventId: eventId === noEvent ? null : Number(eventId) })}
            accessibilityLabel="Elegir evento"
          />
        )}
      </RemoteContent>
    </FieldGroup>
  );
}
