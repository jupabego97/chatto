export const load = ({ url }) => {
  return {
    /** URL to redirect to after login (default: /) */
    redirectUrl: url.searchParams.get('redirect') || '/',

    /** Whether the user just completed a password reset */
    passwordResetSuccess: url.searchParams.get('reset') === 'success'
  };
};
