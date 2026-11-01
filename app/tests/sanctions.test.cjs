const assert = require('node:assert/strict');
const { test } = require('node:test');
const { readFileSync } = require('node:fs');
const { dirname, join, resolve } = require('node:path');
const ts = require('typescript');

// The same loader as history.test.cjs: relative imports are resolved and loaded.
function loadModule(path) {
  const source = readFileSync(path, 'utf8');
  const compiled = ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS } });
  const exports = {};
  const requireSibling = (specifier) => loadModule(resolve(dirname(path), `${specifier}.ts`));
  new Function('exports', 'require', compiled.outputText)(exports, requireSibling);
  return exports;
}

const sanctionsDir = join(__dirname, '../src/features/sanctions');
const form = loadModule(join(sanctionsDir, 'sanctionForm.ts'));
const text = loadModule(join(sanctionsDir, 'sanctionText.ts'));
const { sanctionQuery, emptySanctionFilter, sanctionSubjects } = loadModule(join(sanctionsDir, 'sanctionFilter.ts'));

const penaltyTypes = [
  { code: 'reprimand', name: 'Reprimenda', unit: null, maxValue: null },
  { code: 'time_penalty', name: 'Penalización de tiempo', unit: 'seconds', maxValue: 300 },
];
const valid = { ...form.emptySanctionForm, driverId: 4, penaltyType: 'time_penalty', penaltyValue: '5', reason: 'Exceder los límites de pista' };

test('a new sanction needs its subject, a reason and a penalty that fits its unit', () => {
  assert.deepEqual(form.validateNewSanction(valid, penaltyTypes), {});
  assert.ok(form.validateNewSanction({ ...valid, driverId: null }, penaltyTypes).driverId);
  assert.ok(form.validateNewSanction({ ...valid, subjectKind: 'team' }, penaltyTypes).teamId);
  assert.ok(form.validateNewSanction({ ...valid, reason: '  ab ' }, penaltyTypes).reason);
  assert.ok(form.validateNewSanction({ ...valid, penaltyType: '' }, penaltyTypes).penaltyType);
  for (const penaltyValue of ['', '0', '301', '2.5', '-1', '5s']) {
    assert.ok(form.validateNewSanction({ ...valid, penaltyValue }, penaltyTypes).penaltyValue, penaltyValue);
  }
  assert.deepEqual(form.validateNewSanction({ ...valid, penaltyType: 'reprimand', penaltyValue: '' }, penaltyTypes), {});
});

test('only the chosen subject and a value the penalty takes travel to the API', () => {
  assert.deepEqual(form.toNewSanction(valid, penaltyTypes), {
    driverId: 4, penaltyType: 'time_penalty', penaltyValue: 5, reason: 'Exceder los límites de pista', eventId: null,
  });
  const onTeam = form.toNewSanction({ ...valid, subjectKind: 'team', teamId: 2, penaltyType: 'reprimand' }, penaltyTypes);
  assert.equal(onTeam.driverId, undefined);
  assert.equal(onTeam.teamId, 2);
  assert.equal(onTeam.penaltyValue, null);
});

test('an amendment needs the reason of the change', () => {
  assert.ok(form.validateAmendment({ ...valid, note: '   ' }, penaltyTypes).note);
  assert.deepEqual(form.validateAmendment({ ...valid, note: 'Revisión' }, penaltyTypes), {});
  assert.equal(form.toAmendment({ ...valid, note: '  Revisión ' }, penaltyTypes).note, 'Revisión');
});

test('the event list follows the championship of the chosen subject', () => {
  const drivers = [{ id: 4, categoryCode: 'f1' }];
  const teams = [{ id: 6, categoryCode: 'f2' }];
  assert.equal(form.subjectCategory(valid, drivers, teams), 'f1');
  assert.equal(form.subjectCategory({ ...valid, subjectKind: 'team', teamId: 6 }, drivers, teams), 'f2');
  assert.equal(form.subjectCategory({ ...valid, driverId: null }, drivers, teams), null);
});

test('penalties read with their unit, and a unit-less one by its name', () => {
  assert.equal(text.penaltyText({ name: 'Multa', unit: 'euros', value: 10000 }), `Multa: ${(10000).toLocaleString('es-AR')} €`);
  assert.equal(text.penaltyText({ name: 'Pérdida de puestos', unit: 'grid_places', value: 1 }), 'Pérdida de puestos: 1 puesto');
  assert.equal(text.penaltyText({ name: 'Reprimenda', unit: null, value: null }), 'Reprimenda');
});

test('a team is told when a sanction is on itself or on one of its drivers', () => {
  const driver = { kind: 'driver', driverFirstName: 'Lewis', driverLastName: 'Hamilton' };
  assert.equal(text.noticeAudienceText(driver, true), 'Sanción a tu piloto Lewis Hamilton');
  assert.equal(text.noticeAudienceText({ kind: 'team' }, true), 'Sanción a tu escudería');
  assert.equal(text.noticeAudienceText(driver, false), null);
});

test('the list query leaves out every filter that is not set', () => {
  assert.deepEqual(JSON.parse(JSON.stringify(sanctionQuery(emptySanctionFilter))), {});
  assert.deepEqual(
    JSON.parse(JSON.stringify(sanctionQuery({ ...emptySanctionFilter, driverId: '4', status: 'revoked', from: '2026-10-01' }))),
    { from: '2026-10-01', driverId: 4, status: 'revoked' },
  );
});

test('historical filters retain a released driver and both former teams without duplicating subjects', () => {
  const subject = { driverId: 7, driverFirstName: 'Ana', driverLastName: 'Pérez' };
  const oldTeam = { id: 4, name: 'Azul' };
  const newTeam = { id: 8, name: 'Rojo' };
  const options = sanctionSubjects([
    { subject, team: newTeam },
    { subject, team: oldTeam },
    { subject: { driverId: null }, team: oldTeam },
  ]);
  assert.deepEqual(options.drivers, [{ value: '7', label: 'Ana Pérez' }]);
  assert.deepEqual(options.teams, [{ value: '4', label: 'Azul' }, { value: '8', label: 'Rojo' }]);
});
