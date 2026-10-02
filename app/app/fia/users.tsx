import Ionicons from '@expo/vector-icons/Ionicons';
import { useState } from 'react';
import { StyleSheet, View, useWindowDimensions } from 'react-native';

import { AppText, Avatar, Badge, Button, Card, ChipGroup, IconButton, Screen, TextField } from '@/components';
import { roleLabel } from '@/auth/roles';
import { useManagedUsers, type RoleFilter } from '@/data/hooks';
import { MockNotice } from '@/features/MockNotice';
import { teamById } from '@/mocks/catalog';
import type { ManagedUser } from '@/mocks/data';
import { colors, spacing } from '@/theme';

const roleFilterOptions = [
  { value: 'all' as const, label: 'Todos' },
  { value: 'fia_admin' as const, label: 'Administradores' },
  { value: 'team_admin' as const, label: 'Escuderías' },
];

const tableBreakpoint = 720;

export default function FiaUsersScreen() {
  const [search, setSearch] = useState('');
  const [role, setRole] = useState<RoleFilter>('all');
  const users = useManagedUsers(search, role);
  const showsTable = useWindowDimensions().width >= tableBreakpoint;

  return (
    <Screen
      title="Gestión de usuarios"
      subtitle="Creá, modificá o desactivá cuentas y asigná su rol."
      headerAction={<Button label="Nuevo usuario" icon="add" compact disabled />}
    >
      <MockNotice />
      <ChipGroup variant="underline" options={roleFilterOptions} selected={role} onSelect={setRole} accessibilityLabel="Filtrar por rol" />
      <TextField
        label="Buscar usuario o email…"
        icon="search-outline"
        value={search}
        onChangeText={setSearch}
        autoCapitalize="none"
        autoCorrect={false}
      />
      <Card flush>
        {showsTable ? <TableHeader /> : null}
        {users.length === 0 ? (
          <AppText color="textMuted" style={styles.empty}>
            No hay usuarios que coincidan con la búsqueda.
          </AppText>
        ) : (
          users.map((user) => <UserRow key={user.id} user={user} showsTable={showsTable} />)
        )}
      </Card>
    </Screen>
  );
}

function TableHeader() {
  return (
    <View style={[styles.row, styles.headerRow]}>
      {['Usuario', 'Rol', 'Estado', 'Acciones'].map((label, index) => (
        <AppText key={label} variant="overline" color="textMuted" style={columnStyleList[index]}>
          {label}
        </AppText>
      ))}
    </View>
  );
}

function UserRow({ user, showsTable }: { user: ManagedUser; showsTable: boolean }) {
  const team = teamById(user.teamId);
  const roleText = user.role === 'team_admin' ? `${roleLabel[user.role]} · ${team.name}` : roleLabel[user.role];
  const status = (
    <Badge
      label={user.isActive ? 'Activo' : 'Desactivado'}
      tone={user.isActive ? 'success' : 'neutral'}
      icon={user.isActive ? 'checkmark' : 'remove'}
    />
  );
  const actions = (
    <View style={styles.actions}>
      <IconButton icon="create-outline" accessibilityLabel={`Editar ${user.username}`} disabled />
      <IconButton icon="trash-outline" tone="danger" accessibilityLabel={`Desactivar ${user.username}`} disabled />
    </View>
  );

  return (
    <View style={styles.row}>
      <View style={[columnStyleList[0], styles.userCell]}>
        <Avatar name={user.username} color={user.role === 'team_admin' ? team.color : colors.accent} />
        <View style={styles.userText}>
          <AppText variant="bodyStrong" numberOfLines={1}>
            {user.username}
          </AppText>
          <AppText variant="caption" color="textMuted" numberOfLines={1}>
            {showsTable ? user.email : roleText}
          </AppText>
        </View>
      </View>
      {showsTable ? (
        <>
          <View style={[columnStyleList[1], styles.roleCell]}>
            <Ionicons name={user.role === 'fia_admin' ? 'shield-checkmark-outline' : 'car-sport-outline'} size={16} color={colors.textMuted} />
            <AppText variant="caption" color="textMuted" numberOfLines={1} style={styles.userText}>
              {roleText}
            </AppText>
          </View>
          <View style={columnStyleList[2]}>{status}</View>
          <View style={columnStyleList[3]}>{actions}</View>
        </>
      ) : (
        <View style={styles.compactTrailing}>
          {status}
          {actions}
        </View>
      )}
    </View>
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
  headerRow: { backgroundColor: colors.surfaceRaised, borderTopWidth: 0, paddingVertical: spacing.sm },
  userCell: { flexDirection: 'row', alignItems: 'center', gap: spacing.md },
  userText: { flexShrink: 1 },
  roleCell: { flexDirection: 'row', alignItems: 'center', gap: spacing.sm },
  actions: { flexDirection: 'row', gap: spacing.sm },
  compactTrailing: { alignItems: 'flex-end', gap: spacing.sm },
  empty: { padding: spacing.lg },
});

const columnStyles = StyleSheet.create({
  user: { flex: 3 },
  role: { flex: 3 },
  status: { flex: 1.4 },
  actions: { flex: 1.4 },
});

const columnStyleList = [columnStyles.user, columnStyles.role, columnStyles.status, columnStyles.actions];
