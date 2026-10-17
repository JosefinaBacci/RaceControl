import { SelectField, FieldGroup } from '@/components';
import { useTeams } from '@/data/accounts';
import { RemoteContent } from '@/features/RemoteContent';

type TeamPickerProps = {
  teamId: number | null;
  onChange: (teamId: number) => void;
  error?: string;
  excludeTeamId?: number | null;
};

export function TeamPicker({ teamId, onChange, error, excludeTeamId }: TeamPickerProps) {
  const { state, reload } = useTeams();

  return (
    <FieldGroup label="Escudería" error={error}>
      <RemoteContent state={state} onRetry={reload}>
        {(teams) => (
          <SelectField
            options={teams.filter((team) => team.id !== excludeTeamId).map((team) => ({ value: String(team.id), label: `${team.name} · ${team.categoryName}` }))}
            selected={teamId === null ? '' : String(teamId)}
            onSelect={(selected) => onChange(Number(selected))}
            accessibilityLabel="Elegir escudería"
          />
        )}
      </RemoteContent>
    </FieldGroup>
  );
}
