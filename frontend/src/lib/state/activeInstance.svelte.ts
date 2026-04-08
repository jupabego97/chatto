import { createContext } from 'svelte';

/**
 * Svelte context for the active instance ID.
 *
 * Set by the [[instanceId=hostname]] layout to make the URL-derived
 * instance ID available to all child components.
 *
 * The value is a getter function — call it to get the current instance ID.
 * Must be called during component initialization (not in event handlers).
 */
export const [getActiveInstance, setActiveInstance] = createContext<() => string>();
