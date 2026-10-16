import type { ReactNode } from 'react';
import { StyleSheet, View, useWindowDimensions } from 'react-native';

import { AppText, Avatar, Card } from '@/components';
import { formatPoints } from '@/format';
import { colors, spacing } from '@/theme';

// A ranked line: a driver of a race or a championship, or a team. The detail is
// the secondary column, such as the driver's team; a team row leaves it out.
export type StandingsRow = {
  key: string;
  position: number;
  name: string;
  color: string;
  detail?: string;
  points: number;
};

type StandingsTableProps = {
  rows: StandingsRow[];
  title?: string;
  action?: ReactNode;
  nameLabel?: string;
  detailLabel?: string;
  emptyMessage?: string;
};

const compactBreakpoint = 480;

export function StandingsTable({
  rows,
  title = 'Clasificación de pilotos',
  action,
  nameLabel = 'Piloto',
  detailLabel = 'Equipo',
  emptyMessage = 'Todavía no hay puntajes cargados.',
}: StandingsTableProps) {
  const isCompact = useWindowDimensions().width < compactBreakpoint;

  return (
    <Card title={title} action={action} flush>
      <View style={[styles.row, styles.headerRow]}>
        <AppText variant="overline" color="textMuted" style={styles.position}>
          Pos
        </AppText>
        <AppText variant="overline" color="textMuted" style={styles.name}>
          {nameLabel}
        </AppText>
        <AppText variant="overline" color="textMuted" style={isCompact ? styles.detailCompact : styles.detail}>
          {isCompact ? '' : detailLabel}
        </AppText>
        <AppText variant="overline" color="textMuted" style={styles.points}>
          Pts
        </AppText>
      </View>
      {rows.length === 0 ? (
        <AppText color="textMuted" style={styles.empty}>
          {emptyMessage}
        </AppText>
      ) : null}
      {rows.map((row) => (
        <View key={row.key} style={styles.row}>
          <AppText variant="heading" color={row.position === 1 ? 'accent' : 'text'} style={styles.position}>
            {row.position}
          </AppText>
          <View style={[styles.name, styles.nameCell]}>
            <Avatar name={row.name} color={row.color} size={32} />
            <AppText variant="bodyStrong" numberOfLines={1} style={styles.flexText}>
              {row.name}
            </AppText>
          </View>
          <View style={[isCompact ? styles.detailCompact : styles.detail, styles.detailCell]} accessibilityLabel={row.detail}>
            <View style={[styles.colorBar, { backgroundColor: row.color }]} />
            {isCompact || !row.detail ? null : (
              <AppText variant="caption" color="textMuted" numberOfLines={1} style={styles.flexText}>
                {row.detail}
              </AppText>
            )}
          </View>
          <AppText variant="heading" style={styles.points}>
            {formatPoints(row.points)}
          </AppText>
        </View>
      ))}
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
  empty: { paddingHorizontal: spacing.lg, paddingVertical: spacing.md },
  position: { width: 32 },
  name: { flex: 3 },
  nameCell: { flexDirection: 'row', alignItems: 'center', gap: spacing.sm },
  detail: { flex: 2 },
  detailCompact: { width: 4 },
  detailCell: { flexDirection: 'row', alignItems: 'center', gap: spacing.sm },
  colorBar: { width: 4, height: 20, borderRadius: 2 },
  points: { width: 56, textAlign: 'right' },
  flexText: { flexShrink: 1 },
});
