import { View } from 'react-native';

import type { SanctionFollowUp, SanctionVersionFollowUp } from '@/api/sanctions';
import { Avatar, Badge, Card, CardLink, ListItem } from '@/components';
import { AcknowledgementBadge } from '@/features/AcknowledgementBadge';
import { acknowledgementText } from '@/features/scores/scoreNotice';
import { formatDateTime } from '@/format';
import { teamLogo } from '@/logos';
import { teamById } from '@/mocks/catalog';

import { actionLabels, penaltyText, subjectText } from './sanctionText';

type SanctionFollowUpCardProps = {
  followUp: SanctionFollowUp;
  onOpen?: () => void;
};

// A sanction as the FIA follows it: each version, who wrote it, and which teams
// of the championship confirmed receiving it.
export function SanctionFollowUpCard({ followUp, onOpen }: SanctionFollowUpCardProps) {
  const { sanction } = followUp;

  return (
    <Card
      title={`${subjectText(sanction.subject, sanction.team)} · ${penaltyText(sanction.current.penalty)}`}
      action={onOpen ? <CardLink label="Ver sanción" onPress={onOpen} /> : undefined}
    >
      {followUp.versions.map((version) => (
        <VersionFollowUp key={version.id} version={version} />
      ))}
    </Card>
  );
}

function VersionFollowUp({ version }: { version: SanctionVersionFollowUp }) {
  const action = actionLabels[version.action];
  const replaced = version.isCurrent ? '' : ' · reemplazada por una versión posterior';

  return (
    <View>
      <ListItem
        leading={<Badge label={action.label} tone={action.tone} />}
        title={`${action.verb} por ${version.changedBy} el ${formatDateTime(version.changedAt)}`}
        subtitle={`${version.confirmedCount} de ${version.teams.length} escuderías confirmaron${replaced}`}
      />
      {version.teams.map((recipient, index) => (
        <ListItem
          key={recipient.team.id}
          leading={<Avatar name={recipient.team.name} color={teamById(recipient.team.id).color} image={teamLogo(recipient.team.id)} />}
          title={recipient.team.name}
          subtitle={acknowledgementText(recipient)}
          trailing={<AcknowledgementBadge acknowledged={recipient.status === 'confirmed'} acknowledgedLabel="Confirmada" pendingLabel="Pendiente" />}
          isLast={index === version.teams.length - 1}
        />
      ))}
    </View>
  );
}
