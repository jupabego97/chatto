<script lang="ts">
  import { goto, invalidateAll } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { clearCachedUser } from '$lib/auth/loadAuth';
  import AuthLayout from '$lib/components/AuthLayout.svelte';
  import { graphql } from '$lib/gql';
  import { instanceRegistry } from '$lib/state/instance/registry.svelte';
  import { graphqlClientManager } from '$lib/state/instance/graphqlClient.svelte';
  import { Divider } from '$lib/ui';
  import PageTitle from '$lib/ui/PageTitle.svelte';
  import { TextInput, FormError, Button } from '$lib/ui/form';

  const { data } = $props();

  let identifier = $state('');
  let password = $state('');
  let error = $state('');
  let isLoading = $state(false);

  const canSubmit = $derived(identifier.trim() && password);

  // Standalone detection: no origin instance means no local backend to log in to.
  // Only applies when there's no redirect param — a redirect means the backend sent
  // us here (e.g. OAuth authorize flow), so the origin probe just hasn't completed yet.
  const isStandalone = $derived(
    !instanceRegistry.originInstance &&
    instanceRegistry.originProbed &&
    data.redirectUrl === '/'
  );

  $effect(() => {
    if (data.user) {
      navigateAfterLogin(data.redirectUrl);
    }
  });

  // Fetch enabled auth providers and registration setting from GraphQL
  const LoginInfoQuery = graphql(`
    query LoginPageInfo {
      instance {
        enabledAuthProviders
        directRegistrationEnabled
      }
    }
  `);

  let enabledProviders = $state<string[]>([]);
  let directRegistrationEnabled = $state(true);

  graphqlClientManager.originClient.client
    .query(LoginInfoQuery, {})
    .toPromise()
    .then((result) => {
      if (result.data) {
        enabledProviders = result.data.instance.enabledAuthProviders;
        directRegistrationEnabled = result.data.instance.directRegistrationEnabled;
      }
    });

  /**
   * Navigate after a successful login. Uses `window.location.href` for backend
   * routes (e.g. `/oauth/authorize`) that are served by Gin, not SvelteKit.
   */
  function navigateAfterLogin(url: string) {
    if (url.startsWith('/oauth/')) {
      window.location.href = url;
    } else {
      // eslint-disable-next-line svelte/no-navigation-without-resolve -- url is either a returnUrl from sessionStorage (already resolved) or a backend redirect path
      goto(url);
    }
  }

  async function handleSubmit(e: Event) {
    e.preventDefault();
    error = '';
    isLoading = true;

    try {
      const response = await fetch('/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ identifier, password }),
        credentials: 'include'
      });

      const result = await response.json();

      if (!response.ok) {
        error = result.error || 'Login failed';
        return;
      }

      clearCachedUser();
      await invalidateAll();

      const returnUrl = sessionStorage.getItem('returnUrl');
      if (returnUrl) {
        sessionStorage.removeItem('returnUrl');
        navigateAfterLogin(returnUrl);
      } else {
        navigateAfterLogin(data.redirectUrl);
      }
    } catch (err) {
      error = err instanceof Error ? err.message : 'Login failed';
    } finally {
      isLoading = false;
    }
  }
</script>

<PageTitle title={isStandalone ? "Welcome" : "Sign In"} />

{#if !data.user}
  {#if isStandalone}
    <AuthLayout>
      <div class="flex flex-col items-center gap-6 text-center">
        <h1 class="text-2xl font-bold">Welcome to Chatto</h1>
        <p class="text-muted">
          Connect to a Chatto instance to get started. You can add multiple instances and switch between them.
        </p>
        <a href={resolve('/instances/add')} class="btn-primary btn-lg block w-full cursor-pointer text-center">
          Add Instance
        </a>
      </div>
    </AuthLayout>
  {:else}
    <AuthLayout>
      <h1 class="mb-6 text-center text-2xl font-bold">Sign In</h1>

      {#if data.passwordResetSuccess}
        <div
          class="mb-4 rounded-lg bg-green-100 p-3 text-center text-sm text-green-800 dark:bg-green-900/30 dark:text-green-200"
        >
          Password reset successful! Please sign in with your new password.
        </div>
      {/if}

      <!-- SSO providers -->
      {#if enabledProviders.includes('oidc')}
        <div class="flex flex-col gap-3">
          <!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- /auth/oidc is a backend route, not a SvelteKit route -->
          <a href="/auth/oidc?redirect={encodeURIComponent(data.redirectUrl)}" class="flex cursor-pointer items-center justify-center gap-3 rounded-lg border border-border bg-white px-4 py-3 font-medium text-gray-700 transition-colors hover:bg-gray-50 dark:bg-gray-800 dark:text-gray-200 dark:hover:bg-gray-700">
            <span class="iconify text-lg mdi--shield-account"></span>
            <span>Continue with Chatto Hub</span>
          </a>

          <Divider label="or" />
        </div>
      {/if}

      <form onsubmit={handleSubmit} class="flex flex-col gap-4">
        <TextInput
          id="identifier"
          label="Username or Email"
          bind:value={identifier}
          placeholder="you@example.com"
          disabled={isLoading}
          required
          autocomplete="username"
        />

        <TextInput
          id="password"
          label="Password"
          type="password"
          bind:value={password}
          placeholder="Your password"
          disabled={isLoading}
          required
          autocomplete="current-password"
        />

        <FormError {error} />

        <Button type="submit" size="lg" disabled={!canSubmit} loading={isLoading} loadingText="Signing in...">
          Sign In
        </Button>
      </form>

      <div class="mt-4 text-center">
        <a href={resolve('/forgot-password')} class="text-muted hover:text-primary hover:underline">
          Forgot password?
        </a>
      </div>

      {#if directRegistrationEnabled}
        <Divider label="or" />

        <a href={resolve('/register')} class="btn-secondary btn-lg block w-full text-center">
          Create Account
        </a>
      {/if}
    </AuthLayout>
  {/if}
{/if}
