<script lang="ts">
  import { provideEventBus } from '$lib/eventBus.svelte';
  import { createTypingIndicator } from './useTypingIndicator.svelte';

  let {
    serverId,
    roomId,
    threadRootEventId = null,
    currentUserId
  }: {
    serverId: string;
    roomId: string;
    threadRootEventId?: string | null;
    currentUserId: string | null;
  } = $props();

  provideEventBus(() => serverId);

  const typingIndicator = createTypingIndicator(() => ({
    roomId,
    threadRootEventId,
    currentUserId
  }));
</script>

<div data-testid="typing-users">{typingIndicator.userIds.join(',')}</div>
