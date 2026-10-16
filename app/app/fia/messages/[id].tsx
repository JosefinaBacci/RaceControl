import { useLocalSearchParams } from 'expo-router';

import { ConversationScreen } from '@/features/messages/ConversationScreen';

export default function ConversationRoute() {
  const { id } = useLocalSearchParams<{ id: string }>();
  return <ConversationScreen id={Number(id)} />;
}
