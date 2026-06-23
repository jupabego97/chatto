<script lang="ts">
  import type { MediaViewerItem } from '$lib/ui/mediaViewer';

  let {
    items,
    index,
    onclose
  }: {
    items: MediaViewerItem[];
    index: number;
    onclose: () => void;
  } = $props();
</script>

<div
  data-testid="media-viewer"
  data-index={index}
  data-items={JSON.stringify(items)}
>
  <button type="button" onclick={onclose}>Close media viewer</button>
</div>
