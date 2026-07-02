import { readdir, readFile } from 'node:fs/promises';
import path from 'node:path';

const root = process.cwd();
const scanRoots = ['src'];
const blockedPattern = /serverRegistry\.getStore\s*\(\s*getActiveServer\s*\(/;
const ignoredRuntimeNames = new Set([
	'RoomSidebarTestHarness.svelte',
	'VoiceCallPanelStoryHarness.svelte'
]);

async function* walk(dir) {
	for (const entry of await readdir(dir, { withFileTypes: true })) {
		const fullPath = path.join(dir, entry.name);
		if (entry.isDirectory()) {
			yield* walk(fullPath);
			continue;
		}
		if (!entry.isFile()) continue;
		if (!/\.(svelte|svelte\.[tj]s|[tj]s)$/.test(entry.name)) continue;
		if (entry.name.includes('.spec.')) continue;
		if (ignoredRuntimeNames.has(entry.name)) continue;
		yield fullPath;
	}
}

const violations = [];
for (const relativeRoot of scanRoots) {
	for await (const file of walk(path.join(root, relativeRoot))) {
		const source = await readFile(file, 'utf8');
		if (blockedPattern.test(source)) {
			violations.push(path.relative(root, file));
		}
	}
}

if (violations.length > 0) {
	console.error(
		[
			'Do not derive room subtree state with serverRegistry.getStore(getActiveServer()).',
			'Use useActiveServerScope() so the active server, segment, store, and connection share one tracked scope.',
			'',
			...violations.map((file) => `- ${file}`)
		].join('\n')
	);
	process.exit(1);
}
