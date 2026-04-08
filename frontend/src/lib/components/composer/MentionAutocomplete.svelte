<!--
@component

Discord-style @mention autocomplete popup.
Shows matching room members when typing @username in chat input.

**Props:**
- `query` - Current search query (without the leading @)
- `members` - Room members to search through
- `onSelect` - Callback when a member is selected (receives login and whether Tab was used)
- `onClose` - Callback to close the popup
-->
<script lang="ts">
  import type { RoomMember } from '$lib/state/room';
  import { fuzzyMatch } from '$lib/fuzzyMatch';
  import { getAvatarInitials } from '$lib/utils/initials';
  import SkeletonImg from '$lib/ui/SkeletonImg.svelte';
  import AutocompletePopup from './AutocompletePopup.svelte';

  type ScoredMember = { member: RoomMember; score: number };

  type Props = {
    query: string;
    members: RoomMember[];
    onSelect: (login: string, viaTab: boolean) => void;
    onClose: () => void;
  };

  let { query, members, onSelect, onClose }: Props = $props();

  let results = $derived.by(() => {
    const scored: ScoredMember[] = [];

    for (const m of members) {
      const loginScore = fuzzyMatch(query, m.login);
      const displayScore = fuzzyMatch(query, m.displayName);
      const bestScore = Math.max(loginScore ?? -1, displayScore ?? -1);

      if (bestScore > 0) {
        scored.push({ member: m, score: bestScore });
      }
    }

    scored.sort((a, b) => b.score - a.score);
    return scored.slice(0, 10);
  });

  let popupRef = $state<{ handleKeyDown: (e: KeyboardEvent) => boolean } | null>(null);

  export function handleKeyDown(event: KeyboardEvent): boolean {
    return popupRef?.handleKeyDown(event) ?? false;
  }

  function handleSelect(result: ScoredMember, key: string) {
    onSelect(result.member.login, key === 'Tab');
  }
</script>

<AutocompletePopup
  bind:this={popupRef}
  items={results}
  getKey={(r) => r.member.id}
  selectKeys={['Tab']}
  onSelect={handleSelect}
  {onClose}
  testid="mention-autocomplete"
  class="md:w-72"
>
  {#snippet item({ item: result })}
    {#if result.member.avatarUrl}
      <SkeletonImg
        loading="lazy"
        src={result.member.avatarUrl}
        alt={result.member.login}
        class="h-6 w-6 shrink-0 rounded-full object-cover"
      />
    {:else}
      <div
        class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-surface-200 text-xs font-semibold text-muted"
      >
        {getAvatarInitials(result.member.displayName, result.member.login)}
      </div>
    {/if}
    <span class="min-w-0 truncate text-sm text-text">{result.member.displayName}</span>
    <span class="min-w-0 truncate text-sm text-muted">@{result.member.login}</span>
  {/snippet}
</AutocompletePopup>
