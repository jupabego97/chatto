<script lang="ts">
  import { resolve } from '$app/paths';
  import { graphql } from '$lib/gql';
  import { instanceIdToSegment } from '$lib/navigation';
  import { useQuery } from '$lib/hooks';
  import { getActiveInstance } from '$lib/state/activeInstance.svelte';
  import { Panel, StatCard, DataTable, formatBytes, formatNumber } from '$lib/components/admin';
  import PaneHeader from '$lib/ui/PaneHeader.svelte';
  import PageTitle from '$lib/ui/PageTitle.svelte';

  const getInstanceId = getActiveInstance();

  const AdminSystemInfoQuery = graphql(`
    query AdminSystemInfo {
      admin {
        systemInfo {
          connection {
            connected
            serverID
            serverName
            version
            maxPayload
            rtt
          }
          account {
            memory
            memoryUsed
            storage
            storageUsed
            streams
            streamsUsed
            consumers
            consumersUsed
          }
          streams {
            name
            messages
            bytes
            consumers
            created
            numSubjects
          }
          kvBuckets {
            name
            keys
            bytes
            history
            ttl
          }
          objectStores {
            name
            size
            sealed
          }
        }
      }
    }
  `);

  const systemQuery = useQuery(AdminSystemInfoQuery, () => ({}));

  let systemInfo = $derived(systemQuery.data?.admin?.systemInfo ?? null);
  let loading = $derived(systemQuery.loading);
  let error = $derived(systemQuery.error ?? null);

  function formatLimit(
    used: number,
    limit: number,
    formatter: (n: number) => string = String
  ): string {
    // NATS uses -1 for unlimited storage/memory, and 0 for unlimited streams/consumers
    return limit <= 0 ? 'unlimited' : formatter(limit);
  }
</script>

<PageTitle title="System | Admin" />

<PaneHeader title="System" subtitle="NATS/JetStream status and metrics" showMobileNav />

<div class="flex flex-col gap-6 overflow-auto p-6">
  {#if loading}
    <div class="text-muted">Loading system information...</div>
  {:else if error}
    <div class="rounded-xl border border-danger/20 bg-danger/10 p-4 text-danger">{error}</div>
  {:else if systemInfo}
    <!-- Connection Info -->
    <Panel title="Connection" icon="iconify uil--plug">
      <div class="grid grid-cols-2 gap-4 md:grid-cols-3 lg:grid-cols-5">
        <div>
          <div class="text-sm text-muted">Status</div>
          <div class="flex items-center gap-2">
            {systemInfo.connection.connected ? 'Connected' : 'Disconnected'}
            <span
              class="h-2 w-2 rounded-full"
              class:bg-success={systemInfo.connection.connected}
              class:bg-danger={!systemInfo.connection.connected}
            ></span>
          </div>
        </div>
        <div>
          <div class="text-sm text-muted">Version</div>
          <div class="font-mono text-sm">{systemInfo.connection.version}</div>
        </div>
        <div>
          <div class="text-sm text-muted">RTT</div>
          <div class="font-mono text-sm">{systemInfo.connection.rtt || '—'}</div>
        </div>
        <div>
          <div class="text-sm text-muted">Max Payload</div>
          <div class="font-mono text-sm">{formatBytes(systemInfo.connection.maxPayload)}</div>
        </div>
        <div>
          <div class="text-sm text-muted">Server ID</div>
          <div class="truncate font-mono text-xs" title={systemInfo.connection.serverID}>
            {systemInfo.connection.serverID.slice(0, 12)}...
          </div>
        </div>
      </div>
    </Panel>

    <!-- Account Usage -->
    <div class="grid grid-cols-2 gap-4 md:grid-cols-4">
      <StatCard
        value={formatBytes(systemInfo.account.storageUsed)}
        label="Storage"
        icon="iconify uil--hdd"
        color="primary"
        subtitle="of {formatLimit(
          systemInfo.account.storageUsed,
          systemInfo.account.storage,
          formatBytes
        )}"
      />
      <StatCard
        value={formatBytes(systemInfo.account.memoryUsed)}
        label="Memory"
        icon="iconify uil--processor"
        color="success"
        subtitle="of {formatLimit(
          systemInfo.account.memoryUsed,
          systemInfo.account.memory,
          formatBytes
        )}"
      />
      <StatCard
        value={systemInfo.account.streamsUsed}
        label="Streams"
        icon="iconify uil--exchange"
        color="warning"
        subtitle="of {formatLimit(systemInfo.account.streamsUsed, systemInfo.account.streams)}"
      />
      <StatCard
        value={systemInfo.account.consumersUsed}
        label="Consumers"
        icon="iconify uil--users-alt"
        color="danger"
        subtitle="of {formatLimit(systemInfo.account.consumersUsed, systemInfo.account.consumers)}"
      />
    </div>

    <!-- Streams -->
    <Panel title="Streams" icon="iconify uil--exchange" count={systemInfo.streams.length} noPadding>
      <DataTable items={systemInfo.streams} columns={6} emptyMessage="No streams">
        {#snippet header()}
          <th class="px-4 py-3 font-medium">Name</th>
          <th class="px-4 py-3 text-right font-medium">Messages</th>
          <th class="px-4 py-3 text-right font-medium">Size</th>
          <th class="px-4 py-3 text-right font-medium">Consumers</th>
          <th class="px-4 py-3 text-right font-medium">Subjects</th>
          <th class="px-4 py-3 font-medium">Created</th>
        {/snippet}
        {#snippet row(stream)}
          {@const streamHref =
            resolve('/chat/[instanceId]/admin/data', { instanceId: instanceIdToSegment(getInstanceId()) }) + `?stream=${encodeURIComponent(stream.name)}`}
          <td class="px-4 py-3 font-mono text-sm">
            <!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- streamHref is constructed with resolve() above -->
            <a href={streamHref} class="cursor-pointer hover:text-primary hover:underline">
              {stream.name}
            </a>
          </td>
          <td class="px-4 py-3 text-right font-mono text-sm">{formatNumber(stream.messages)}</td>
          <td class="px-4 py-3 text-right font-mono text-sm">{formatBytes(stream.bytes)}</td>
          <td class="px-4 py-3 text-right font-mono text-sm">{stream.consumers}</td>
          <td class="px-4 py-3 text-right font-mono text-sm">{stream.numSubjects}</td>
          <td class="px-4 py-3 text-sm text-muted">{stream.created}</td>
        {/snippet}
      </DataTable>
    </Panel>

    <!-- KV Buckets -->
    <Panel
      title="KV Buckets"
      icon="iconify uil--key-skeleton"
      count={systemInfo.kvBuckets.length}
      noPadding
    >
      <DataTable items={systemInfo.kvBuckets} columns={5} emptyMessage="No KV buckets">
        {#snippet header()}
          <th class="px-4 py-3 font-medium">Name</th>
          <th class="px-4 py-3 text-right font-medium">Keys</th>
          <th class="px-4 py-3 text-right font-medium">Size</th>
          <th class="px-4 py-3 text-right font-medium">History</th>
          <th class="px-4 py-3 font-medium">TTL</th>
        {/snippet}
        {#snippet row(bucket)}
          {@const bucketHref =
            resolve('/chat/[instanceId]/admin/data', { instanceId: instanceIdToSegment(getInstanceId()) }) + `?kv=${encodeURIComponent(bucket.name)}`}
          <td class="px-4 py-3 font-mono text-sm">
            <!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- bucketHref is constructed with resolve() above -->
            <a href={bucketHref} class="cursor-pointer hover:text-primary hover:underline">
              {bucket.name}
            </a>
          </td>
          <td class="px-4 py-3 text-right font-mono text-sm">{formatNumber(bucket.keys)}</td>
          <td class="px-4 py-3 text-right font-mono text-sm">{formatBytes(bucket.bytes)}</td>
          <td class="px-4 py-3 text-right font-mono text-sm">{bucket.history}</td>
          <td class="px-4 py-3 text-sm text-muted">{bucket.ttl}</td>
        {/snippet}
      </DataTable>
    </Panel>

    <!-- Object Stores -->
    <Panel
      title="Object Stores"
      icon="iconify uil--box"
      count={systemInfo.objectStores.length}
      noPadding
    >
      <DataTable items={systemInfo.objectStores} columns={3} emptyMessage="No object stores">
        {#snippet header()}
          <th class="px-4 py-3 font-medium">Name</th>
          <th class="px-4 py-3 text-right font-medium">Size</th>
          <th class="px-4 py-3 font-medium">Sealed</th>
        {/snippet}
        {#snippet row(store)}
          <td class="px-4 py-3 font-mono text-sm">{store.name}</td>
          <td class="px-4 py-3 text-right font-mono text-sm">{formatBytes(store.size)}</td>
          <td class="px-4 py-3 text-sm">{store.sealed ? 'Yes' : 'No'}</td>
        {/snippet}
      </DataTable>
    </Panel>
  {/if}
</div>
