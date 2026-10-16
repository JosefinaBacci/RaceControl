import { router } from 'expo-router';

import { ConversationListScreen } from '@/features/messages/ConversationListScreen';

export default function TeamMessagesScreen() {
  return (
    <ConversationListScreen
      subtitle="Tus conversaciones con la FIA. Solo las ven los administradores de tu escudería y de la FIA."
      showsTeam={false}
      onOpen={(id) => router.push({ pathname: '/team/messages/[id]', params: { id: String(id) } })}
      onStart={() => router.push('/team/messages/new')}
    />
  );
}
