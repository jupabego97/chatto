<script lang="ts">
  import RoomRouteLayout from '../../[roomId]/RoomRouteLayout.svelte';

  let { data, children } = $props();
</script>

<RoomRouteLayout {data} {children} routeKind="name" />
