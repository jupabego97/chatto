<script lang="ts">
  import { getPermissionDescription } from '$lib/permissions';

  type RoleWithPermissions = {
    name: string;
    displayName: string;
    permissions: string[];
    permissionDenials: string[];
  };

  type EffectivePermission = {
    id: string;
    state: 'granted' | 'denied' | 'none';
    grantedBy: string[];
    deniedBy: string[];
  };

  let {
    allPermissions,
    userRoleNames,
    roles
  }: {
    allPermissions: string[];
    userRoleNames: string[];
    roles: RoleWithPermissions[];
  } = $props();

  const effectivePermissions = $derived.by(() => {
    const userRoles = roles.filter((r) => userRoleNames.includes(r.name));

    return allPermissions.map((permId): EffectivePermission => {
      const grantedBy = userRoles
        .filter((r) => r.permissions.includes(permId))
        .map((r) => r.displayName);
      const deniedBy = userRoles
        .filter((r) => r.permissionDenials.includes(permId))
        .map((r) => r.displayName);

      // Deny wins over grant
      const state = deniedBy.length > 0 ? 'denied' : grantedBy.length > 0 ? 'granted' : 'none';

      return { id: permId, state, grantedBy, deniedBy };
    });
  });
</script>

<div class="flex flex-col">
  <div class="grid grid-cols-[1fr_auto_1fr] gap-x-4 gap-y-1 text-sm">
    <div class="border-b border-border pb-2 font-medium text-muted">Permission</div>
    <div class="border-b border-border pb-2 text-center font-medium text-muted">Status</div>
    <div class="border-b border-border pb-2 font-medium text-muted">Source</div>

    {#each effectivePermissions as perm (perm.id)}
      <div class="border-b border-border/50 py-2">
        <div class="font-medium">{perm.id}</div>
        <div class="text-xs text-muted">{getPermissionDescription(perm.id)}</div>
      </div>
      <div class="flex items-center justify-center border-b border-border/50 py-2">
        {#if perm.state === 'granted'}
          <span class="iconify text-lg text-success uil--check-circle" title="Granted"></span>
        {:else if perm.state === 'denied'}
          <span class="iconify text-lg text-danger uil--times-circle" title="Denied"></span>
        {:else}
          <span class="iconify text-lg text-muted uil--minus-circle" title="No access"></span>
        {/if}
      </div>
      <div class="flex items-center border-b border-border/50 py-2 text-xs text-muted">
        {#if perm.state === 'denied'}
          <span>Denied by {perm.deniedBy.join(', ')}</span>
          {#if perm.grantedBy.length > 0}
            <span class="ml-1">(also granted by {perm.grantedBy.join(', ')})</span>
          {/if}
        {:else if perm.state === 'granted'}
          <span>{perm.grantedBy.join(', ')}</span>
        {:else}
          <span class="italic">—</span>
        {/if}
      </div>
    {/each}
  </div>
</div>
