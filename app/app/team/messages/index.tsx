import { router } from 'expo-router';

import { ConversationListScreen } from '@/features/messages/ConversationListScreen';

export default function TeamMessagesScreen() {
  return (
    <ConversationListScreen
      subtitle="Conversaciones privadas con la FIA y otras escuderías. Solo acceden las partes participantes."
      showsTeam={false}
      onOpen={(id) => router.push({ pathname: '/team/messages/[id]', params: { id: String(id) } })}
      onStart={() => router.push('/team/messages/new')}
    />
  );
}
