import { graphql } from '$lib/gql';
import { graphqlClientManager } from '$lib/state/server/graphqlClient.svelte';
import { goto } from '$app/navigation';
import { serverIdToSegment } from '$lib/navigation';
import { roomPathForSegment } from '$lib/roomUrls';

const StartDMMutation = graphql(`
  mutation StartDM($input: StartDMInput!) {
    startDM(input: $input) {
      id
    }
  }
`);

/**
 * Start a DM conversation with a user and navigate to it.
 */
export async function startDMWith(serverId: string, userId: string): Promise<void> {
  const result = await graphqlClientManager
    .getClient(serverId)
    .client.mutation(StartDMMutation, { input: { participantIds: [userId] } })
    .toPromise();

  if (result.data?.startDM) {
    goto(roomPathForSegment(serverIdToSegment(serverId), result.data.startDM.id));
  }
}
