import { router } from 'expo-router';

import { NewConversationScreen } from '@/features/messages/NewConversationScreen';

export default function NewTeamConversationScreen() {
  return (
    <NewConversationScreen
      subtitle="Escribile a la FIA en nombre de tu escudería."
      picksTeam={false}
      onStarted={(conversation) => router.replace({ pathname: '/team/messages/[id]', params: { id: String(conversation.id) } })}
    />
  );
}
