<script lang="ts">
  import { goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { graphql } from '$lib/gql';
  import { instanceIdToSegment } from '$lib/navigation';
  import { useMutation } from '$lib/hooks';
  import { getAdminPermissions } from '$lib/state/instance/permissions.svelte';
  import { getActiveInstance } from '$lib/state/activeInstance.svelte';
  import { Panel } from '$lib/components/admin';
  import PaneHeader from '$lib/ui/PaneHeader.svelte';
  import PageTitle from '$lib/ui/PageTitle.svelte';
  import { Button, FormError } from '$lib/ui/form';
  import { RoleForm } from '$lib/components/rbac';

  const getInstanceId = getActiveInstance();

  // Get permissions context from layout
  const adminPerms = getAdminPermissions();
  const canManage = $derived(adminPerms.hasPermission('admin.manage-roles'));

  let name = $state('');
  let displayName = $state('');
  let description = $state('');
  let error = $state<string | null>(null);

  const createRoleMutation = useMutation(
    graphql(`
      mutation CreateRole($input: CreateRoleInput!) {
        createRole(input: $input) {
          name
        }
      }
    `)
  );

  async function createRole() {
    error = null;

    const result = await createRoleMutation.execute({
      input: { name, displayName, description }
    });

    if (result.error) {
      error = result.error;
    } else {
      goto(resolve('/chat/[instanceId]/admin/roles/[name]', { instanceId: instanceIdToSegment(getInstanceId()), name }));
    }
  }
</script>

<PageTitle title="Create Role | Admin" />

<PaneHeader title="Create Role" subtitle="Create a new custom instance role" showMobileNav>
  {#snippet actions()}
    <Button variant="secondary" href={resolve('/chat/[instanceId]/admin/roles', { instanceId: instanceIdToSegment(getInstanceId()) })}>Cancel</Button>
  {/snippet}
</PaneHeader>

<div class="flex flex-col gap-6 overflow-y-auto p-6">
  {#if !canManage}
    <div class="text-danger">
      You need the <code class="rounded bg-surface-200 px-1">admin.manage-roles</code> permission to create
      roles.
    </div>
  {:else}
    {#if error}
      <FormError {error} />
    {/if}

    <Panel title="Role Details" icon="iconify uil--plus-circle">
      <RoleForm
        bind:name
        bind:displayName
        bind:description
        isInstanceRole={true}
        saving={createRoleMutation.loading}
        submitLabel="Create Role"
        savingLabel="Creating..."
        onSubmit={createRole}
      />
      <p class="mt-4 text-sm text-muted">
        After creating the role, you can assign permissions to it on the edit page.
      </p>
    </Panel>
  {/if}
</div>
