import { router, useLocalSearchParams } from 'expo-router';
import { useState } from 'react';
import { StyleSheet, View } from 'react-native';

import { replaceScores, type ScoreGrid, type ScoreGridEntry } from '@/api/scores';
import { AppText, Button, Card, IconButton, Notice, Screen, TextField } from '@/components';
import { useScoreGrid } from '@/data/scores';
import { useFormSubmission } from '@/data/useFormSubmission';
import { RemoteContent } from '@/features/RemoteContent';
import { initialSheet, lineError, sheetEntries, type ScoreLine, type ScoreSheet } from '@/features/scores/scoreSheet';
import { Flag } from '@/illustrations';
import { spacing } from '@/theme';

const fields = ['entries'] as const;
const incompleteSheetMessage = 'Hay pilotos sin puntos: completalos o marcá que no corrieron.';

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
  const [sheet, setSheet] = useState<ScoreSheet>(() => initialSheet(grid));
  const [hasTriedToSave, setHasTriedToSave] = useState(false);
  const [savedMessage, setSavedMessage] = useState<string | null>(null);
  const { fieldErrors, formError, isSubmitting, submit, setFormError } = useFormSubmission(fields);
  const error = formError ?? fieldErrors.entries;
  const setLine = (driverId: number, line: ScoreLine) => setSheet((previous) => ({ ...previous, [driverId]: line }));

  const save = async () => {
    setSavedMessage(null);
    setHasTriedToSave(true);
    const entries = sheetEntries(grid, sheet);
    if (entries === null) {
      setFormError(incompleteSheetMessage);
      return;
    }
    const updated = await submit(() => replaceScores(grid.eventId, entries));
    if (updated) {
      setSheet(initialSheet(updated));
      setHasTriedToSave(false);
      setSavedMessage('Puntajes guardados.');
      onSaved(updated);
    }
  };

  return (
    <Card title={grid.categoryName}>
      {error ? <Notice tone="danger" message={error} /> : null}
      {savedMessage ? <Notice tone="success" message={savedMessage} /> : null}
      <AppText variant="caption" color="textMuted">
        Cargá 0 al piloto que corrió sin sumar, y marcá que no corrió al que no participó.
      </AppText>
      {grid.entries.map((entry) => (
        <ScoreRow
          key={entry.driverId}
          entry={entry}
          line={sheet[entry.driverId] ?? { didRace: false }}
          showError={hasTriedToSave}
          onChange={(line) => setLine(entry.driverId, line)}
        />
      ))}
      <View style={styles.actions}>
        <Button label="Guardar puntajes" icon="save-outline" onPress={save} loading={isSubmitting} />
      </View>
    </Card>
  );
}

type ScoreRowProps = {
  entry: ScoreGridEntry;
  line: ScoreLine;
  showError: boolean;
  onChange: (line: ScoreLine) => void;
};

function ScoreRow({ entry, line, showError, onChange }: ScoreRowProps) {
  const name = `${entry.firstName} ${entry.lastName}`;
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
      {line.didRace ? (
        <>
          <View style={styles.pointsField}>
            <TextField
              label="Puntos"
              value={line.points}
              onChangeText={(points) => onChange({ didRace: true, points })}
              keyboardType="decimal-pad"
              error={showError ? lineError(line) : null}
            />
          </View>
          <IconButton
            icon="close"
            accessibilityLabel={`Marcar que ${name} no corrió`}
            onPress={() => onChange({ didRace: false })}
          />
        </>
      ) : (
        <>
          <AppText color="textMuted" style={styles.pointsField}>
            No corrió
          </AppText>
          <IconButton
            icon="add"
            accessibilityLabel={`Cargar los puntos de ${name}`}
            onPress={() => onChange({ didRace: true, points: '' })}
          />
        </>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  row: { flexDirection: 'row', alignItems: 'center', gap: spacing.md },
  name: { flex: 1 },
  pointsField: { width: 110 },
  actions: { flexDirection: 'row', justifyContent: 'flex-end', gap: spacing.md },
});
