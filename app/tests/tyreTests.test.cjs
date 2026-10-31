const assert = require('node:assert/strict');
const { test } = require('node:test');
const { join } = require('node:path');

const { loadModule } = require('./loadModule.cjs');

const { compoundRecords, untestedCompounds, compoundError } = loadModule(
  join(__dirname, '../src/features/tyreTests/tyreRecord.ts'),
);

const soft = { code: 'c5', name: 'C5' };
const wet = { code: 'wet', name: 'Lluvia' };
const revision = (id, type, isCurrent) => ({ id, type, isCurrent });

test('each compound keeps its chain and the result in force', () => {
  const records = compoundRecords([revision(1, soft, false), revision(2, wet, true), revision(3, soft, true)]);

  assert.deepEqual(records.map((record) => record.compound.code), ['c5', 'wet']);
  assert.deepEqual(records[0].revisions.map((entry) => entry.id), [1, 3]);
  assert.equal(records[0].current.id, 3);
  assert.equal(records[1].current.id, 2);
});

test('only compounds without a result can get an original one', () => {
  assert.deepEqual(untestedCompounds([soft, wet], [revision(1, soft, true)]), [wet]);
  assert.deepEqual(untestedCompounds([soft], [revision(1, soft, false), revision(2, soft, true)]), []);
});

test('the compound is required', () => {
  assert.ok(compoundError(''));
  assert.equal(compoundError('c1'), undefined);
});
