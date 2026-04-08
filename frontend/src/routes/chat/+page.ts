import { redirect } from '@sveltejs/kit';

// /chat is a legacy route — redirect to root which handles all navigation.
export const load = () => {
	redirect(302, '/');
};
