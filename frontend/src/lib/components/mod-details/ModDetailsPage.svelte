<script lang="ts">
  import { mdiImport } from '@mdi/js';
  import { getContextClient, queryStore } from '@urql/svelte';

  import ModDetails from './ModDetails.svelte';
  import ModDetailsButtons from './ModDetailsButtons.svelte';
  import type { OfflineMod } from './helpers';

  import Markdown from '$lib/components/Markdown.svelte';
  import Marquee from '$lib/components/Marquee.svelte';
  import SvgIcon from '$lib/components/SVGIcon.svelte';
  import T from '$lib/components/T.svelte';
  import { GetModDetailsDocument } from '$lib/generated';
  import { expandedMod } from '$lib/store/generalStore';
  import { offline } from '$lib/store/settingsStore';
  import { OfflineGetMod } from '$wailsjs/go/ficsitcli/ficsitCLI';

  export let focusOnEntry: HTMLElement | undefined = undefined;

  const client = getContextClient();

  $: modQuery = queryStore(
    {
      query: GetModDetailsDocument,
      client,
      pause: !$expandedMod || !!$offline,
      variables: {
        modReference: $expandedMod ?? '',
      },
    },
  );
  
  let offlineMod: OfflineMod = {
    offline: true,
    mod_reference: '',
    name: '',
    versions: [],
    authors: [],
    logo: undefined,
  };

  $: if($offline && $expandedMod) {
    OfflineGetMod($expandedMod).then((mod) => {
      offlineMod = {
        ...mod,
        authors: mod.authors.map((author) => ({
          role: 'creator',
          user: {
            username: author,
          },
        })),
        offline: true,
      };
    });
  }

  $: mod = $offline ? offlineMod : ($modQuery.fetching ? null : $modQuery.data?.mod);

  function close() {
    $expandedMod = null;
  }
</script>

<div class="@container/mod-details h-full flex w-full bg-surface-200-700-token @3xl/mod-details:text-base text-sm">
  <div style="border-right-color: rgba(239, 239, 239, 0.12);" class="p-4 space-y-4 flex flex-col h-full @3xl/mod-details:w-64 w-52 overflow-y-auto">
    <ModDetails {mod} bind:focusOnEntry />
    <button
      class="btn px-4 h-8 w-full bg-secondary-600 text-sm"
      on:click={close}>
      <SvgIcon
        class="h-5 w-5 -scale-x-100"
        icon={mdiImport}/>
      <Marquee>
        <T defaultValue="Close" keyName="common.close" />
      </Marquee>
    </button>
  </div>
  <div class="flex flex-col flex-1 mr-3 my-4 min-w-0">
    <div class="break-words overflow-wrap-anywhere flex-1 px-3 overflow-y-scroll overflow-x-hidden w-full">
      {#if $offline}
        <div class="flex items-center justify-center h-full text-center font-bold">
          <T defaultValue="Offline mode is enabled. Changelogs and descriptions are not available." keyName="mod-details.offline-mode" />
        </div>
      {:else if mod && 'full_description' in mod && mod.full_description}
        <Markdown markdown={mod.full_description} />
      {:else}
        <div class="p-4 space-y-4 h-full flex flex-col">
          <div class="placeholder" />
          <div class="grid grid-cols-3 gap-4 md:gap-8">
            <div class="placeholder" />
            <div class="placeholder" />
            <div class="placeholder" />
          </div>
          <div class="grid grid-cols-4 gap-2 md:gap-4">
            <div class="placeholder" />
            <div class="placeholder" />
            <div class="placeholder" />
            <div class="placeholder" />
          </div>
          <div class="placeholder grow" />
        </div>
      {/if}
    </div>
    <div class="shrink-0 flex flex-col items-end justify-center mt-4 h-12">
      <ModDetailsButtons {mod} />
    </div>
  </div>
</div>

<style>
  .overflow-wrap-anywhere {
    overflow-wrap: anywhere;
  }
</style>
