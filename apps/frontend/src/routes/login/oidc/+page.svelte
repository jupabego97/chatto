<script lang="ts">
  import { goto, invalidateAll } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { onMount } from 'svelte';
  import { csrfFetch } from '$lib/auth/csrf';
  import AuthLayout from '$lib/components/AuthLayout.svelte';
  import * as m from '$lib/i18n/messages';
  import {
    serverRegistry,
    type AuthenticatedUserSummary
  } from '$lib/state/server/registry.svelte';
  import Hint from '$lib/ui/Hint.svelte';
  import PageTitle from '$lib/ui/PageTitle.svelte';
  import { TextInput, FormError, Button } from '$lib/ui/form';

  type PendingOIDC = {
    providerId: string;
    providerLabel: string;
    email?: string;
    emailVerified?: boolean;
    name?: string;
    username?: string;
    mode?: string;
    canLinkCurrent?: boolean;
  };

  const { data } = $props();
  const token = $derived(data.token);

  let pending = $state<PendingOIDC | null>(null);
  let loading = $state(true);
  let submitting = $state<'create' | 'link-current' | 'link-password' | 'cancel' | null>(null);
  let error = $state('');
  let identifier = $state('');
  let password = $state('');

  const displayIdentity = $derived(
    pending?.email || pending?.username || pending?.name || pending?.providerLabel || ''
  );
  const canLinkWithPassword = $derived(identifier.trim() !== '' && password !== '');

  onMount(async () => {
    if (!token) {
      loading = false;
      error = m['auth.oidc.pending_missing']();
      return;
    }
    try {
      const originToken = serverRegistry.originServer?.token;
      const response = await fetch(`/auth/pending-oidc/${encodeURIComponent(token)}`, {
        credentials: 'include',
        headers: originToken ? { Authorization: `Bearer ${originToken}` } : undefined,
        signal: AbortSignal.timeout(10000)
      });
      const result = await response.json();
      if (!response.ok) {
        error = result.error || m['auth.oidc.pending_missing']();
        return;
      }
      pending = result;
    } catch (err) {
      error = err instanceof Error ? err.message : m['auth.oidc.load_failed']();
    } finally {
      loading = false;
    }
  });

  async function authenticateOrigin(
    token: string,
    user: AuthenticatedUserSummary | null
  ): Promise<void> {
    const [{ serverRegistry }, { clearCachedUser }] = await Promise.all([
      import('$lib/state/server/registry.svelte'),
      import('$lib/auth/loadAuth')
    ]);
    serverRegistry.authenticateOrigin(token, user);
    clearCachedUser();
  }

  async function followAuthRedirect(redirectUrl: string) {
    const target = new URL(redirectUrl || '/', window.location.origin);
    if (target.origin !== window.location.origin) {
      window.location.href = target.href;
      return;
    }
    const bearerToken = target.searchParams.get('token');
    if (bearerToken) {
      target.searchParams.delete('token');
      await authenticateOrigin(bearerToken, null);
      await invalidateAll();
    }
    const path = target.pathname + target.search + target.hash;
    if (path.startsWith('/oauth/')) {
      window.location.href = path;
    } else {
      // eslint-disable-next-line svelte/no-navigation-without-resolve -- backend returns a validated same-origin path after auth completion
      goto(path || resolve('/'), { replaceState: true });
    }
  }

  async function postPendingAction(
    action: 'create' | 'link-current' | 'link-password' | 'cancel',
    body?: Record<string, string>
  ) {
    if (!token || submitting) return;
    submitting = action;
    error = '';
    try {
      const headers: Record<string, string> = {
        Accept: 'application/json',
        ...(body ? { 'Content-Type': 'application/json' } : {})
      };
      const originToken = action === 'link-current' ? serverRegistry.originServer?.token : undefined;
      if (originToken) {
        headers.Authorization = `Bearer ${originToken}`;
      }

      const response = await csrfFetch(
        `/auth/pending-oidc/${encodeURIComponent(token)}/${action}`,
        {
          method: 'POST',
          credentials: 'include',
          headers,
          body: body ? JSON.stringify(body) : undefined,
          signal: AbortSignal.timeout(10000)
        }
      );

      const contentType = response.headers.get('content-type') ?? '';
      const result = contentType.includes('application/json') ? await response.json() : {};
      if (!response.ok) {
        error = result.error || m['auth.oidc.submit_failed']();
        return;
      }

      if (action === 'cancel') {
        goto(resolve('/login'), { replaceState: true });
        return;
      }
      if (result.redirectUrl) {
        await followAuthRedirect(result.redirectUrl);
      } else if (response.redirected) {
        window.location.href = response.url;
      } else {
        error = m['auth.oidc.missing_redirect']();
      }
    } catch (err) {
      if (err instanceof DOMException && err.name === 'AbortError') {
        error = m['auth.oidc.submit_timeout']();
      } else {
        error = err instanceof Error ? err.message : m['auth.oidc.submit_failed']();
      }
    } finally {
      submitting = null;
    }
  }

  function createAccount() {
    void postPendingAction('create');
  }

  function linkCurrentAccount() {
    void postPendingAction('link-current');
  }

  function linkWithPassword(e: Event) {
    e.preventDefault();
    if (!canLinkWithPassword) return;
    void postPendingAction('link-password', { identifier, password });
  }

  function cancel() {
    void postPendingAction('cancel');
  }
</script>

<PageTitle title={m['auth.oidc.title']()} />

<AuthLayout>
  <div class="flex flex-col gap-6">
    <div class="text-center">
      <div
        class="mb-4 inline-flex h-12 w-12 items-center justify-center rounded-full bg-accent/10 text-accent"
      >
        <span class="iconify text-2xl mdi--shield-account"></span>
      </div>
      <h1 class="text-2xl font-bold">{m['auth.oidc.heading']()}</h1>
    </div>

    {#if loading}
      <div class="flex justify-center py-8">
        <span class="iconify animate-spin text-3xl text-muted mdi--loading"></span>
      </div>
    {:else if pending}
      <div class="flex flex-col gap-5">
        <Hint>
          {m['auth.oidc.unlinked_notice']({ provider: pending.providerLabel })}
        </Hint>

        {#if displayIdentity}
          <div class="surface-box p-4 text-sm">
            <div class="text-muted">{m['auth.oidc.identity_label']()}</div>
            <div class="mt-1 font-medium">{displayIdentity}</div>
          </div>
        {/if}

        {#if data.user && pending.canLinkCurrent}
          <div class="flex flex-col gap-3">
            <Button
              size="lg"
              fullWidth
              loading={submitting === 'link-current'}
              loadingText={m['auth.oidc.linking']()}
              disabled={submitting !== null}
              onclick={linkCurrentAccount}
            >
              <span class="iconify mdi--link-variant"></span>
              {m['auth.oidc.link_current']()}
            </Button>
            <Button
              variant="secondary"
              size="lg"
              fullWidth
              loading={submitting === 'create'}
              loadingText={m['auth.oidc.creating']()}
              disabled={submitting !== null}
              onclick={createAccount}
            >
              <span class="iconify mdi--account-plus"></span>
              {m['auth.oidc.create_new']()}
            </Button>
          </div>
        {:else}
          <div class="flex flex-col gap-3">
            <Button
              size="lg"
              fullWidth
              loading={submitting === 'create'}
              loadingText={m['auth.oidc.creating']()}
              disabled={submitting !== null}
              onclick={createAccount}
            >
              <span class="iconify mdi--account-plus"></span>
              {m['auth.oidc.create_new']()}
            </Button>

            <form class="flex flex-col gap-3" onsubmit={linkWithPassword}>
              <TextInput
                id="identifier"
                label={m['auth.login.identifier_label']()}
                bind:value={identifier}
                placeholder={m['common.email_placeholder']()}
                disabled={submitting !== null}
                autocomplete="username"
              />
              <TextInput
                id="password"
                label={m['common.password']()}
                type="password"
                bind:value={password}
                placeholder={m['common.password_placeholder']()}
                disabled={submitting !== null}
                autocomplete="current-password"
              />
              <Button
                type="submit"
                variant="secondary"
                size="lg"
                fullWidth
                loading={submitting === 'link-password'}
                loadingText={m['auth.oidc.linking']()}
                disabled={!canLinkWithPassword || submitting !== null}
              >
                <span class="iconify mdi--link-variant"></span>
                {m['auth.oidc.link_existing']()}
              </Button>
            </form>
          </div>
        {/if}

        <FormError {error} />

        <Button
          variant="ghost"
          fullWidth
          loading={submitting === 'cancel'}
          disabled={submitting !== null}
          onclick={cancel}
        >
          {m['common.cancel']()}
        </Button>
      </div>
    {:else}
      <div class="flex flex-col gap-4 text-center">
        <FormError {error} />
        <Button variant="secondary" size="lg" fullWidth onclick={() => goto(resolve('/login'))}>
          {m['auth.oidc.return_login']()}
        </Button>
      </div>
    {/if}
  </div>
</AuthLayout>
