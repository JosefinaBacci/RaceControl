import { StyleSheet, View } from 'react-native';

import type { Sanction, SanctionDetail as SanctionDetailData, SanctionVersion } from '@/api/sanctions';
import { AppText, Badge, Card, FieldGroup, ListItem, Notice } from '@/components';
import { formatInstantDay } from '@/features/history/day';
import { formatDateTime } from '@/format';
import { spacing } from '@/theme';

import { actionLabels, penaltyText, subjectText } from './sanctionText';

// What every viewer sees of a sanction: who it falls on, why, what it imposes
// and when the FIA issued it. A revocation stays on the record with its note.
export function SanctionSummaryCard({ sanction }: { sanction: Sanction }) {
  const { current } = sanction;
  const action = actionLabels[current.action];

  return (
    <Card title={subjectText(sanction.subject, sanction.team)} action={<Badge label={action.label} tone={action.tone} />}>
      {sanction.status === 'revoked' ? (
        <Notice tone="info" message={`Revocada el ${formatDateTime(current.changedAt)}: ${current.note ?? ''}`} />
      ) : null}
      <FieldGroup label="Penalización">
        <AppText variant="bodyStrong">{penaltyText(current.penalty)}</AppText>
      </FieldGroup>
      <FieldGroup label="Motivo">
        <AppText>{current.reason}</AppText>
      </FieldGroup>
      <View style={styles.facts}>
        <Fact label="Destinatario" value={sanction.subject.kind === 'team' ? 'Escudería' : 'Piloto'} />
        <Fact label="Escudería" value={sanction.team.name} />
        <Fact label="Campeonato" value={sanction.categoryName} />
        <Fact label="Evento" value={current.event?.name ?? 'Sin evento asociado'} />
        <Fact label="Fecha" value={formatInstantDay(sanction.issuedAt)} />
      </View>
    </Card>
  );
}

function Fact({ label, value }: { label: string; value: string }) {
  return (
    <View style={styles.fact}>
      <FieldGroup label={label}>
        <AppText>{value}</AppText>
      </FieldGroup>
    </View>
  );
}

export function SanctionHistoryCard({ sanction }: { sanction: SanctionDetailData }) {
  return (
    <Card title="Historial de la sanción">
      {sanction.versions.map((version, index) => (
        <ListItem
          key={version.id}
          leading={<Badge label={actionLabels[version.action].label} tone={actionLabels[version.action].tone} />}
          title={penaltyText(version.penalty)}
          subtitle={versionText(version)}
          isLast={index === sanction.versions.length - 1}
        />
      ))}
    </Card>
  );
}

function versionText(version: SanctionVersion): string {
  const when = `${actionLabels[version.action].verb} por la FIA el ${formatDateTime(version.changedAt)}`;
  return version.note ? `${when} · ${version.note}` : when;
}

const styles = StyleSheet.create({
  facts: { flexDirection: 'row', flexWrap: 'wrap', gap: spacing.md },
  fact: { flexGrow: 1, flexBasis: 160 },
});
