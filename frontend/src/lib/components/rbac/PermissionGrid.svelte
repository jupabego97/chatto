<script lang="ts">
  import { getPermissionDescription } from '$lib/permissions';

  type PermissionState = 'allow' | 'deny' | 'neutral';

  // Default category order - can be overridden via prop
  const DEFAULT_CATEGORY_ORDER = [
    'space',
    'room',
    'message',
    'member',
    'role',
    'admin',
    'dm',
    'user'
  ];

  let {
    permissions,
    grantedPermissions,
    deniedPermissions = [],
    disabled = false,
    updatingPermission = null,
    categoryOrder = DEFAULT_CATEGORY_ORDER,
    onSetState
  }: {
    permissions: string[];
    grantedPermissions: string[];
    deniedPermissions?: string[];
    disabled?: boolean;
    updatingPermission?: string | null;
    categoryOrder?: string[];
    onSetState: (permission: string, state: PermissionState) => void;
  } = $props();

  // Category metadata with display info
  const categoryMeta: Record<string, { title: string; description: string }> = {
    space: {
      title: 'Space Operations',
      description: 'Control who can browse, create, join, and manage spaces'
    },
    room: {
      title: 'Room Operations',
      description: 'Control who can create, join, and manage rooms'
    },
    message: {
      title: 'Messages',
      description: 'Control what users can do with messages'
    },
    member: {
      title: 'Member Management',
      description: 'Control who can invite and remove space members'
    },
    role: {
      title: 'Role Management',
      description: 'Control who can create roles and assign them to users'
    },
    admin: {
      title: 'Instance Administration',
      description: 'Access to instance-wide admin functions'
    },
    dm: {
      title: 'Direct Messages',
      description: 'Control access to direct messaging'
    },
    user: {
      title: 'User Management',
      description: 'Control user account operations'
    }
  };

  // Extract category from permission ID (e.g., "message.delete.any" -> "message")
  function getCategory(permission: string): string {
    const dotIndex = permission.indexOf('.');
    return dotIndex > 0 ? permission.slice(0, dotIndex) : permission;
  }

  // Group permissions by category
  const groupedPermissions = $derived.by(() => {
    // eslint-disable-next-line svelte/prefer-svelte-reactivity -- Map is ephemeral within derived computation
    const groups = new Map<string, string[]>();

    for (const perm of permissions) {
      const category = getCategory(perm);
      if (!groups.has(category)) {
        groups.set(category, []);
      }
      groups.get(category)!.push(perm);
    }

    // Sort permissions within each group
    for (const perms of groups.values()) {
      perms.sort((a, b) => a.localeCompare(b));
    }

    // Return as ordered array of [category, permissions] pairs
    const result: Array<{ category: string; permissions: string[] }> = [];
    for (const category of categoryOrder) {
      const perms = groups.get(category);
      if (perms && perms.length > 0) {
        result.push({ category, permissions: perms });
      }
    }

    // Add any categories not in the predefined order
    for (const [category, perms] of groups) {
      if (!categoryOrder.includes(category) && perms.length > 0) {
        result.push({ category, permissions: perms });
      }
    }

    return result;
  });

  function getPermissionState(id: string): PermissionState {
    if (grantedPermissions.includes(id)) return 'allow';
    if (deniedPermissions.includes(id)) return 'deny';
    return 'neutral';
  }
</script>

<div class="flex flex-col gap-6">
  {#each groupedPermissions as group (group.category)}
    {@const meta = categoryMeta[group.category]}
    <div class="flex flex-col gap-2">
      <!-- Category header -->
      <div class="mb-1">
        <h3 class="text-sm font-semibold">{meta?.title ?? group.category}</h3>
        {#if meta?.description}
          <p class="text-xs text-muted">{meta.description}</p>
        {/if}
      </div>

      <!-- Permissions in this category -->
      {#each group.permissions as permission (permission)}
        {@const state = getPermissionState(permission)}
        {@const isUpdating = updatingPermission === permission}
        {@const isDisabled = disabled || isUpdating}

        <div
          class={[
            'flex items-center gap-4 rounded-lg border p-3',
            state === 'allow'
              ? 'border-success/50 bg-success/5'
              : state === 'deny'
                ? 'border-danger/50 bg-danger/5'
                : 'border-border',
            isDisabled ? 'opacity-50' : '',
            isUpdating ? 'animate-pulse' : ''
          ]}
        >
          <!-- Permission name and description -->
          <div class="min-w-48 flex-1">
            <code
              class={[
                'text-sm',
                state === 'allow' ? 'text-success' : state === 'deny' ? 'text-danger' : ''
              ]}
            >
              {permission}
            </code>
            <div class="text-xs text-muted">{getPermissionDescription(permission)}</div>
            <div class="text-xs text-muted/70">
              {#if state === 'allow'}
                Granted
              {:else if state === 'deny'}
                Denied (overrides grants from other roles)
              {:else}
                Neutral (no effect)
              {/if}
            </div>
          </div>

          <!-- Allow checkbox -->
          <label
            class={[
              'flex items-center gap-1.5 text-sm',
              isDisabled ? 'cursor-not-allowed' : 'cursor-pointer'
            ]}
          >
            <input
              type="checkbox"
              checked={state === 'allow'}
              disabled={isDisabled}
              class="accent-success"
              onchange={() => onSetState(permission, state === 'allow' ? 'neutral' : 'allow')}
            />
            <span class="text-success">Allow</span>
          </label>

          <!-- Deny checkbox -->
          <label
            class={[
              'flex items-center gap-1.5 text-sm',
              isDisabled ? 'cursor-not-allowed' : 'cursor-pointer'
            ]}
          >
            <input
              type="checkbox"
              checked={state === 'deny'}
              disabled={isDisabled}
              class="accent-danger"
              onchange={() => onSetState(permission, state === 'deny' ? 'neutral' : 'deny')}
            />
            <span class="text-danger">Deny</span>
          </label>
        </div>
      {/each}
    </div>
  {/each}
</div>
