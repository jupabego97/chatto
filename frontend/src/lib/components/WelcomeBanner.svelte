<!--
@component

Shows a welcome banner to newly verified users. The banner auto-dismisses
after 5 seconds and can be manually dismissed. When shown, removes the
`welcome` query parameter from the URL.

Only renders when the `welcome=true` query parameter is present.
-->
<script lang="ts">
  import { page } from '$app/state';

  let showWelcome = $state(page.url.searchParams.get('welcome') === 'true');

  // Clear the welcome param from URL after showing
  $effect(() => {
    if (showWelcome) {
      // Remove the query param from URL without navigation
      const url = new URL(window.location.href);
      url.searchParams.delete('welcome');
      window.history.replaceState({}, '', url.toString());

      // Auto-dismiss after 5 seconds
      const timer = setTimeout(() => {
        showWelcome = false;
      }, 5000);
      return () => clearTimeout(timer);
    }
  });
</script>

{#if showWelcome}
  <div
    class="mb-2 flex items-center justify-between rounded-lg bg-green-500/10 px-4 py-2 text-sm text-green-600 dark:text-green-400"
  >
    <span>Welcome to Chatto! Your email has been verified and your account is ready.</span>
    <button
      type="button"
      class="ml-4 hover:text-green-800 dark:hover:text-green-200"
      onclick={() => (showWelcome = false)}
      title="Dismiss"
    >
      <span class="iconify uil--times"></span>
    </button>
  </div>
{/if}
