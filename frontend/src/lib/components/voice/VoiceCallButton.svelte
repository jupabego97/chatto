<!--
@component

Phone icon button for joining voice calls. Shown in the room header.
Hidden when already in this call (the VoiceCallPanel provides leave controls).

States:
- Idle: phone icon (click to join)
- Connecting: spinner
- In call: hidden (panel handles it)
- Disabled: already in another call

**Props:**
- `spaceId` - The space ID
- `roomId` - The room ID
- `livekitUrl` - The LiveKit server WebSocket URL
-->
<script lang="ts">
  import { instanceRegistry } from '$lib/state/instance/registry.svelte';
  import { getActiveInstance } from '$lib/state/activeInstance.svelte';

  const getInstanceId = getActiveInstance();
  const voiceCallState = instanceRegistry.getStore(getInstanceId()).voiceCall;
  import { toast } from '$lib/ui/toast';

  let {
    spaceId,
    roomId,
    livekitUrl
  }: {
    spaceId: string;
    roomId: string;
    livekitUrl: string;
  } = $props();

  let isInThisCall = $derived(voiceCallState.isInCall(spaceId, roomId));
  let isInAnotherCall = $derived(voiceCallState.isInAnyCall && !isInThisCall);
  let isConnecting = $derived(
    voiceCallState.connecting && voiceCallState.spaceId === spaceId && voiceCallState.roomId === roomId
  );

  async function handleJoin() {
    try {
      await voiceCallState.join(livekitUrl, spaceId, roomId);
    } catch {
      toast.error('Failed to join voice call');
    }
  }
</script>

{#if isConnecting}
  <span
    class="iconify animate-spin text-muted uil--spinner"
    title="Connecting..."
  ></span>
{:else if !isInThisCall}
  <button
    type="button"
    class="iconify cursor-pointer text-muted uil--phone hover:text-text disabled:cursor-not-allowed disabled:opacity-50"
    onclick={handleJoin}
    disabled={isInAnotherCall}
    title={isInAnotherCall ? 'Already in another call' : 'Join voice call'}
  >
  </button>
{/if}
