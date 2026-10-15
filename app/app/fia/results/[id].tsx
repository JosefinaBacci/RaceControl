import { router, useLocalSearchParams } from 'expo-router';
import { useState } from 'react';
import { StyleSheet, View } from 'react-native';

import { replaceScores, type ScoreEntryInput, type ScoreGrid, type ScoreGridEntry } from '@/api/scores';
import { AppText, Button, Card, Notice, Screen, TextField } from '@/components';
import { useScoreGrid } from '@/data/scores';
import { useFormSubmission } from '@/data/useFormSubmission';
import { RemoteContent } from '@/features/RemoteContent';
import { Flag } from '@/illustrations';
import { spacing } from '@/theme';

const fields = ['entries'] as const;
const invalidPointsMessage = 'Cada puntaje tiene que ser un número mayor o igual a cero.';

export default function ScoreEditorScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const { state, reload, setData } = useScoreGrid(Number(id));

  return (
    <Screen
      title="Cargar puntajes"
      headerAction={<Button label="Volver" icon="arrow-back" variant="ghost" compact onPress={() => router.back()} />}
    >
      <RemoteContent state={state} onRetry={reload}>
        {(grid) => <ScoreForm grid={grid} onSaved={setData} />}
      </RemoteContent>
    </Screen>
  );
}

function ScoreForm({ grid, onSaved }: { grid: ScoreGrid; onSaved: (grid: ScoreGrid) => void }) {
  const [points, setPoints] = useState<Record<number, string>>(() => initialPoints(grid));
  const [savedMessage, setSavedMessage] = useState<string | null>(null);
  const { fieldErrors, formError, isSubmitting, submit, setFormError } = useFormSubmission(fields);
  const entries = parseEntries(grid, points);
  const error = formError ?? fieldErrors.entries;

  const save = async () => {
    setSavedMessage(null);
    if (entries === null) {
      setFormError(invalidPointsMessage);
      return;
    }
    const updated = await submit(() => replaceScores(grid.eventId, entries));
    if (updated) {
      setPoints(initialPoints(updated));
      setSavedMessage('Puntajes guardados.');
      onSaved(updated);
    }
  };

  return (
    <Card title={grid.categoryName}>
      {error ? <Notice tone="danger" message={error} /> : null}
      {savedMessage ? <Notice tone="success" message={savedMessage} /> : null}
      <AppText variant="caption" color="textMuted">
        Dejá vacío al piloto que no sumó puntos: se le quita el puntaje.
      </AppText>
      {grid.entries.map((entry) => (
        <ScoreRow
          key={entry.driverId}
          entry={entry}
          value={points[entry.driverId] ?? ''}
          onChange={(value) => setPoints((previous) => ({ ...previous, [entry.driverId]: value }))}
        />
      ))}
      <View style={styles.actions}>
        <Button label="Guardar puntajes" icon="save-outline" onPress={save} loading={isSubmitting} disabled={entries === null} />
      </View>
    </Card>
  );
}

function ScoreRow({ entry, value, onChange }: { entry: ScoreGridEntry; value: string; onChange: (value: string) => void }) {
  return (
    <View style={styles.row}>
      <Flag code={entry.countryCode} width={24} />
      <View style={styles.name}>
        <AppText variant="bodyStrong">
          {entry.firstName} {entry.lastName}
        </AppText>
        <AppText variant="caption" color="textMuted">
          {entry.teamName}
          {entry.number === null ? '' : ` · N.º ${entry.number}`}
        </AppText>
      </View>
      <View style={styles.pointsField}>
        <TextField label="Puntos" value={value} onChangeText={onChange} keyboardType="decimal-pad" />
      </View>
    </View>
  );
}

// A blank field is a driver left out; a non-numeric or negative value makes the
// whole grid invalid (null) so the save never reaches the API.
function parseEntries(grid: ScoreGrid, points: Record<number, string>): ScoreEntryInput[] | null {
  const entries: ScoreEntryInput[] = [];
  for (const entry of grid.entries) {
    const raw = (points[entry.driverId] ?? '').trim();
    if (raw === '') {
      continue;
    }
    const value = Number(raw);
    if (!Number.isFinite(value) || value < 0) {
      return null;
    }
    entries.push({ driverId: entry.driverId, points: value });
  }
  return entries;
}

function initialPoints(grid: ScoreGrid): Record<number, string> {
  const points: Record<number, string> = {};
  for (const entry of grid.entries) {
    points[entry.driverId] = entry.points === null ? '' : String(entry.points);
  }
  return points;
}

const styles = StyleSheet.create({
  row: { flexDirection: 'row', alignItems: 'center', gap: spacing.md },
  name: { flex: 1 },
  pointsField: { width: 110 },
  actions: { flexDirection: 'row', justifyContent: 'flex-end', gap: spacing.md },
});
