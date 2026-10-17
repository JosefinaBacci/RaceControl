const assert = require('node:assert/strict');
const { test } = require('node:test');
const { readFileSync } = require('node:fs');
const { join } = require('node:path');
const ts = require('typescript');

function loadForm(relativePath) {
  const source = readFileSync(join(__dirname, '../src/features', relativePath), 'utf8');
  const compiled = ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS } });
  const exports = {};
  new Function('exports', compiled.outputText)(exports);
  return exports;
}

const { timingOf, newFormValues } = loadForm('calendar/eventForm.ts');
const { initialSheet, lineError, sheetEntries } = loadForm('scores/scoreSheet.ts');
const values = { ...newFormValues(1, 'f1'), date: '2028-02-29', startTime: '00:00', endTime: '23:59' };

test('calendar accepts leap days and boundary hours', () => {
  assert.deepEqual(timingOf(values), { startsAt: '2028-02-29T00:00:00Z', endsAt: '2028-02-29T23:59:00Z' });
});

test('calendar rejects impossible dates and hours on either input', () => {
  for (const date of ['2026-02-29', '2026-04-31', '2026-13-01', '2026-1-01', '12345-01-01']) {
    assert.equal(timingOf({ ...values, date }), null, date);
  }
  for (const time of ['24:00', '23:60', '99:99', '9:00', '09:000', '-1:30']) {
    assert.equal(timingOf({ ...values, startTime: time }), null, time);
    assert.equal(timingOf({ ...values, endTime: time }), null, time);
  }
});

test('fresh score sheets default to zero while existing absences stay absent', () => {
  assert.deepEqual(initialSheet({ entries: [{ driverId: 1, points: null }] }), { 1: { didRace: true, points: '0' } });
  const grid = { entries: [{ driverId: 1, points: 0 }, { driverId: 2, points: null }] };
  assert.deepEqual(sheetEntries(grid, initialSheet(grid)), [{ driverId: 1, points: 0 }]);
});

test('scores reject blanks, non-decimal syntax, excess precision and overflow', () => {
  for (const points of ['', ' ', '-1', '1.001', '10000', 'Infinity', '0x10', '1e2', 'abc']) {
    assert.notEqual(lineError({ didRace: true, points }), null, points);
  }
  for (const points of ['0', '0.01', '25', '9999.99']) {
    assert.equal(lineError({ didRace: true, points }), null, points);
  }
});

test('driver numbers reject invalid text instead of silently dropping it', () => {
  const { driverNumberError } = loadForm('drivers/driverForm.ts');
  for (const value of ['abc', '1.5', '1e2', '-1', '1000']) assert.notEqual(driverNumberError(value), null);
  for (const value of ['', '0', '99', '999']) assert.equal(driverNumberError(value), null);
  assert.notEqual(driverNumberError('', 44), null);
});
