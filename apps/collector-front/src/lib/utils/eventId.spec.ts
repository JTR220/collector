import { describe, it, expect } from 'vitest';
import { toEventUuid, fromEventUuid } from './eventId';

describe('eventId', () => {
	it('mappe les IDs comme le Go (ToEventUUID)', () => {
		expect(toEventUuid(0)).toBe('00000000-0000-0000-0000-000000000000');
		expect(toEventUuid(1)).toBe('00000000-0000-0000-0000-000000000001');
		expect(toEventUuid(255)).toBe('00000000-0000-0000-0000-0000000000ff');
	});

	it('est reversible', () => {
		for (const id of [0, 1, 42, 255, 100000]) {
			expect(fromEventUuid(toEventUuid(id))).toBe(id);
		}
	});
});
