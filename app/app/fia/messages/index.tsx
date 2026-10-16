import { router } from 'expo-router';

import { ConversationListScreen } from '@/features/messages/ConversationListScreen';

export default function FiaMessagesScreen() {
  return (
    <ConversationListScreen
      subtitle="Las conversaciones de la FIA con cada escudería."
      showsTeam
      onOpen={(id) => router.push({ pathname: '/fia/messages/[id]', params: { id: String(id) } })}
      onStart={() => router.push('/fia/messages/new')}
    />
  );
}
