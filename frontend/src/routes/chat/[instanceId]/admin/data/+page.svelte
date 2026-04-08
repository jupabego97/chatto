<script lang="ts">
  import { goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { graphql } from '$lib/gql';
  import { instanceIdToSegment } from '$lib/navigation';
  import { useQuery } from '$lib/hooks';
  import { getActiveInstance } from '$lib/state/activeInstance.svelte';
  import { Panel, DataTable, formatNumber } from '$lib/components/admin';
  import PaneHeader from '$lib/ui/PaneHeader.svelte';
  import PageTitle from '$lib/ui/PageTitle.svelte';

  const getInstanceId = getActiveInstance();
  const instanceSegment = $derived(instanceIdToSegment(getInstanceId()));

  let { data } = $props();

  // Selected items from URL (reactive via load function)
  const selectedStream = $derived(data.selectedStream);
  const selectedKV = $derived(data.selectedKV);

  // Sorting options
  type SortMode = 'name' | 'messages';
  let subjectSortMode = $state<SortMode>('messages');

  // Main overview query
  const overview = useQuery(
    graphql(`
      query AdminDataOverview {
        admin {
          systemInfo {
            streams {
              name
              messages
              numSubjects
            }
            kvBuckets {
              name
              keys
            }
          }
        }
      }
    `),
    () => ({})
  );

  const streams = $derived(overview.data?.admin?.systemInfo.streams ?? []);
  const kvBuckets = $derived(overview.data?.admin?.systemInfo.kvBuckets ?? []);

  // Detail queries — driven by URL params, skip when not selected
  const streamSubjectsQuery = useQuery(
    graphql(`
      query AdminStreamSubjects($name: String!) {
        admin {
          streamSubjects(name: $name) {
            subject
            messages
          }
        }
      }
    `),
    () => ({ name: selectedStream ?? '' }),
    { skip: () => !selectedStream }
  );

  const kvKeysQuery = useQuery(
    graphql(`
      query AdminKVKeys($name: String!) {
        admin {
          kvKeys(name: $name)
        }
      }
    `),
    () => ({ name: selectedKV ?? '' }),
    { skip: () => !selectedKV }
  );

  const streamSubjects = $derived(streamSubjectsQuery.data?.admin?.streamSubjects ?? []);
  const kvKeys = $derived(kvKeysQuery.data?.admin?.kvKeys ?? []);
  const detailLoading = $derived(streamSubjectsQuery.loading || kvKeysQuery.loading);
  const detailError = $derived(streamSubjectsQuery.error ?? kvKeysQuery.error ?? null);

  const sortedSubjects = $derived.by(() => {
    if (subjectSortMode === 'messages') {
      return [...streamSubjects].sort((a, b) => b.messages - a.messages);
    }
    return [...streamSubjects].sort((a, b) => a.subject.localeCompare(b.subject));
  });

  function selectStream(name: string) {
    const url = resolve('/chat/[instanceId]/admin/data', { instanceId: instanceSegment }) + `?stream=${encodeURIComponent(name)}`;
    // eslint-disable-next-line svelte/no-navigation-without-resolve -- url is constructed with resolve() above
    goto(url, { replaceState: false });
  }

  function selectKV(name: string) {
    const url = resolve('/chat/[instanceId]/admin/data', { instanceId: instanceSegment }) + `?kv=${encodeURIComponent(name)}`;
    // eslint-disable-next-line svelte/no-navigation-without-resolve -- url is constructed with resolve() above
    goto(url, { replaceState: false });
  }

  function clearSelection() {
    goto(resolve('/chat/[instanceId]/admin/data', { instanceId: instanceSegment }), { replaceState: false });
  }
</script>

<PageTitle title="Data | Admin" />

<PaneHeader title="Data" subtitle="Explore stream subjects and KV bucket keys" showMobileNav />

<div class="flex flex-1 gap-6 overflow-hidden p-6">
  {#if overview.loading}
    <div class="text-muted">Loading data...</div>
  {:else if overview.error}
    <div class="rounded-xl border border-danger/20 bg-danger/10 p-4 text-danger">
      {overview.error}
    </div>
  {:else}
    <!-- Left side: Lists -->
    <div class="flex w-80 shrink-0 flex-col gap-4 overflow-auto">
      <!-- Streams -->
      <Panel title="Streams" icon="iconify uil--exchange" count={streams.length} noPadding>
        <div class="flex flex-col">
          {#each streams as stream (stream.name)}
            <button
              type="button"
              class={[
                'flex cursor-pointer items-center justify-between px-4 py-2 text-left transition-colors hover:bg-surface-200',
                selectedStream === stream.name ? 'bg-primary/10 text-primary' : ''
              ]}
              onclick={() => selectStream(stream.name)}
            >
              <span class="truncate font-mono text-sm">{stream.name}</span>
              <span class="ml-2 text-xs text-muted">{formatNumber(stream.numSubjects)}</span>
            </button>
          {:else}
            <div class="px-4 py-3 text-sm text-muted">No streams</div>
          {/each}
        </div>
      </Panel>

      <!-- KV Buckets -->
      <Panel title="KV Buckets" icon="iconify uil--key-skeleton" count={kvBuckets.length} noPadding>
        <div class="flex flex-col">
          {#each kvBuckets as bucket (bucket.name)}
            <button
              type="button"
              class={[
                'flex cursor-pointer items-center justify-between px-4 py-2 text-left transition-colors hover:bg-surface-200',
                selectedKV === bucket.name ? 'bg-primary/10 text-primary' : ''
              ]}
              onclick={() => selectKV(bucket.name)}
            >
              <span class="truncate font-mono text-sm">{bucket.name}</span>
              <span class="ml-2 text-xs text-muted">{formatNumber(bucket.keys)}</span>
            </button>
          {:else}
            <div class="px-4 py-3 text-sm text-muted">No KV buckets</div>
          {/each}
        </div>
      </Panel>
    </div>

    <!-- Right side: Detail view -->
    <div class="flex flex-1 flex-col overflow-hidden">
      {#if selectedStream}
        <Panel
          title="Subjects in {selectedStream}"
          icon="iconify uil--list-ul"
          count={streamSubjects.length}
          noPadding
        >
          {#snippet actions()}
            <button
              type="button"
              class="hover:text-foreground iconify cursor-pointer text-muted uil--times"
              onclick={clearSelection}
              title="Close"
            ></button>
          {/snippet}

          {#if detailLoading}
            <div class="px-4 py-3 text-sm text-muted">Loading subjects...</div>
          {:else if detailError}
            <div class="px-4 py-3 text-sm text-danger">{detailError}</div>
          {:else}
            <DataTable items={sortedSubjects} columns={2} emptyMessage="No subjects">
              {#snippet header()}
                <th class="px-4 py-3 text-left font-medium">
                  <button
                    type="button"
                    class={[
                      'cursor-pointer hover:text-primary',
                      subjectSortMode === 'name' ? 'text-primary' : ''
                    ]}
                    onclick={() => (subjectSortMode = 'name')}
                  >
                    Subject {subjectSortMode === 'name' ? '↑' : ''}
                  </button>
                </th>
                <th class="px-4 py-3 text-right font-medium">
                  <button
                    type="button"
                    class={[
                      'cursor-pointer hover:text-primary',
                      subjectSortMode === 'messages' ? 'text-primary' : ''
                    ]}
                    onclick={() => (subjectSortMode = 'messages')}
                  >
                    {subjectSortMode === 'messages' ? '↓ ' : ''}Messages
                  </button>
                </th>
              {/snippet}
              {#snippet row(subject)}
                <td class="px-4 py-3 font-mono text-sm">{subject.subject}</td>
                <td class="px-4 py-3 text-right font-mono text-sm"
                  >{formatNumber(subject.messages)}</td
                >
              {/snippet}
            </DataTable>
          {/if}
        </Panel>
      {:else if selectedKV}
        <Panel
          title="Keys in {selectedKV}"
          icon="iconify uil--list-ul"
          count={kvKeys.length}
          noPadding
        >
          {#snippet actions()}
            <button
              type="button"
              class="hover:text-foreground iconify cursor-pointer text-muted uil--times"
              onclick={clearSelection}
              title="Close"
            ></button>
          {/snippet}

          {#if detailLoading}
            <div class="px-4 py-3 text-sm text-muted">Loading keys...</div>
          {:else if detailError}
            <div class="px-4 py-3 text-sm text-danger">{detailError}</div>
          {:else}
            <div class="max-h-full overflow-auto">
              {#each kvKeys as key (key)}
                <div class="border-b border-border px-4 py-2 font-mono text-sm last:border-b-0">
                  {key}
                </div>
              {:else}
                <div class="px-4 py-3 text-sm text-muted">No keys</div>
              {/each}
            </div>
          {/if}
        </Panel>
      {:else}
        <div class="flex flex-1 items-center justify-center text-muted">
          Select a stream or KV bucket to view its contents
        </div>
      {/if}
    </div>
  {/if}
</div>
