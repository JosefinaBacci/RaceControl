import type { DriverChanges, NewDriver, TeamDriver } from '@/api/drivers';

export type DriverFieldValues = {
  firstName: string;
  lastName: string;
  licenseNumber: string;
  countryCode: string;
  isStarter: boolean;
  number: string;
};

export const driverFields = ['firstName', 'lastName', 'licenseNumber', 'countryCode', 'number'] as const;

export type DriverField = (typeof driverFields)[number];

export function emptyDriverForm(): DriverFieldValues {
  return { firstName: '', lastName: '', licenseNumber: '', countryCode: '', isStarter: true, number: '' };
}

export function initialDriverForm(driver: TeamDriver): DriverFieldValues {
  return {
    firstName: driver.firstName,
    lastName: driver.lastName,
    licenseNumber: driver.licenseNumber,
    countryCode: driver.countryCode,
    isStarter: driver.isStarter,
    number: driver.number === null ? '' : String(driver.number),
  };
}

export function newDriverInput(values: DriverFieldValues): NewDriver {
  return {
    firstName: values.firstName,
    lastName: values.lastName,
    licenseNumber: values.licenseNumber,
    countryCode: values.countryCode,
    isStarter: values.isStarter,
    number: parseNumber(values.number),
  };
}

// Only the fields the form actually changed, so an untouched driver sends an empty
// patch and the role is not re-recorded as a new period for nothing.
export function driverChanges(driver: TeamDriver, values: DriverFieldValues): DriverChanges {
  const changes: DriverChanges = {};
  if (values.firstName !== driver.firstName) changes.firstName = values.firstName;
  if (values.lastName !== driver.lastName) changes.lastName = values.lastName;
  if (values.licenseNumber !== driver.licenseNumber) changes.licenseNumber = values.licenseNumber;
  if (values.countryCode !== driver.countryCode) changes.countryCode = values.countryCode;
  if (values.isStarter !== driver.isStarter) changes.isStarter = values.isStarter;
  const number = parseNumber(values.number);
  if (number !== null && number !== driver.number) changes.number = number;
  return changes;
}

export function hasDriverChanges(changes: DriverChanges): boolean {
  return Object.values(changes).some((value) => value !== undefined);
}

function parseNumber(value: string): number | null {
  const trimmed = value.trim();
  if (trimmed === '') return null;
  const parsed = Number(trimmed);
  return Number.isFinite(parsed) ? parsed : null;
}

export function driverNumberError(value: string, previousNumber: number | null = null): string | null {
  const trimmed = value.trim();
  if (trimmed === '') {
    return previousNumber === null ? null : 'El número asignado no se puede quitar; ingresá otro número válido.';
  }
  return /^\d{1,3}$/.test(trimmed) ? null : 'El número debe ser un entero entre 0 y 999.';
}
