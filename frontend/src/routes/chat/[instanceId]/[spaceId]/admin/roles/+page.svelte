<script lang="ts">
  import { goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import { instanceIdToSegment } from '$lib/navigation';
  import { getActiveInstance } from '$lib/state/activeInstance.svelte';
  import { graphql } from '$lib/gql';
  import { useQuery, useMutation } from '$lib/hooks';
  import type { InstanceRoleSpaceConfig } from '$lib/gql/graphql';
  import { Panel } from '$lib/components/admin';
  import PaneHeader from '$lib/ui/PaneHeader.svelte';
  import PageTitle from '$lib/ui/PageTitle.svelte';
  import { Button } from '$lib/ui/form';
  import { toast } from '$lib/ui/toast';
  import { RoleTable, type Role } from '$lib/components/rbac';

  const SpaceRolesQuery = graphql(`
    query SpaceRoles($spaceId: ID!) {
      space(id: $spaceId) {
        id
        name
        roles {
          name
          displayName
          description
          permissions
          permissionDenials
          isSystem
          position
        }
        viewerCanManageRoles
        instanceRoleConfigs {
          role {
            name
            displayName
            description
            position
            isSystem
          }
          permissions
          permissionDenials
        }
      }
    }
  `);

  const ReorderSpaceRolesMutation = graphql(`
    mutation ReorderSpaceRoles($input: ReorderSpaceRolesInput!) {
      reorderSpaceRoles(input: $input) {
        name
        displayName
        description
        permissions
        permissionDenials
        isSystem
        position
      }
    }
  `);

  const getInstanceId = getActiveInstance();
  const instanceSegment = $derived(instanceIdToSegment(getInstanceId()));
  const spaceId = $derived(page.params.spaceId!);

  // Local state for optimistic updates after reorder
  let localRoles = $state<Role[] | null>(null);

  const rolesQuery = useQuery(SpaceRolesQuery, () => ({ spaceId }));
  const reorderMutation = useMutation(ReorderSpaceRolesMutation);

  let roles = $derived(localRoles ?? rolesQuery.data?.space?.roles ?? []);
  let instanceRoleConfigs = $derived(
    (rolesQuery.data?.space?.instanceRoleConfigs ?? []) as InstanceRoleSpaceConfig[]
  );
  let canManageRoles = $derived(rolesQuery.data?.space?.viewerCanManageRoles ?? false);
  let loading = $derived(rolesQuery.loading);
  let error = $derived(
    rolesQuery.error ?? (!rolesQuery.loading && !rolesQuery.data?.space ? 'Space not found' : null)
  );
  let reordering = $derived(reorderMutation.loading);

  function handleEditRole(role: Role) {
    goto(resolve('/chat/[instanceId]/[spaceId]/admin/roles/[name]', { instanceId: instanceSegment, spaceId, name: role.name }));
  }

  function handleEditInstanceRole(config: InstanceRoleSpaceConfig) {
    goto(resolve('/chat/[instanceId]/[spaceId]/admin/roles/instance/[name]', { instanceId: instanceSegment, spaceId, name: config.role.name }));
  }

  function goToNewRole() {
    goto(resolve('/chat/[instanceId]/[spaceId]/admin/roles/new', { instanceId: instanceSegment, spaceId }));
  }

  // Count how many space permissions are configured for instance roles
  function getConfiguredCount(config: InstanceRoleSpaceConfig): number {
    return config.permissions.length + config.permissionDenials.length;
  }

  async function handleReorder(roleNames: string[]) {
    if (reordering) return;

    const result = await reorderMutation.execute({ input: { spaceId, roleNames } });

    if (result.error) {
      toast.error(`Failed to reorder roles: ${result.error}`);
      // Reset local state to trigger refetch from query
      localRoles = null;
      rolesQuery.refetch();
    } else if (result.data?.reorderSpaceRoles) {
      // Update local state with reordered roles
      localRoles = result.data.reorderSpaceRoles;
      toast.success('Role order updated');
    }
  }
</script>

<PageTitle title="Roles | Space Admin" />

<div class="flex min-h-0 min-w-0 flex-1 flex-col">
  <PaneHeader title="Roles" subtitle="Manage space roles and permissions" showMobileNav>
    {#snippet actions()}
      {#if canManageRoles}
        <Button variant="primary" onclick={goToNewRole}>Create Role</Button>
      {/if}
    {/snippet}
  </PaneHeader>

  <div class="flex flex-col gap-6 overflow-y-auto p-6">
    {#if loading}
      <div class="text-muted">Loading roles...</div>
    {:else if error}
      <div class="text-danger">{error}</div>
    {:else}
      <Panel title="Space Roles" icon="iconify uil--shield-check">
        <p class="mb-4 text-sm text-muted">
          {#if canManageRoles}
            Manage roles and their permissions. Drag custom roles to reorder them. System roles
            (admin, member) cannot be deleted or reordered.
          {:else}
            View roles and their permissions. You need the
            <code class="rounded bg-surface-200 px-1">roles.manage</code> permission to make changes.
          {/if}
        </p>

        <RoleTable
          {roles}
          canManage={canManageRoles}
          onEdit={canManageRoles ? handleEditRole : undefined}
          onReorder={canManageRoles ? handleReorder : undefined}
        />
      </Panel>

      {#if canManageRoles && instanceRoleConfigs.length > 0}
        <Panel title="Instance Roles" icon="iconify uil--globe">
          <p class="mb-4 text-sm text-muted">
            Configure space-level permissions for users based on their instance roles. Instance
            roles are defined at the instance level and cannot be modified here.
          </p>

          <div class="overflow-x-auto">
            <table class="w-full">
              <thead>
                <tr class="border-b border-surface-300 text-left text-sm text-muted">
                  <th class="pr-4 pb-2 font-medium">Role</th>
                  <th class="pr-4 pb-2 font-medium">Display Name</th>
                  <th class="pr-4 pb-2 font-medium">Configured</th>
                  <th class="pb-2 font-medium"></th>
                </tr>
              </thead>
              <tbody class="divide-y divide-surface-200">
                {#each instanceRoleConfigs as config (config.role.name)}
                  <tr class="group">
                    <td class="py-3 pr-4">
                      <div class="flex items-center gap-2">
                        <span
                          class="rounded bg-accent/10 px-1.5 py-0.5 text-xs font-medium text-accent"
                        >
                          instance:
                        </span>
                        <code class="text-sm">{config.role.name}</code>
                      </div>
                    </td>
                    <td class="py-3 pr-4 text-sm">{config.role.displayName}</td>
                    <td class="py-3 pr-4 text-sm">
                      {#if getConfiguredCount(config) > 0}
                        <span class="text-muted">
                          {getConfiguredCount(config)} permission{getConfiguredCount(config) !== 1
                            ? 's'
                            : ''}
                        </span>
                      {:else}
                        <span class="text-muted/50">Not configured</span>
                      {/if}
                    </td>
                    <td class="py-3 text-right">
                      <Button
                        variant="ghost"
                        size="sm"
                        onclick={() => handleEditInstanceRole(config)}
                      >
                        Configure
                      </Button>
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        </Panel>
      {/if}
    {/if}
  </div>
</div>
