import type { PageLoad } from './$types';

export const load: PageLoad = async ({ parent, url }) => {
  const { user } = await parent();
  return {
    token: url.searchParams.get('token') ?? '',
    user
  };
};
