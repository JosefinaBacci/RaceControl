import { router } from 'expo-router';
import { useState } from 'react';
import { StyleSheet, View } from 'react-native';

import { useSession } from '@/auth/SessionProvider';
import type { Conversation, Message } from '@/api/messages';
import { AppText, Badge, Button, Card, Notice, Screen, TextField } from '@/components';
import { useConversation } from '@/data/messages';
import { useFormSubmission } from '@/data/useFormSubmission';
import { RemoteContent } from '@/features/RemoteContent';
import { formatDateTime } from '@/format';
import { colors, radius, spacing } from '@/theme';

import { maxBodyLength } from './limits';

const fields = ['body'] as const;

export function ConversationScreen({ id }: { id: number }) {
  const { state, reload, reply } = useConversation(id);
  const { user } = useSession();
  const mailbox = user?.role === 'fia_admin' ? '/fia/messages' : '/team/messages';

  return (
    <Screen title="Conversación" headerAction={
      <View style={styles.actions}>
        <Button label="Bandeja" icon="arrow-back" variant="ghost" compact onPress={() => router.replace(mailbox)} />
        <Button label="Actualizar" icon="refresh" variant="ghost" compact onPress={() => void reload()} />
      </View>
    }>
      <RemoteContent state={state} onRetry={reload}>
        {(conversation) => <Thread conversation={conversation} onReply={reply} />}
      </RemoteContent>
    </Screen>
  );
}

function Thread({ conversation, onReply }: { conversation: Conversation; onReply: (body: string) => Promise<void> }) {
  return (
    <>
      <View style={styles.header}>
        <AppText variant="title" accessibilityRole="header">
          {conversation.subject}
        </AppText>
        <AppText color="textMuted">Conversación entre {conversation.team.name} y la FIA</AppText>
      </View>
      <View style={styles.messages}>
        {conversation.messages.map((message) => (
          <MessageBubble key={message.id} message={message} />
        ))}
      </View>
      <Composer onSend={onReply} />
    </>
  );
}

// The body is shown as plain text, never as markup, whatever it contains.
function MessageBubble({ message }: { message: Message }) {
  const side = message.author.role === 'fia_admin' ? 'FIA' : 'Escudería';

  return (
    <View style={[styles.bubble, message.isOwn ? styles.ownBubble : styles.theirBubble, message.isUnread && styles.unreadBubble]}>
      <View style={styles.bubbleHeader}>
        <AppText variant="caption" color="textMuted">
          {message.isOwn ? 'Vos' : `${message.author.username} · ${side}`} · {formatDateTime(message.sentAt)}
        </AppText>
        {message.isUnread ? <Badge label="Nuevo" tone="accent" /> : null}
      </View>
      <AppText selectable>{message.body}</AppText>
    </View>
  );
}

function Composer({ onSend }: { onSend: (body: string) => Promise<void> }) {
  const [body, setBody] = useState('');
  const { fieldErrors, formError, isSubmitting, submit } = useFormSubmission(fields);

  const send = async () => {
    const sent = await submit(async () => {
      await onSend(body);
      return true;
    });
    if (sent) {
      setBody('');
    }
  };

  return (
    <Card>
      {formError ? <Notice tone="danger" message={formError} /> : null}
      <TextField
        label="Escribí tu respuesta…"
        icon="chatbubble-outline"
        value={body}
        onChangeText={setBody}
        error={fieldErrors.body}
        multiline
        maxLength={maxBodyLength}
        style={styles.composerInput}
      />
      <View style={styles.actions}>
        <Button label="Enviar" icon="send" onPress={send} loading={isSubmitting} disabled={body.trim() === ''} />
      </View>
    </Card>
  );
}

const styles = StyleSheet.create({
  header: { gap: spacing.xs },
  messages: { gap: spacing.md },
  bubble: { maxWidth: '85%', gap: spacing.xs, padding: spacing.md, borderRadius: radius.md, borderWidth: 1, borderColor: colors.border },
  ownBubble: { alignSelf: 'flex-end', backgroundColor: colors.surfaceRaised },
  theirBubble: { alignSelf: 'flex-start', backgroundColor: colors.surface },
  unreadBubble: { borderColor: colors.accent },
  bubbleHeader: { flexDirection: 'row', alignItems: 'center', gap: spacing.sm, flexWrap: 'wrap' },
  composerInput: { minHeight: 80, paddingVertical: spacing.md, textAlignVertical: 'top' },
  actions: { flexDirection: 'row', justifyContent: 'flex-end' },
});
