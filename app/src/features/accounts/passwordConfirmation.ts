export function passwordConfirmationError(newPassword: string, confirmation: string): string | null {
  if (confirmation === '' || confirmation === newPassword) {
    return null;
  }
  return 'Las contraseñas no coinciden';
}
