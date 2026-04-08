<script lang="ts">
  import { goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import { instanceIdToSegment } from '$lib/navigation';
  import { getActiveInstance } from '$lib/state/activeInstance.svelte';
  import { useConnection } from '$lib/state/instance/connection.svelte';
  import { graphql } from '$lib/gql';
  import { Panel } from '$lib/components/admin';
  import PaneHeader from '$lib/ui/PaneHeader.svelte';
  import PageTitle from '$lib/ui/PageTitle.svelte';
  import { Button } from '$lib/ui/form';
  import { FormError } from '$lib/ui/form';
  import { toast } from '$lib/ui/toast';
  import { PermissionGrid, type PermissionState } from '$lib/components/rbac';

  const getInstanceId = getActiveInstance();
  const connection = useConnection();
  const spaceId = $derived(page.params.spaceId!);
  const instanceRoleName = $derived(page.params.name!);

  type InstanceRoleConfig = {
    role: {
      name: string;
      displayName: string;
      description: string;
    };
    permissions: string[];
    permissionDenials: string[];
  };

  let roleConfig = $state<InstanceRoleConfig | null>(null);
  let allPermissions = $state<string[]>([]);
  let canManageRoles = $state(false);
  let loading = $state(true);
  let updating = $state<string | null>(null);
  let error = $state<string | null>(null);

  async function loadData() {
    loading = true;
    error = null;

    const resp = await connection().client.query(
      graphql(`
        query InstanceRoleSpaceDetail($spaceId: ID!) {
          space(id: $spaceId) {
            id
            name
            instanceRoleConfigs {
              role {
                name
                displayName
                description
              }
              permissions
              permissionDenials
            }
            availablePermissions
            viewerCanManageRoles
          }
        }
      `),
      { spaceId }
    );

    if (resp.error) {
      error = resp.error.message;
      loading = false;
      return;
    }

    if (!resp.data?.space) {
      error = 'Space not found';
      loading = false;
      return;
    }

    // Find the specific instance role config
    const configs = resp.data.space.instanceRoleConfigs;
    const found = configs.find((c) => c.role.name === instanceRoleName);

    roleConfig = found ?? null;
    allPermissions = resp.data.space.availablePermissions;
    canManageRoles = resp.data.space.viewerCanManageRoles;

    loading = false;
  }

  $effect(() => {
    if (spaceId && instanceRoleName) {
      loadData();
    }
  });

  async function setPermissionState(permission: string, newState: PermissionState) {
    if (!roleConfig) return;

    updating = permission;
    error = null;

    let mutation;
    switch (newState) {
      case 'allow':
        mutation = graphql(`
          mutation GrantInstanceRoleSpacePermission($input: GrantInstanceRoleSpacePermissionInput!) {
            grantInstanceRoleSpacePermission(input: $input)
          }
        `);
        break;
      case 'deny':
        mutation = graphql(`
          mutation DenyInstanceRoleSpacePermission($input: DenyInstanceRoleSpacePermissionInput!) {
            denyInstanceRoleSpacePermission(input: $input)
          }
        `);
        break;
      case 'neutral':
        mutation = graphql(`
          mutation ClearInstanceRoleSpacePermission($input: ClearInstanceRoleSpacePermissionInput!) {
            clearInstanceRoleSpacePermission(input: $input)
          }
        `);
        break;
    }

    const resp = await connection().client.mutation(mutation, {
      input: { spaceId, instanceRole: instanceRoleName, permission }
    });

    if (resp.error) {
      error = resp.error.message;
    } else {
      // Optimistically update local state
      roleConfig.permissions = roleConfig.permissions.filter((p) => p !== permission);
      roleConfig.permissionDenials = roleConfig.permissionDenials.filter((p) => p !== permission);

      if (newState === 'allow') {
        roleConfig.permissions = [...roleConfig.permissions, permission];
        toast.success(`Granted ${permission}`);
      } else if (newState === 'deny') {
        roleConfig.permissionDenials = [...roleConfig.permissionDenials, permission];
        toast.success(`Denied ${permission}`);
      } else {
        toast.success(`Cleared ${permission}`);
      }
    }

    updating = null;
  }

  function goBack() {
    goto(resolve('/chat/[instanceId]/[spaceId]/admin/roles', { instanceId: instanceIdToSegment(getInstanceId()), spaceId }));
  }
</script>

<PageTitle title={`instance:${roleConfig?.role.displayName ?? instanceRoleName} | Space Admin`} />

<div class="flex min-h-0 min-w-0 flex-1 flex-col">
  <PaneHeader
    title="Instance Role Permissions"
    subtitle={roleConfig ? `instance:${roleConfig.role.displayName}` : 'Loading...'}
    showMobileNav
  >
    {#snippet actions()}
      <Button variant="secondary" onclick={goBack}>Back to Roles</Button>
    {/snippet}
  </PaneHeader>

  <div class="flex flex-col gap-6 overflow-y-auto p-6">
    {#if loading}
      <div class="text-muted">Loading instance role...</div>
    {:else if !roleConfig}
      <div class="text-danger">Instance role not found</div>
    {:else if !canManageRoles}
      <div class="text-danger">
        You need the <code class="rounded bg-surface-200 px-1">admin.manage-roles</code> permission to
        configure instance role permissions.
      </div>
    {:else}
      {#if error}
        <FormError {error} />
      {/if}

      <!-- Instance Role Info Notice -->
      <div class="rounded-lg border border-warning/30 bg-warning/10 p-4">
        <div class="flex items-start gap-3">
          <span class="mt-0.5 iconify text-lg text-warning uil--info-circle"></span>
          <div>
            <div class="font-medium text-warning">Instance Role</div>
            <p class="text-foreground/80 text-sm">
              This is an instance-level role. Role settings (name, description) are managed by
              instance administrators. Here you can only configure space-level permissions for users
              who have this instance role.
            </p>
          </div>
        </div>
      </div>

      <!-- Role Metadata (read-only) -->
      <Panel title="Role Details" icon="iconify uil--info-circle">
        <div class="flex flex-col gap-4">
          <div>
            <div class="mb-1 text-sm font-medium">Instance Role Name</div>
            <code class="rounded bg-surface-200 px-2 py-1">instance:{roleConfig.role.name}</code>
          </div>

          <div>
            <div class="mb-1 text-sm font-medium">Display Name</div>
            <div class="text-foreground">{roleConfig.role.displayName}</div>
          </div>

          <div>
            <div class="mb-1 text-sm font-medium">Description</div>
            <div class="text-muted">{roleConfig.role.description || '(No description)'}</div>
          </div>
        </div>
      </Panel>

      <!-- Space Permissions -->
      <Panel title="Space Permissions" icon="iconify uil--shield-check">
        <p class="mb-4 text-sm text-muted">
          Configure which space permissions users with this instance role should have in this space.
          These permissions are in addition to (or override) permissions from their space roles.
          Changes are saved immediately.
        </p>

        <PermissionGrid
          permissions={allPermissions}
          grantedPermissions={roleConfig.permissions}
          deniedPermissions={roleConfig.permissionDenials}
          disabled={false}
          updatingPermission={updating}
          categoryOrder={['member', 'role', 'space', 'room', 'message']}
          onSetState={setPermissionState}
        />
      </Panel>
    {/if}
  </div>
</div>
