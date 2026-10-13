import { test, before, after, beforeEach } from 'node:test';
import { readFileSync } from 'node:fs';
import {
  initializeTestEnvironment, assertFails, assertSucceeds,
} from '@firebase/rules-unit-testing';
import {
  doc, getDoc, setDoc, updateDoc, deleteDoc, serverTimestamp, Timestamp, GeoPoint,
} from 'firebase/firestore';

let env;
before(async () => {
  env = await initializeTestEnvironment({
    projectId: 'demo-coexist',
    firestore: { rules: readFileSync(new URL('./firestore.rules', import.meta.url), 'utf8') },
  });
});
after(() => env.cleanup());
beforeEach(() => env.clearFirestore());

const ctx = (uid, claims = {}) => env.authenticatedContext(uid, claims).firestore();
const citizen = () => ctx('c1');
const anon = () => ctx('a1', { firebase: { sign_in_provider: 'anonymous' } });
const official = () => ctx('o1', { role: 'OFFICIAL', region: 'JAPAN', orgId: 'org1' });
const officialOther = () => ctx('o2', { role: 'OFFICIAL', region: 'APAC', orgId: 'org2' });
const admin = () => ctx('adm', { role: 'ADMIN' });
const unauth = () => env.unauthenticatedContext().firestore();
const seed = (path, data) =>
  env.withSecurityRulesDisabled((c) => setDoc(doc(c.firestore(), path), data));
const pt = new GeoPoint(35, 139);

const incident = (o = {}) => ({
  reporterId: 'c1', reporterRole: 'CITIZEN', category: 'WILDLIFE_SIGHTING', region: 'JAPAN',
  location: pt, timestamp: Timestamp.now(), photoUrls: ['https://x.example/a.jpg'],
  verificationStatus: 'PENDING', ...o,
});

test('geofences: public read, admin-only write', async () => {
  await seed('geofences/g', { a: 1 });
  await assertSucceeds(getDoc(doc(unauth(), 'geofences/g')));
  await assertFails(setDoc(doc(citizen(), 'geofences/g'), { a: 2 }));
  await assertFails(setDoc(doc(official(), 'geofences/g'), { a: 2 }));
  await assertSucceeds(setDoc(doc(admin(), 'geofences/g'), { a: 2 }));
});

test('incidents: create rules', async () => {
  await assertSucceeds(setDoc(doc(citizen(), 'incidents/i'), incident()));
  await assertSucceeds(setDoc(doc(anon(), 'incidents/i2'), incident({ reporterId: 'a1' })));
  await assertFails(setDoc(doc(unauth(), 'incidents/i3'), incident()));
  await assertFails(setDoc(doc(citizen(), 'incidents/i4'), incident({ reporterId: 'other' })));
  await assertFails(setDoc(doc(citizen(), 'incidents/i5'), incident({ verificationStatus: 'VERIFIED' })));
  await assertFails(setDoc(doc(citizen(), 'incidents/i6'), incident({ reporterRole: 'ADMIN' })));
  await assertFails(setDoc(doc(citizen(), 'incidents/i7'), incident({ category: 'BOGUS' })));
  await assertFails(setDoc(doc(citizen(), 'incidents/i8'), incident({ extra: 1 })));
  await assertFails(setDoc(doc(citizen(), 'incidents/i9'), incident({ photoUrls: Array(6).fill('https://x.example/a') })));
  await assertFails(setDoc(doc(citizen(), 'incidents/i10'), incident({ description: 'x'.repeat(2001) })));
  await assertFails(setDoc(doc(citizen(), 'incidents/i11'), incident({ photoUrls: 'nope' })));
  // official reports as OFFICIAL only
  await assertSucceeds(setDoc(doc(official(), 'incidents/i12'), incident({ reporterId: 'o1', reporterRole: 'OFFICIAL' })));
  await assertFails(setDoc(doc(official(), 'incidents/i13'), incident({ reporterId: 'o1', reporterRole: 'CITIZEN' })));
});

test('incidents: read rules', async () => {
  await seed('incidents/p', incident());
  await seed('incidents/v', incident({ reporterId: 'zz', verificationStatus: 'VERIFIED' }));
  await assertSucceeds(getDoc(doc(citizen(), 'incidents/p'))); // owner
  await assertFails(getDoc(doc(ctx('c2'), 'incidents/p'))); // other citizen, pending
  await assertSucceeds(getDoc(doc(ctx('c2'), 'incidents/v')));
  await assertFails(getDoc(doc(unauth(), 'incidents/v')));
  await assertFails(getDoc(doc(anon(), 'incidents/v')));
  await assertSucceeds(getDoc(doc(official(), 'incidents/p')));
  await assertFails(getDoc(doc(officialOther(), 'incidents/p')));
  await assertSucceeds(getDoc(doc(admin(), 'incidents/p')));
});

test('incidents: update/delete rules', async () => {
  await seed('incidents/p', incident());
  const ts = serverTimestamp();
  await assertFails(updateDoc(doc(citizen(), 'incidents/p'), { description: 'hi' }));
  await assertFails(updateDoc(doc(citizen(), 'incidents/p'), { verificationStatus: 'VERIFIED' }));
  await assertFails(updateDoc(doc(officialOther(), 'incidents/p'), { verificationStatus: 'VERIFIED', verifiedBy: 'o2', updatedAt: ts }));
  await assertFails(updateDoc(doc(official(), 'incidents/p'), { verificationStatus: 'VERIFIED', verifiedBy: 'o1', updatedAt: ts, description: 'x' }));
  await assertFails(updateDoc(doc(official(), 'incidents/p'), { verificationStatus: 'PENDING', verifiedBy: 'o1', updatedAt: ts }));
  await assertFails(updateDoc(doc(official(), 'incidents/p'), { verificationStatus: 'VERIFIED', verifiedBy: 'someoneelse', updatedAt: ts }));
  await assertFails(updateDoc(doc(official(), 'incidents/p'), { reporterRole: 'ADMIN' }));
  await assertSucceeds(updateDoc(doc(official(), 'incidents/p'), { verificationStatus: 'VERIFIED', verifiedBy: 'o1', updatedAt: ts }));
  await assertSucceeds(updateDoc(doc(admin(), 'incidents/p'), { description: 'admin edit' }));
  await assertFails(deleteDoc(doc(official(), 'incidents/p')));
  await assertSucceeds(deleteDoc(doc(admin(), 'incidents/p')));
});

const rk = (o = {}) => ({
  reporterId: 'c1', location: pt, species: 'DEER', clearanceStatus: 'REPORTED',
  createdAt: serverTimestamp(), ...o,
});

test('roadkill: create, workflow', async () => {
  await assertSucceeds(getDoc(doc(unauth(), 'roadkill/none')));
  await assertFails(setDoc(doc(anon(), 'roadkill/a'), rk({ reporterId: 'a1' })));
  await assertFails(setDoc(doc(citizen(), 'roadkill/b'), rk({ clearanceStatus: 'CLEARED' })));
  await assertFails(setDoc(doc(citizen(), 'roadkill/c'), rk({ assignedOfficialId: 'o1' })));
  await assertSucceeds(setDoc(doc(citizen(), 'roadkill/r'), rk()));
  const ts = serverTimestamp();
  await assertFails(updateDoc(doc(citizen(), 'roadkill/r'), { clearanceStatus: 'DISPATCHED' }));
  // skip a step
  await assertFails(updateDoc(doc(official(), 'roadkill/r'), { clearanceStatus: 'CLEARED', assignedOfficialId: 'o1', clearedAt: ts, updatedAt: ts }));
  // other keys
  await assertFails(updateDoc(doc(official(), 'roadkill/r'), { clearanceStatus: 'DISPATCHED', assignedOfficialId: 'o1', updatedAt: ts, species: 'TIGER' }));
  await assertSucceeds(updateDoc(doc(official(), 'roadkill/r'), { clearanceStatus: 'DISPATCHED', assignedOfficialId: 'o1', updatedAt: ts }));
  await assertFails(updateDoc(doc(official(), 'roadkill/r'), { clearanceStatus: 'REPORTED', updatedAt: ts }));
  await assertSucceeds(updateDoc(doc(official(), 'roadkill/r'), { clearanceStatus: 'CLEARED', clearedAt: ts, updatedAt: ts }));
  await assertSucceeds(updateDoc(doc(admin(), 'roadkill/r'), { description: 'x' }));
});

test('kaki_trees_public: public read, admin write', async () => {
  const d = { geohash: 'xn76u', hazardTier: 'HIGH', status: 'UNMANAGED', region: 'JAPAN' };
  await assertFails(setDoc(doc(citizen(), 'kaki_trees_public/t'), d));
  await assertSucceeds(setDoc(doc(admin(), 'kaki_trees_public/t'), d));
  await assertFails(setDoc(doc(admin(), 'kaki_trees_public/t2'), { ...d, geohash: 'xn76u8abc' }));
  await assertFails(setDoc(doc(admin(), 'kaki_trees_public/t3'), { ...d, ownerId: 'c1' }));
  await assertSucceeds(getDoc(doc(unauth(), 'kaki_trees_public/t')));
});

const tree = (o = {}) => ({
  ownerId: 'c1', region: 'JAPAN', orgId: 'org1', location: pt, status: 'UNMANAGED',
  isTop5PercentHazard: false, createdAt: serverTimestamp(), ...o,
});

test('kaki_trees: create/read/update', async () => {
  await assertSucceeds(setDoc(doc(citizen(), 'kaki_trees/t'), tree()));
  await assertFails(setDoc(doc(citizen(), 'kaki_trees/t1'), tree({ status: 'FELLED' })));
  await assertFails(setDoc(doc(citizen(), 'kaki_trees/t2'), tree({ isTop5PercentHazard: true })));
  await assertFails(setDoc(doc(citizen(), 'kaki_trees/t3'), tree({ hazardScore: 90 })));
  await assertSucceeds(setDoc(doc(citizen(), 'kaki_trees/t4'), tree({ hazardScore: 0 })));
  await assertFails(setDoc(doc(citizen(), 'kaki_trees/t5'), tree({ verifiedByOfficialId: 'x' })));
  await assertFails(setDoc(doc(citizen(), 'kaki_trees/t6'), tree({ ownerId: 'other' })));
  await assertFails(setDoc(doc(anon(), 'kaki_trees/t7'), tree({ ownerId: 'a1' })));

  await assertFails(getDoc(doc(unauth(), 'kaki_trees/t')));
  await assertFails(getDoc(doc(ctx('c2'), 'kaki_trees/t')));
  await assertSucceeds(getDoc(doc(citizen(), 'kaki_trees/t')));
  await assertSucceeds(getDoc(doc(official(), 'kaki_trees/t')));
  await assertFails(getDoc(doc(officialOther(), 'kaki_trees/t')));
  await assertFails(getDoc(doc(ctx('o3', { role: 'OFFICIAL', region: 'JAPAN', orgId: 'org9' }), 'kaki_trees/t')));
  await assertSucceeds(getDoc(doc(admin(), 'kaki_trees/t')));

  const ts = serverTimestamp();
  await assertFails(updateDoc(doc(citizen(), 'kaki_trees/t'), { status: 'FELLED', updatedAt: ts }));
  await assertFails(updateDoc(doc(officialOther(), 'kaki_trees/t'), { status: 'PRUNED', verifiedByOfficialId: 'o2', updatedAt: ts }));
  await assertFails(updateDoc(doc(official(), 'kaki_trees/t'), { status: 'UNMANAGED', verifiedByOfficialId: 'o1', updatedAt: ts }));
  await assertFails(updateDoc(doc(official(), 'kaki_trees/t'), { status: 'PRUNED', verifiedByOfficialId: 'o1', updatedAt: ts, hazardScore: 5 }));
  await assertSucceeds(updateDoc(doc(official(), 'kaki_trees/t'), { status: 'PRUNED', verifiedByOfficialId: 'o1', ownerConsentRef: 'consent/1', updatedAt: ts }));
  await assertSucceeds(updateDoc(doc(admin(), 'kaki_trees/t'), { hazardScore: 80, isTop5PercentHazard: true }));
});

const crop = (o = {}) => ({
  reporterId: 'c1', region: 'JAPAN', location: pt, cropType: 'persimmon',
  growthStage: 'RIPENING', createdAt: serverTimestamp(), ...o,
});

test('crop_attractants', async () => {
  await assertFails(getDoc(doc(unauth(), 'crop_attractants/x')));
  await assertSucceeds(setDoc(doc(citizen(), 'crop_attractants/c'), crop()));
  await assertFails(setDoc(doc(citizen(), 'crop_attractants/c2'), crop({ attractantRiskIndex: 99 })));
  await assertFails(setDoc(doc(citizen(), 'crop_attractants/c3'), crop({ growthStage: 'HARVESTED_CLEARED' })));
  await assertSucceeds(getDoc(doc(ctx('c2'), 'crop_attractants/c')));
  const ts = serverTimestamp();
  await assertFails(updateDoc(doc(citizen(), 'crop_attractants/c'), { growthStage: 'READY_FOR_HARVEST', updatedAt: ts }));
  await assertFails(updateDoc(doc(official(), 'crop_attractants/c'), { growthStage: 'READY_FOR_HARVEST', cropType: 'x', updatedAt: ts }));
  await assertSucceeds(updateDoc(doc(official(), 'crop_attractants/c'), { growthStage: 'READY_FOR_HARVEST', updatedAt: ts }));
  await assertSucceeds(updateDoc(doc(admin(), 'crop_attractants/c'), { attractantRiskIndex: 50 }));
});

const pasture = (o = {}) => ({
  ownerId: 'c1', region: 'JAPAN', name: 'North', location: pt,
  verificationStatus: 'PENDING', createdAt: serverTimestamp(), ...o,
});

test('livestock_pastures', async () => {
  await assertSucceeds(setDoc(doc(citizen(), 'livestock_pastures/p'), pasture()));
  for (const k of [{ isKillSite: true }, { killReportedAt: Timestamp.now() },
                   { cachingRiskExpiry: Timestamp.now() }, { activeThreatTier: 'HIGH' }]) {
    await assertFails(setDoc(doc(citizen(), 'livestock_pastures/q'), pasture(k)));
  }
  await assertFails(setDoc(doc(citizen(), 'livestock_pastures/q2'), pasture({ verificationStatus: 'VERIFIED' })));
  await assertFails(getDoc(doc(unauth(), 'livestock_pastures/p')));
  await assertSucceeds(getDoc(doc(ctx('c2'), 'livestock_pastures/p')));
  const ts = serverTimestamp();
  await assertFails(updateDoc(doc(citizen(), 'livestock_pastures/p'), { isKillSite: true }));
  await assertFails(updateDoc(doc(official(), 'livestock_pastures/p'), { isKillSite: true, updatedAt: ts }));
  await assertSucceeds(updateDoc(doc(official(), 'livestock_pastures/p'), { verificationStatus: 'VERIFIED', verifiedBy: 'o1', updatedAt: ts }));
  await assertSucceeds(updateDoc(doc(admin(), 'livestock_pastures/p'), { isKillSite: true }));
});

test('users', async () => {
  await seed('users/c1', { displayName: 'K', createdAt: Timestamp.now(), orgId: 'org1' });
  await seed('users/c9', { displayName: 'Z', orgId: 'org9' });
  await assertSucceeds(getDoc(doc(citizen(), 'users/c1')));
  await assertFails(getDoc(doc(ctx('c2'), 'users/c1')));
  await assertFails(getDoc(doc(unauth(), 'users/c1')));
  await assertSucceeds(getDoc(doc(official(), 'users/c1')));
  await assertFails(getDoc(doc(official(), 'users/c9')));
  await assertSucceeds(getDoc(doc(admin(), 'users/c1')));
  await assertSucceeds(updateDoc(doc(citizen(), 'users/c1'), { displayName: 'New' }));
  await assertFails(updateDoc(doc(citizen(), 'users/c1'), { role: 'ADMIN' }));
  await assertFails(updateDoc(doc(citizen(), 'users/c1'), { orgId: 'org2' }));
  await assertFails(updateDoc(doc(citizen(), 'users/c1'), { region: 'JAPAN' }));
  await assertFails(updateDoc(doc(citizen(), 'users/c1'), { displayName: 'x'.repeat(101) }));
  await assertFails(updateDoc(doc(ctx('c2'), 'users/c1'), { displayName: 'hax' }));
  await assertFails(setDoc(doc(ctx('c3'), 'users/c3'), { displayName: 'A', role: 'ADMIN' }));
  await assertSucceeds(setDoc(doc(ctx('c3'), 'users/c3'), { displayName: 'A' }));
  await assertSucceeds(updateDoc(doc(admin(), 'users/c1'), { role: 'OFFICIAL' }));
});

test('incident_memories + default deny', async () => {
  await seed('incident_memories/m', { a: 1 });
  await assertFails(getDoc(doc(citizen(), 'incident_memories/m')));
  await assertFails(getDoc(doc(official(), 'incident_memories/m')));
  await assertSucceeds(getDoc(doc(admin(), 'incident_memories/m')));
  await assertFails(setDoc(doc(admin(), 'incident_memories/m'), { a: 2 }));
  await assertFails(getDoc(doc(admin(), 'other/x')));
  await assertFails(getDoc(doc(unauth(), 'other/x')));
});
