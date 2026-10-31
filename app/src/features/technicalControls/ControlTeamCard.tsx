import { useState } from 'react';
import { StyleSheet, View } from 'react-native';

import { correctControlResult, recordControlResult, type ControlSheetTeam } from '@/api/technicalControls';
import { Badge, Button, Card, Notice, TextField } from '@/components';
import { ResultChain } from '@/features/history/ResultChain';
import { ResultFields } from '@/features/history/ResultFields';
import { spacing } from '@/theme';

import { useControlResultForm } from './useControlResultForm';

type ControlTeamCardProps = {
  eventId: number;
  team: ControlSheetTeam;
  defaultDay: string;
  onSaved: () => void;
};

// One team on the FIA sheet: its chain of results and the form that records the
// first one or corrects the one in force.
export function ControlTeamCard({ eventId, team, defaultDay, onSaved }: ControlTeamCardProps) {
  const current = team.revisions.at(-1);
  const [isCorrecting, setIsCorrecting] = useState(false);

  return (
    <Card title={team.name} action={current ? null : <Badge label="Pendiente" tone="warning" />}>
      {current ? <ResultChain revisions={team.revisions} /> : null}
      {!current ? <RecordForm eventId={eventId} teamId={team.id} defaultDay={defaultDay} onSaved={onSaved} /> : null}
      {current && isCorrecting ? (
        <CorrectForm resultId={current.id} defaultDay={current.conductedOn} onSaved={onSaved} onCancel={() => setIsCorrecting(false)} />
      ) : null}
      {current && !isCorrecting ? (
        <View style={styles.actions}>
          <Button label="Corregir" icon="create-outline" variant="secondary" compact onPress={() => setIsCorrecting(true)} />
        </View>
      ) : null}
    </Card>
  );
}

type RecordFormProps = {
  eventId: number;
  teamId: number;
  defaultDay: string;
  onSaved: () => void;
};

function RecordForm({ eventId, teamId, defaultDay, onSaved }: RecordFormProps) {
  const form = useControlResultForm(defaultDay);
  const save = async () => {
    if (await form.record((input) => recordControlResult(eventId, teamId, input))) {
      onSaved();
    }
  };

  return (
    <>
      {form.formError ? <Notice tone="danger" message={form.formError} /> : null}
      <ResultFields values={form.values} onChange={form.setValues} errors={form.errors} />
      <View style={styles.actions}>
        <Button label="Registrar resultado" icon="save-outline" onPress={save} loading={form.isSubmitting} />
      </View>
    </>
  );
}

type CorrectFormProps = {
  resultId: number;
  defaultDay: string;
  onSaved: () => void;
  onCancel: () => void;
};

function CorrectForm({ resultId, defaultDay, onSaved, onCancel }: CorrectFormProps) {
  const form = useControlResultForm(defaultDay);
  const save = async () => {
    if (await form.correct((input) => correctControlResult(resultId, input))) {
      onSaved();
    }
  };

  return (
    <>
      {form.formError ? <Notice tone="danger" message={form.formError} /> : null}
      <ResultFields values={form.values} onChange={form.setValues} errors={form.errors} />
      <TextField
        label="Motivo de la corrección"
        value={form.correctionNote}
        onChangeText={form.setCorrectionNote}
        error={form.errors.correctionNote}
        multiline
      />
      <View style={styles.actions}>
        <Button label="Volver" variant="ghost" onPress={onCancel} disabled={form.isSubmitting} />
        <Button label="Guardar corrección" icon="save-outline" onPress={save} loading={form.isSubmitting} />
      </View>
    </>
  );
}

const styles = StyleSheet.create({
  actions: { flexDirection: 'row', justifyContent: 'flex-end', gap: spacing.md },
});
