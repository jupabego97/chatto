<script lang="ts">
  import { page } from '$app/state';
  import { SvelteSet } from 'svelte/reactivity';
  import { useConnection } from '$lib/state/instance/connection.svelte';
  import { graphql } from '$lib/gql';
  import { Panel } from '$lib/components/admin';
  import PaneHeader from '$lib/ui/PaneHeader.svelte';
  import PageTitle from '$lib/ui/PageTitle.svelte';
  import { toast } from '$lib/ui/toast';
  import { PermissionGrid, type PermissionState } from '$lib/components/rbac';

  type RoleOverride = {
    roleName: string;
    displayName: string;
    isInstanceRole: boolean;
    isSystem: boolean;
    position: number;
    permissions: string[];
    permissionDenials: string[];
  };

  const connection = useConnection();
  const spaceId = $derived(page.params.spaceId!);
  const roomId = $derived(page.params.roomId!);

  let spaceRoles = $state<RoleOverride[]>([]);
  let instanceRoles = $state<RoleOverride[]>([]);
  let availablePermissions = $state<string[]>([]);
  let loading = $state(true);
  let error = $state<string | null>(null);
  let updating = $state<string | null>(null);

  // Track which role sections are expanded
  let expandedRoles = new SvelteSet<string>();

  function toggleRole(roleName: string) {
    if (expandedRoles.has(roleName)) {
      expandedRoles.delete(roleName);
    } else {
      expandedRoles.add(roleName);
    }
  }

  async function loadData() {
    const currentSpaceId = spaceId;
    const currentRoomId = roomId;

    loading = true;
    error = null;

    const resp = await connection().client.query(
      graphql(`
        query RoomPermissionOverrides($spaceId: ID!, $roomId: ID!) {
          room(spaceId: $spaceId, roomId: $roomId) {
            id
            name
            availableRoomPermissions
            roomPermissionOverrides {
              roleName
              displayName
              isInstanceRole
              isSystem
              position
              permissions
              permissionDenials
            }
          }
        }
      `),
      { spaceId: currentSpaceId, roomId: currentRoomId }
    );

    // Stale response guard
    if (spaceId !== currentSpaceId || roomId !== currentRoomId) return;

    loading = false;

    if (resp.error) {
      error = resp.error.message;
      return;
    }

    if (!resp.data?.room) {
      error = 'Room not found';
      return;
    }

    availablePermissions = resp.data.room.availableRoomPermissions;

    const overrides = resp.data.room.roomPermissionOverrides;
    spaceRoles = overrides.filter((r) => !r.isInstanceRole).sort((a, b) => a.position - b.position);
    instanceRoles = overrides
      .filter((r) => r.isInstanceRole)
      .sort((a, b) => a.position - b.position);
  }

  $effect(() => {
    if (spaceId && roomId) {
      loadData();
    }
  });

  function findRole(roleName: string): RoleOverride | undefined {
    return (
      spaceRoles.find((r) => r.roleName === roleName) ??
      instanceRoles.find((r) => r.roleName === roleName)
    );
  }

  async function setPermissionState(
    roleName: string,
    permission: string,
    newState: PermissionState
  ) {
    updating = `${roleName}:${permission}`;
    error = null;

    let mutation;
    switch (newState) {
      case 'allow':
        mutation = graphql(`
          mutation GrantRoomPermission($input: GrantRoomPermissionInput!) {
            grantRoomPermission(input: $input)
          }
        `);
        break;
      case 'deny':
        mutation = graphql(`
          mutation DenyRoomPermission($input: DenyRoomPermissionInput!) {
            denyRoomPermission(input: $input)
          }
        `);
        break;
      case 'neutral':
        mutation = graphql(`
          mutation ClearRoomPermission($input: ClearRoomPermissionInput!) {
            clearRoomPermission(input: $input)
          }
        `);
        break;
    }

    const resp = await connection().client.mutation(mutation, {
      input: { spaceId, roomId, role: roleName, permission }
    });

    if (resp.error) {
      error = resp.error.message;
    } else {
      // Optimistic update
      const role = findRole(roleName);
      if (role) {
        role.permissions = role.permissions.filter((p) => p !== permission);
        role.permissionDenials = role.permissionDenials.filter((p) => p !== permission);

        if (newState === 'allow') {
          role.permissions = [...role.permissions, permission];
          toast.success(`Granted ${permission} for ${role.displayName}`);
        } else if (newState === 'deny') {
          role.permissionDenials = [...role.permissionDenials, permission];
          toast.success(`Denied ${permission} for ${role.displayName}`);
        } else {
          toast.success(`Cleared ${permission} for ${role.displayName}`);
        }
      }
    }

    updating = null;
  }

  function hasOverrides(role: RoleOverride): boolean {
    return role.permissions.length > 0 || role.permissionDenials.length > 0;
  }
</script>

<PageTitle title="Room Permissions" />

<div class="flex min-h-0 min-w-0 flex-1 flex-col">
  <PaneHeader
    title="Room Permissions"
    subtitle="Configure per-room permission overrides for roles"
    showMobileNav
  />

  <div class="flex flex-col gap-6 overflow-y-auto p-6">
    {#if error}
      <div class="rounded-lg border border-danger/20 bg-danger/10 p-4 text-danger">
        {error}
      </div>
    {/if}

    {#if loading}
      <div class="text-muted">Loading...</div>
    {:else}
      <div class="bg-surface-2 rounded-lg border border-border p-4 text-sm text-muted">
        Room-level overrides take precedence over space defaults. Set <strong>Allow</strong> or
        <strong>Deny</strong> to override. <strong>Neutral</strong> means the permission inherits from
        the role's space-level configuration.
      </div>

      {#if spaceRoles.length > 0}
        <Panel title="Space Roles" icon="iconify uil--users-alt">
          <div class="flex flex-col divide-y divide-border">
            {#each spaceRoles as role (role.roleName)}
              <div>
                <button
                  class="hover:bg-surface-2 flex w-full cursor-pointer items-center gap-3 px-4 py-3 text-left"
                  onclick={() => toggleRole(role.roleName)}
                >
                  <span
                    class={[
                      'iconify text-lg transition-transform',
                      expandedRoles.has(role.roleName) ? 'uil--angle-down' : 'uil--angle-right'
                    ]}
                  ></span>
                  <span class="font-medium">{role.displayName}</span>
                  {#if role.isSystem}
                    <span class="bg-surface-3 rounded px-1.5 py-0.5 text-xs text-muted">
                      system
                    </span>
                  {/if}
                  {#if hasOverrides(role)}
                    <span class="rounded bg-primary/10 px-1.5 py-0.5 text-xs text-primary">
                      {role.permissions.length + role.permissionDenials.length} override(s)
                    </span>
                  {/if}
                </button>

                {#if expandedRoles.has(role.roleName)}
                  <div class="px-4 pb-4">
                    <PermissionGrid
                      permissions={availablePermissions}
                      grantedPermissions={role.permissions}
                      deniedPermissions={role.permissionDenials}
                      updatingPermission={updating?.startsWith(`${role.roleName}:`)
                        ? updating.slice(role.roleName.length + 1)
                        : null}
                      onSetState={(permission, state) =>
                        setPermissionState(role.roleName, permission, state)}
                    />
                  </div>
                {/if}
              </div>
            {/each}
          </div>
        </Panel>
      {/if}

      {#if instanceRoles.length > 0}
        <Panel title="Instance Roles" icon="iconify uil--globe">
          <div class="flex flex-col divide-y divide-border">
            {#each instanceRoles as role (role.roleName)}
              <div>
                <button
                  class="hover:bg-surface-2 flex w-full cursor-pointer items-center gap-3 px-4 py-3 text-left"
                  onclick={() => toggleRole(role.roleName)}
                >
                  <span
                    class={[
                      'iconify text-lg transition-transform',
                      expandedRoles.has(role.roleName) ? 'uil--angle-down' : 'uil--angle-right'
                    ]}
                  ></span>
                  <span class="font-medium">{role.displayName}</span>
                  {#if role.isSystem}
                    <span class="bg-surface-3 rounded px-1.5 py-0.5 text-xs text-muted">
                      system
                    </span>
                  {/if}
                  {#if hasOverrides(role)}
                    <span class="rounded bg-primary/10 px-1.5 py-0.5 text-xs text-primary">
                      {role.permissions.length + role.permissionDenials.length} override(s)
                    </span>
                  {/if}
                </button>

                {#if expandedRoles.has(role.roleName)}
                  <div class="px-4 pb-4">
                    <PermissionGrid
                      permissions={availablePermissions}
                      grantedPermissions={role.permissions}
                      deniedPermissions={role.permissionDenials}
                      updatingPermission={updating?.startsWith(`${role.roleName}:`)
                        ? updating.slice(role.roleName.length + 1)
                        : null}
                      onSetState={(permission, state) =>
                        setPermissionState(role.roleName, permission, state)}
                    />
                  </div>
                {/if}
              </div>
            {/each}
          </div>
        </Panel>
      {/if}
    {/if}
  </div>
</div>
