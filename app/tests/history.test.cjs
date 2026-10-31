const assert = require('node:assert/strict');
const { test } = require('node:test');
const { join } = require('node:path');

const { loadModule } = require('./loadModule.cjs');

const historyDir = join(__dirname, '../src/features/history');
const { isDay, formatDay, formatInstantDay, todayDay } = loadModule(join(historyDir, 'day.ts'));
const { validateDateRange, historyQuery } = loadModule(join(historyDir, 'historyFilter.ts'));
const { validateResultForm, correctionNoteError, resultPayload } = loadModule(join(historyDir, 'resultForm.ts'));

test('days accept leap days and reject impossible or malformed ones', () => {
  assert.equal(isDay('2028-02-29'), true);
  for (const day of ['2026-02-29', '2026-04-31', '2026-13-01', '2026-1-01', '01/03/2026', '']) {
    assert.equal(isDay(day), false, day);
  }
});

test('days are formatted from their UTC digits, never shifted by the device zone', () => {
  assert.equal(formatDay('2026-03-01'), '01/03/2026');
  assert.equal(formatInstantDay('2026-03-01T01:30:00Z'), '01/03/2026');
  assert.equal(formatInstantDay('2026-02-28T23:30:00-03:00'), '01/03/2026');
  assert.equal(todayDay(new Date('2026-03-01T02:00:00Z')), '2026-03-01');
});

test('date ranges reject malformed bounds and reversed ranges, and accept open ones', () => {
  assert.deepEqual(validateDateRange({ from: '', to: '' }), {});
  assert.deepEqual(validateDateRange({ from: '2026-03-01', to: '2026-03-01' }), {});
  assert.ok(validateDateRange({ from: '2026-02-30', to: '' }).from);
  assert.ok(validateDateRange({ from: '', to: 'mañana' }).to);
  assert.ok(validateDateRange({ from: '2026-03-02', to: '2026-03-01' }).to);
});

test('history queries leave unset filters out', () => {
  assert.deepEqual(historyQuery({ from: ' ', to: '2026-03-01', type: '' }), { from: undefined, to: '2026-03-01', type: undefined });
});

test('result forms require a day and an outcome, and notes when it failed', () => {
  assert.deepEqual(validateResultForm({ conductedOn: '2026-03-01', passed: true, notes: '' }), {});
  const empty = validateResultForm({ conductedOn: '', passed: null, notes: '' });
  assert.ok(empty.conductedOn && empty.passed);
  assert.ok(validateResultForm({ conductedOn: '2026-03-01', passed: false, notes: '  ' }).notes);
  assert.ok(validateResultForm({ conductedOn: '2026-03-01', passed: true, notes: 'x'.repeat(2001) }).notes);
  assert.deepEqual(resultPayload({ conductedOn: ' 2026-03-01 ', passed: false, notes: ' alerón ' }), {
    conductedOn: '2026-03-01',
    passed: false,
    notes: 'alerón',
  });
});

test('a correction needs a reason within the length limit', () => {
  assert.ok(correctionNoteError(' '));
  assert.ok(correctionNoteError('x'.repeat(2001)));
  assert.equal(correctionNoteError('equipo mal elegido'), undefined);
});
