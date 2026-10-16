import { router } from 'expo-router';

import { NewConversationScreen } from '@/features/messages/NewConversationScreen';

export default function NewFiaConversationScreen() {
  return (
    <NewConversationScreen
      subtitle="Elegí la escudería: la conversación la ven todos sus administradores."
      picksTeam
      onStarted={(conversation) => router.replace({ pathname: '/fia/messages/[id]', params: { id: String(conversation.id) } })}
    />
  );
}
