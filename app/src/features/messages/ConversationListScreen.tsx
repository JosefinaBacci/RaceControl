import { useState } from 'react';
import { Pressable, StyleSheet, View } from 'react-native';

import type { ConversationSummary, MailboxBox } from '@/api/messages';
import { AppText, Badge, Button, Card, ChipGroup, Screen } from '@/components';
import { useMailbox } from '@/data/messages';
import { RemoteContent } from '@/features/RemoteContent';
import { formatDateTime } from '@/format';
import { colors, spacing } from '@/theme';

const boxOptions = [
  { value: 'inbox' as const, label: 'Entrada' },
  { value: 'sent' as const, label: 'Enviados' },
];

const emptyBoxMessage: Record<MailboxBox, string> = {
  inbox: 'Todavía no hay conversaciones.',
  sent: 'Todavía no escribiste en ninguna conversación.',
};

type ConversationListScreenProps = {
  subtitle: string;
  showsTeam: boolean;
  onOpen: (id: number) => void;
  onStart: () => void;
};

export function ConversationListScreen({ subtitle, showsTeam, onOpen, onStart }: ConversationListScreenProps) {
  const [box, setBox] = useState<MailboxBox>('inbox');
  const { state, reload } = useMailbox(box);

  return (
    <Screen
      title="Mensajes"
      subtitle={subtitle}
      headerAction={<Button label="Nueva conversación" icon="create-outline" compact onPress={onStart} />}
    >
      <ChipGroup variant="underline" options={boxOptions} selected={box} onSelect={setBox} accessibilityLabel="Elegir bandeja" />
      <Card flush>
        <RemoteContent state={state} onRetry={reload}>
          {(mailbox) =>
            mailbox.conversations.length === 0 ? (
              <AppText color="textMuted" style={styles.empty}>
                {emptyBoxMessage[box]}
              </AppText>
            ) : (
              mailbox.conversations.map((conversation) => (
                <ConversationRow key={conversation.id} conversation={conversation} showsTeam={showsTeam} onPress={() => onOpen(conversation.id)} />
              ))
            )
          }
        </RemoteContent>
      </Card>
    </Screen>
  );
}

function ConversationRow({ conversation, showsTeam, onPress }: { conversation: ConversationSummary; showsTeam: boolean; onPress: () => void }) {
  const { lastMessage, unreadCount } = conversation;
  const hasUnread = unreadCount > 0;
  const author = lastMessage.isOwn ? 'Vos' : lastMessage.author.username;

  return (
    <Pressable
      accessibilityRole="button"
      accessibilityLabel={hasUnread ? `${conversation.subject}, ${unreadCount} sin leer` : conversation.subject}
      onPress={onPress}
      style={({ pressed }) => [styles.row, pressed && styles.pressed]}
    >
      <View style={[styles.unreadDot, hasUnread && styles.unreadDotVisible]} />
      <View style={styles.text}>
        <AppText variant={hasUnread ? 'bodyStrong' : 'body'} numberOfLines={1}>
          {conversation.subject}
        </AppText>
        <AppText variant="caption" color="textMuted" numberOfLines={2}>
          {showsTeam ? `${conversation.team.name} · ` : ''}
          {author}: {lastMessage.body}
        </AppText>
      </View>
      <View style={styles.trailing}>
        <AppText variant="caption" color="textMuted">
          {formatDateTime(lastMessage.sentAt)}
        </AppText>
        {hasUnread ? <Badge label={`${unreadCount} sin leer`} tone="accent" /> : null}
      </View>
    </Pressable>
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
  pressed: { opacity: 0.7 },
  unreadDot: { width: 8, height: 8, borderRadius: 4 },
  unreadDotVisible: { backgroundColor: colors.accent },
  text: { flex: 1, gap: 2 },
  trailing: { alignItems: 'flex-end', gap: spacing.xs },
  empty: { padding: spacing.lg },
});
