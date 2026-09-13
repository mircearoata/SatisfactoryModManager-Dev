<script lang="ts">
  import { getContextClient } from '@urql/svelte';
  import Fuse from 'fuse.js';
  import _ from 'lodash';
  import { createEventDispatcher } from 'svelte';

  import ModListFilters from './ModsListFilters.svelte';
  import ModsListItem from './ModsListItem.svelte';

  import T from '$lib/components/T.svelte';
  import VirtualList from '$lib/components/VirtualList.svelte';
  import AnnouncementsBar from '$lib/components/announcements/AnnouncementsBar.svelte';
  import { GetHiddenModDocument, GetModCountDocument, GetModsDocument } from '$lib/generated';
  import { lockfileMods, manifestMods } from '$lib/store/ficsitCLIStore';
  import { expandedMod, hasFetchedMods } from '$lib/store/generalStore';
  import {
    type OfflineMod, type PartialMod, type PartialSMRMod, filter, filterOptions, order, orderByOptions, search,
    selectedTags,
  } from '$lib/store/modFiltersStore';
  import { offline, startView, tagSearchMode } from '$lib/store/settingsStore';
  import { OfflineGetMods } from '$wailsjs/go/ficsitcli/ficsitCLI';
  import { settings } from '$wailsjs/go/models';

  const dispatch = createEventDispatcher();

  const MODS_PER_PAGE = 100;

  const client = getContextClient();

  let fetchingMods = false;
  let onlineMods: PartialSMRMod[] = [];
  let fetchAttempts = 0;
  async function fetchAllModsOnline() {
    try {
      const result = await client.query(GetModCountDocument, {}, { requestPolicy: 'network-only' }).toPromise();
      const count = result.data?.getMods.count;
      if (count) {
        fetchingMods = true;
        const pages = Math.ceil(count / MODS_PER_PAGE);

        let receivedMods = (await Promise.all(Array.from({ length: pages }).map(async (_, i) => {
          const offset = i * MODS_PER_PAGE;
          const modsPage = await client.query(GetModsDocument, { offset, limit: MODS_PER_PAGE }, { requestPolicy: 'network-only' }).toPromise();
          return modsPage.data?.getMods.mods ?? [];
        }))).flat();
        receivedMods = _.uniqBy(receivedMods, 'mod_reference');
        if (receivedMods.length !== count) {
          console.warn(`Expected ${count} mods, but received ${receivedMods.length}. Will refetch`);
          fetchAttempts++;
          if (fetchAttempts < 3) {
            setTimeout(fetchAllModsOnline, 1000);
          } else {
            console.error('Failed to fetch all mods after 3 attempts');
          }
        }
        onlineMods = receivedMods;
      }
    } finally {
      fetchingMods = false;
      $hasFetchedMods = true;
    }
  }

  let offlineMods: OfflineMod[] = [];
  async function fetchAllModsOffline() {
    offlineMods = (await OfflineGetMods()).map((mod) => ({
      ...mod,
      offline: true,
    } as OfflineMod));
  }
  
  let onlineRefreshInterval: number | undefined;

  $: if($offline !== null) {
    fetchAllModsOffline();
    if (!onlineRefreshInterval) {
      clearInterval(onlineRefreshInterval);
      onlineRefreshInterval = undefined;
    }
    if(!$offline) {
      fetchAllModsOnline();
      // setInterval returns NodeJS.Timeout, but that's not the case for the browser
      // eslint-disable-next-line
      // @ts-ignore
      onlineRefreshInterval = setInterval(fetchAllModsOnline, 5 * 60 * 1000); // 5 minutes
    } else {
      $hasFetchedMods = true;
    }
  }

  $: visibleMods = $offline ? offlineMods : onlineMods;

  $: unknownMods = $hasFetchedMods ? Object.keys($lockfileMods).filter((modReference) => !visibleMods.find((mod) => mod.mod_reference === modReference)) : [];

  let hiddenMods: PartialSMRMod[] = [];

  $: (async () => {
    if ($offline) return;

    hiddenMods = await Promise.all(unknownMods.map(async (modReference) => {
      const result = await client.query(GetHiddenModDocument, { modReference }).toPromise();
      const mod = result.data?.getModByReference;
      if (mod) {
        return mod;
      }
      return null;
    })).then((mods) => mods.filter((mod): mod is PartialSMRMod => mod !== null));
  })();

  $: knownMods = [...visibleMods, ...hiddenMods];

  $: missingModReferences = Object.keys($manifestMods)
    .filter((modReference) => !knownMods.find((knownMod) => knownMod.mod_reference === modReference));

  $: missingMods = missingModReferences.map((modReference) => {
    const offlineMod = offlineMods.find((mod) => mod.mod_reference === modReference);
    const mod = {
      mod_reference: modReference,
      name: offlineMod ? offlineMod.name : modReference,
      logo: offlineMod ? offlineMod.logo : undefined,
      authors: offlineMod ? offlineMod.authors : ['N/A'],
      missing: true,
    } as PartialMod;
    return mod;
  });

  $: mods = _.uniqBy([...knownMods, ...missingMods], 'mod_reference');

  $: availableTags = _.sortBy(_.uniqBy(onlineMods?.map((mod) => mod.tags ?? []).flat() ?? [], 'id'), 'name');

  let filteredMods: PartialMod[] = [];
  let filteringMods = false;

  $: {
    filteringMods = true;
    Promise.all(mods.map((mod) => $filterOptions[$filter].func(mod, client)))
      .then((results) => mods.filter((_, i) => results[i]))
      .then((filtered) => {
        filteredMods = (filtered.filter((mod) => !('tags' in mod) || !mod.tags || matchTags(mod.tags.map((t) => t.id), $selectedTags)));
        filteringMods = false;
      });
  }

  $: matchTags = (modTags: string[], tagIds: string[]): boolean => {
    if (tagIds.length === 0) return true;
    const modTagsSet = new Set(modTags);
    if ($tagSearchMode === settings.TagSearchMode.ALL) {
      return tagIds.every((id) => modTagsSet.has(id));
    }
    return tagIds.some((id) => modTagsSet.has(id));
  };

  let searchedMods: PartialMod[] = [];

  $: if(!$search) {
    searchedMods = filteredMods;
  } else {
    const modifiedSearchString = $search.replace(/(?:author:"(.+?)"|author:([^\s"]+))/g, '="$1$2"');
    
    const fuse = new Fuse(filteredMods, {
      keys: [
        {
          name: 'name',
          weight: 2,
        },
        {
          name: 'short_description',
          weight: 1,
        },
        {
          name: 'full_description',
          weight: 0.75,
        },
        {
          name: $offline ? 'authors' : 'authors.user.username',
          weight: 0.4,
        },
      ],
      useExtendedSearch: true,
      threshold: 0.2,
      ignoreLocation: true,
    });
    searchedMods = fuse.search(modifiedSearchString).map((result) => result.item);
  }

  $: sortedMods = _.sortBy(searchedMods, $orderByOptions[$order].func) as PartialMod[];
  
  $: displayMods = sortedMods;

  let hasCheckedStartView = false;
  $: if($startView && mods.length > 0 && !hasCheckedStartView) {
    hasCheckedStartView = true;
    if($startView === 'expanded') {
      if(displayMods.length > 0) {
        $expandedMod = displayMods[0].mod_reference;
      }
    }
  }

  $: userHasSearchText = $search != '';
  $: userHasSearchFilters = $filter != 'all';
  $: userHasTagFilter = $selectedTags.length > 0;

  const removeSearchText = () => {
    $search = '';
  };
  const removeSearchFilters = () => {
    $filter = 'all';
    $selectedTags = [];
  };

  export let hideMods: boolean = false;
</script>

<div class="h-full flex flex-col">
  <div class="flex-none z-[1]">
    <ModListFilters {availableTags} />
  </div>
  <AnnouncementsBar />
  {#if hideMods}
    <slot />
  {:else}
    <div style="position: relative;" class="py-4 grow h-0 mods-list @container/mods-list bg-surface-200-700-token">
      <div class="mr-4 h-full flex flex-col">
        {#if fetchingMods || filteringMods}
          <div class="flex items-center justify-center">
            <div class="animate-spin rounded-full aspect-square h-8 border-t-2 border-b-2 border-primary-500"/>
          </div>
        {/if}
        {#if displayMods.length === 0 && !fetchingMods && !filteringMods && $hasFetchedMods}
          <div class="flex flex-col h-full items-center justify-center">
            {#if mods.length !== 0}
              <p class="text-xl text-center text-surface-400-700-token"><T defaultValue="No mods matching your search" keyName="mods-list.no-mods-filtered"/></p>
              {#if userHasSearchFilters || userHasTagFilter}
                {#if userHasSearchText}
                  <button
                    class="btn variant-filled-primary mt-4"
                    on:click={removeSearchFilters}
                  >
                    <T defaultValue="Remove search filters" keyName="mods-list.remove-search-filters"/>
                  </button>
                  <button
                    class="btn variant-filled-primary mt-4"
                    on:click={removeSearchText}
                  >
                    <T defaultValue="Remove search text" keyName="mods-list.remove-search-text"/>
                  </button>
                {:else}
                  <button
                    class="btn variant-filled-primary mt-4"
                    on:click={removeSearchFilters}
                  >
                    <T defaultValue="Show all" keyName="mods-list.show-all"/>
                  </button>
                {/if}
              {:else}
                <button
                  class="btn variant-filled-primary mt-4"
                  on:click={removeSearchText}
                >
                  <T defaultValue="Show all" keyName="mods-list.show-all"/>
                </button>
              {/if}
            {:else}
              <p class="text-xl text-center text-surface-400-700-token"><T defaultValue="No mods found" keyName="mods-list.no-mods-found"/></p>
            {/if}
          </div>
        {:else}
          <VirtualList
            itemClass="mx-4"
            itemHeight={84}
            items={displayMods}
            let:index
            let:item={mod}>
            <ModsListItem
              {index}
              {mod}
              selected={$expandedMod == mod.mod_reference}
              on:click={() => {
                $expandedMod = mod.mod_reference;
                dispatch('expandedMod', mod.mod_reference);
              }}
            />
          </VirtualList>
        {/if}
      </div>
    </div>
  {/if}
</div>
