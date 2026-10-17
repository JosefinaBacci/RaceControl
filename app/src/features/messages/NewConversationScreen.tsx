import { useSession } from '@/auth/SessionProvider';
import { router } from 'expo-router';
import { useState } from 'react';
import { StyleSheet, View } from 'react-native';

import { startConversation, type Conversation } from '@/api/messages';
import { ChipGroup, FieldGroup, Button, Card, Notice, Screen, TextField } from '@/components';
import { useFormSubmission } from '@/data/useFormSubmission';
import { TeamPicker } from '@/features/TeamPicker';
import { spacing } from '@/theme';

import { maxBodyLength, maxSubjectLength } from './limits';

const fields = ['teamId', 'recipientTeamId', 'subject', 'body'] as const;

type NewConversationScreenProps = {
  subtitle: string;
  picksTeam: boolean;
  onStarted: (conversation: Conversation) => void;
};

export function NewConversationScreen({ subtitle, picksTeam, onStarted }: NewConversationScreenProps) {
  const ownTeamId = useSession().user?.teamId;
  const [recipient, setRecipient] = useState<'fia' | 'team'>('fia');
  const [teamId, setTeamId] = useState<number | null>(null);
  const [subject, setSubject] = useState('');
  const [body, setBody] = useState('');
  const { fieldErrors, formError, isSubmitting, submit } = useFormSubmission(fields);

  const send = async () => {
    const started = await submit(() => startConversation({ teamId: picksTeam ? teamId ?? undefined : undefined, recipientTeamId: !picksTeam && recipient === 'team' ? teamId ?? 0 : undefined, subject, body }));
    if (started) {
      onStarted(started);
    }
  };

  return (
    <Screen title="Nueva conversación" subtitle={subtitle}>
      <Card>
        {formError ? <Notice tone="danger" message={formError} /> : null}
        {!picksTeam ? <FieldGroup label="Destinatario">
          <ChipGroup options={[{ value: 'fia', label: 'FIA' }, { value: 'team', label: 'Otra escudería' }]} selected={recipient}
            onSelect={(value) => { setRecipient(value); setTeamId(null); }} accessibilityLabel="Tipo de destinatario" />
        </FieldGroup> : null}
        {picksTeam || recipient === 'team' ? <TeamPicker teamId={teamId} onChange={setTeamId}
          excludeTeamId={picksTeam ? null : ownTeamId} error={fieldErrors.teamId ?? fieldErrors.recipientTeamId} /> : null}
        <TextField
          label="Asunto"
          icon="pricetag-outline"
          value={subject}
          onChangeText={setSubject}
          error={fieldErrors.subject}
          maxLength={maxSubjectLength}
        />
        <TextField
          label="Mensaje"
          icon="chatbubble-outline"
          value={body}
          onChangeText={setBody}
          error={fieldErrors.body}
          multiline
          maxLength={maxBodyLength}
          style={styles.bodyInput}
        />
      </Card>
      <View style={styles.actions}>
        <Button label="Cancelar" variant="ghost" onPress={() => router.back()} disabled={isSubmitting} />
        <Button label="Enviar" icon="send" onPress={send} loading={isSubmitting} disabled={!subject.trim() || !body.trim() || ((picksTeam || recipient === 'team') && teamId === null)} />
      </View>
    </Screen>
  );
}

const styles = StyleSheet.create({
  bodyInput: { minHeight: 120, paddingVertical: spacing.md, textAlignVertical: 'top' },
  actions: { flexDirection: 'row', justifyContent: 'flex-end', gap: spacing.md },
});
