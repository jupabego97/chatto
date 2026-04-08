export const load = ({ params }) => {
	return {
		/** The currently active DM conversation (if viewing one). */
		conversationId: params.conversationId as string | undefined
	};
};
