import { StyleSheet, View } from 'react-native';

import type { PublicDriver } from '@/api/drivers';
import type { TeamSummary } from '@/api/teams';
import { ChipGroup, FieldGroup, SelectField } from '@/components';
import { categoryFilterOptions } from '@/features/filters';
import { DateRangeFilter } from '@/features/history/DateRangeFilter';
import type { DateRangeErrors } from '@/features/history/historyFilter';
import { spacing } from '@/theme';

import { anyChoice, type SanctionFilterValue } from './sanctionFilter';
import { statusOptions } from './sanctionText';

type SanctionFiltersProps = {
  value: SanctionFilterValue;
  onChange: (value: SanctionFilterValue) => void;
  errors: DateRangeErrors;
  drivers: readonly PublicDriver[];
  teams: readonly TeamSummary[];
};

const allCategories = 'all';

export function SanctionFilters({ value, onChange, errors, drivers, teams }: SanctionFiltersProps) {
  const driverOptions = [
    { value: anyChoice, label: 'Todos los pilotos' },
    ...drivers.map((driver) => ({ value: String(driver.id), label: `${driver.firstName} ${driver.lastName} · ${driver.teamName}` })),
  ];
  const teamOptions = [
    { value: anyChoice, label: 'Todas las escuderías' },
    ...teams.map((team) => ({ value: String(team.id), label: `${team.name} · ${team.categoryName}` })),
  ];

  return (
    <View style={styles.filters}>
      <ChipGroup
        options={categoryFilterOptions}
        selected={value.category || allCategories}
        onSelect={(category) => onChange({ ...value, category: category === allCategories ? anyChoice : category })}
        accessibilityLabel="Filtrar por categoría"
      />
      <ChipGroup options={statusOptions} selected={value.status} onSelect={(status) => onChange({ ...value, status })} accessibilityLabel="Filtrar por estado" />
      <View style={styles.row}>
        <View style={styles.field}>
          <FieldGroup label="Piloto">
            <SelectField options={driverOptions} selected={value.driverId} onSelect={(driverId) => onChange({ ...value, driverId })} accessibilityLabel="Filtrar por piloto" />
          </FieldGroup>
        </View>
        <View style={styles.field}>
          <FieldGroup label="Escudería">
            <SelectField options={teamOptions} selected={value.teamId} onSelect={(teamId) => onChange({ ...value, teamId })} accessibilityLabel="Filtrar por escudería" />
          </FieldGroup>
        </View>
      </View>
      <DateRangeFilter value={value} onChange={onChange} errors={errors} />
    </View>
  );
}

const styles = StyleSheet.create({
  filters: { gap: spacing.md },
  row: { flexDirection: 'row', flexWrap: 'wrap', gap: spacing.md },
  field: { flexGrow: 1, flexBasis: 220 },
});
