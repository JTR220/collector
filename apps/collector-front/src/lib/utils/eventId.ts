// Mapping deterministe id numerique -> UUID, symetrique de
// catalog-service/events/event.go (ToEventUUID).
export function toEventUuid(id: number): string {
	return '00000000-0000-0000-0000-' + id.toString(16).padStart(12, '0');
}

export function fromEventUuid(uuid: string): number {
	return parseInt(uuid.slice(-12), 16);
}
