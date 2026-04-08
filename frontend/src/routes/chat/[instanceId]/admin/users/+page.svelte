<script lang="ts">
  import { graphql } from '$lib/gql';
  import { useQuery } from '$lib/hooks';
  import { getAdminPermissions } from '$lib/state/instance/permissions.svelte';
  import { UserList } from '$lib/components/admin';
  import PaneHeader from '$lib/ui/PaneHeader.svelte';
  import PageTitle from '$lib/ui/PageTitle.svelte';

  const AdminUsersListQuery = graphql(`
    query AdminUsersList {
      users {
        id
        login
        displayName
        hasVerifiedEmail
        verifiedEmails
      }
    }
  `);

  const adminPerms = getAdminPermissions();
  const canManageUsers = $derived(adminPerms.hasPermission('admin.manage-users'));

  const usersQuery = useQuery(AdminUsersListQuery, () => ({}));

  let users = $derived(usersQuery.data?.users ?? []);
  let loading = $derived(usersQuery.loading);
</script>

<PageTitle title="Users | Admin" />

<PaneHeader title="Users" subtitle="Manage registered users on this instance" showMobileNav />

<div class="flex flex-col gap-6 overflow-y-auto p-6">
  <UserList {users} {loading} clickable={canManageUsers} />
</div>
