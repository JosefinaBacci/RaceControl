const assert = require('node:assert/strict');
const { existsSync } = require('node:fs');
const { test } = require('node:test');
const { join } = require('node:path');

const { loadModule } = require('./loadModule.cjs');

const navigation = loadModule(join(__dirname, '../src/navigation/tabSpecs.ts'));
const { maxPhonePrimaryTabs, maxPhoneLabelLength, moreTab, moreMenuTabs, moreMenuBadge, phoneTabLabel } = navigation;

const groups = [
  { name: 'public', tabs: navigation.publicTabs, directory: '(public)' },
  { name: 'team', tabs: navigation.teamTabs, directory: 'team' },
  { name: 'fia', tabs: navigation.fiaTabs, directory: 'fia' },
];

function routeExists(directory, name) {
  const base = join(__dirname, '../app', directory, name);
  return existsSync(`${base}.tsx`) || existsSync(join(base, '_layout.tsx'));
}

for (const group of groups) {
  test(`${group.name}: the phone bar fits the primary tabs plus "Más"`, () => {
    const primary = group.tabs.filter((tab) => tab.isPhonePrimary);
    assert.ok(primary.length <= maxPhonePrimaryTabs, `${primary.length} primary tabs`);
    for (const tab of [...primary, moreTab]) {
      assert.ok(phoneTabLabel(tab).length <= maxPhoneLabelLength, `"${phoneTabLabel(tab)}" is too long for a phone tab`);
    }
  });

  test(`${group.name}: every tab and "Más" is a route of the group`, () => {
    for (const tab of [...group.tabs, moreTab]) {
      assert.ok(routeExists(group.directory, tab.name), `app/${group.directory}/${tab.name} is missing`);
    }
    assert.equal(new Set(group.tabs.map((tab) => tab.name)).size, group.tabs.length);
  });

  test(`${group.name}: "Más" lists exactly the tabs the phone bar hides`, () => {
    const hidden = group.tabs.filter((tab) => !tab.isPhonePrimary).map((tab) => tab.name);
    assert.deepEqual(moreMenuTabs(group.tabs).map((tab) => tab.name), hidden);
  });
}

test('"Más" adds up only the badges of the tabs it lists', () => {
  const badges = { inbox: 2, notifications: 3, messages: 4 };
  assert.equal(moreMenuBadge(navigation.teamTabs, badges), 3);
  assert.equal(moreMenuBadge(navigation.fiaTabs, { messages: 5 }), 5);
  assert.equal(moreMenuBadge(navigation.publicTabs, {}), 0);
});

test('a short phone label replaces the title only on the phone bar', () => {
  const team = navigation.teamTabs.find((tab) => tab.name === 'index');
  assert.equal(team.title, 'Mi escudería');
  assert.equal(phoneTabLabel(team), 'Escudería');
});
