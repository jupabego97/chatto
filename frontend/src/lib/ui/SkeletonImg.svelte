<script lang="ts">
  import type { HTMLImgAttributes } from 'svelte/elements';

  let { class: className, ...rest }: HTMLImgAttributes = $props();
  let loaded = $state(false);
</script>

<img
  class={[className, !loaded && 'skeleton']}
  onload={() => (loaded = true)}
  onerror={() => (loaded = false)}
  {...rest}
/>
