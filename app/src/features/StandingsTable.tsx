import type { ReactNode } from 'react';
import { StyleSheet, View, useWindowDimensions } from 'react-native';

import { AppText, Avatar, Card } from '@/components';
import { teamById } from '@/mocks/catalog';
import type { StandingRow } from '@/mocks/data';
import { colors, spacing } from '@/theme';

type StandingsTableProps = {
  rows: StandingRow[];
  title?: string;
  action?: ReactNode;
};

const compactBreakpoint = 480;

export function StandingsTable({ rows, title = 'Clasificación de pilotos', action }: StandingsTableProps) {
  const isCompact = useWindowDimensions().width < compactBreakpoint;

  return (
    <Card title={title} action={action} flush>
      <View style={[styles.row, styles.headerRow]}>
        <AppText variant="overline" color="textMuted" style={styles.position}>
          Pos
        </AppText>
        <AppText variant="overline" color="textMuted" style={styles.driver}>
          Piloto
        </AppText>
        <AppText variant="overline" color="textMuted" style={isCompact ? styles.teamCompact : styles.team}>
          {isCompact ? '' : 'Equipo'}
        </AppText>
        <AppText variant="overline" color="textMuted" style={styles.points}>
          Pts
        </AppText>
      </View>
      {rows.map((row) => {
        const team = teamById(row.teamId);
        return (
          <View key={row.driver} style={styles.row}>
            <AppText variant="heading" color={row.position === 1 ? 'accent' : 'text'} style={styles.position}>
              {row.position}
            </AppText>
            <View style={[styles.driver, styles.driverCell]}>
              <Avatar name={row.driver} color={team.color} size={32} />
              <AppText variant="bodyStrong" numberOfLines={1} style={styles.flexText}>
                {row.driver}
              </AppText>
            </View>
            <View style={[isCompact ? styles.teamCompact : styles.team, styles.teamCell]} accessibilityLabel={team.name}>
              <View style={[styles.teamBar, { backgroundColor: team.color }]} />
              {isCompact ? null : (
                <AppText variant="caption" color="textMuted" numberOfLines={1} style={styles.flexText}>
                  {team.name}
                </AppText>
              )}
            </View>
            <AppText variant="heading" style={styles.points}>
              {row.points}
            </AppText>
          </View>
        );
      })}
    </Card>
  );
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
  headerRow: { backgroundColor: colors.surfaceRaised, paddingVertical: spacing.sm },
  position: { width: 32 },
  driver: { flex: 3 },
  driverCell: { flexDirection: 'row', alignItems: 'center', gap: spacing.sm },
  team: { flex: 2 },
  teamCompact: { width: 4 },
  teamCell: { flexDirection: 'row', alignItems: 'center', gap: spacing.sm },
  teamBar: { width: 4, height: 20, borderRadius: 2 },
  points: { width: 48, textAlign: 'right' },
  flexText: { flexShrink: 1 },
});
