import { redirect } from '@sveltejs/kit';
import { resolve } from '$app/paths';
import { browser } from '$app/environment';
import { graphqlClientManager } from '$lib/state/instance/graphqlClient.svelte';
import { graphql } from '$lib/gql';
import type { PageLoad } from './$types';

const InstanceNeedsSetupQuery = graphql(`
  query InstanceNeedsSetup {
    instance {
      needsSetup
    }
  }
`);

export const load: PageLoad = async ({ url }) => {
  if (!browser) {
    return {};
  }

  // Check if instance needs initial setup
  const result = await graphqlClientManager.originClient.client.query(InstanceNeedsSetupQuery, {}).toPromise();

  if (result.data?.instance.needsSetup) {
    redirect(302, resolve('/setup'));
  }

  // Pass through welcome query param if present
  return {
    welcome: url.searchParams.get('welcome') === 'true'
  };
};
