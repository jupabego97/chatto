/**
 * Shared constants for the Chatto frontend.
 *
 * IMPORTANT: The DM_SPACE_ID must match the backend constant in
 * cli/internal/core/dm.go (DMSpaceID = "DM")
 *
 * Note: GraphQL queries use literal strings (e.g., space(id: "DM"))
 * and cannot reference this constant. When searching for usages, also
 * grep for the literal string "DM" in .svelte and .ts files.
 */

/**
 * The well-known space ID for direct messages.
 * DM conversations are rooms within this system space.
 */
export const DM_SPACE_ID = 'DM';
