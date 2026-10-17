import type { Role } from './auth';
import { apiRequest, queryString } from './client';

export type MailboxBox = 'inbox' | 'sent';

export type ConversationTeam = {
  id: number;
  code: string;
  name: string;
};

export type Message = {
  id: number;
  body: string;
  sentAt: string;
  author: { username: string; role: Role };
  isOwn: boolean;
  isUnread: boolean;
  isReadByRecipient: boolean;
};

export type Conversation = {
  id: number;
  subject: string;
  team: ConversationTeam;
  otherTeamId: number | null;
  otherTeamName: string;
  createdAt: string;
  messages: Message[];
};

export type ConversationSummary = {
  id: number;
  subject: string;
  team: ConversationTeam;
  otherTeamId: number | null;
  otherTeamName: string;
  createdAt: string;
  lastMessage: Message;
  unreadCount: number;
};

export type Mailbox = {
  unreadCount: number;
  conversations: ConversationSummary[];
};

export type NewConversationInput = {
  teamId?: number;
  recipientTeamId?: number;
  subject: string;
  body: string;
};

export function listConversations(box: MailboxBox): Promise<Mailbox> {
  return apiRequest<Mailbox>('GET', `/conversations${queryString({ box })}`);
}

export function startConversation(input: NewConversationInput): Promise<Conversation> {
  return apiRequest<Conversation>('POST', '/conversations', input);
}

// Opening a conversation marks it read: its messages come flagged as they were
// before this read.
export function openConversation(id: number): Promise<Conversation> {
  return apiRequest<Conversation>('GET', `/conversations/${id}`);
}

export function replyToConversation(id: number, body: string): Promise<Message> {
  return apiRequest<Message>('POST', `/conversations/${id}/messages`, { body });
}
