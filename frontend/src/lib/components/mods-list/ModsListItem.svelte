<script lang="ts">
  import { mdiArchiveCheck, mdiDownload, mdiEye, mdiLinkLock, mdiPause, mdiPauseCircle, mdiPlay, mdiPlayCircle, mdiStar, mdiStarMinus, mdiStarOutline, mdiStarPlus, mdiSync, mdiTagMultiple, mdiTrashCan, mdiTrayFull, mdiTrayMinus } from '@mdi/js';
  import { ProgressBar } from '@skeletonlabs/skeleton';
  import { getTranslate } from '@tolgee/svelte';
  import { createEventDispatcher } from 'svelte';

  import CompatibilityText from '$lib/components/CompatibilityText.svelte';
  import ResponsiveButton, { type ButtonDisplay } from '$lib/components/ResponsiveButton.svelte';
  import SvgIcon from '$lib/components/SVGIcon.svelte';
  import T from '$lib/components/T.svelte';
  import Thumbhash from '$lib/components/Thumbhash.svelte';
  import Tooltip from '$lib/components/Tooltip.svelte';
  import { CompatibilityState } from '$lib/generated';
  import { type PopupSettings, popup } from '$lib/skeletonExtensions';
  import {
    progressMessage,
    progressPercent,
  } from '$lib/store/ficsitCLIStore';
  import { siteURL } from '$lib/store/generalStore';
  import { type PartialMod, modCompatibilityQueryFor, search } from '$lib/store/modFiltersStore';
  import { favoriteState, modState } from '$lib/store/modState';
  import { largeNumberFormat } from '$lib/utils/dataFormats';
  import { getAuthor } from '$lib/utils/getModAuthor';
  import { compatibilityStore } from '$lib/utils/modCompatibility';

  export let mod: PartialMod;
  export let index: number;

  const { t } = getTranslate();

  const dispatch = createEventDispatcher();

  function listingClick() {
    dispatch('click');
  }

  $: authorClick = () => {
    $search = `author:"${author}"`;
  };

  export let selected: boolean;

  $: actualLogo = ('offline' in mod || 'missing' in mod) ? (mod.logo ? `data:image/png;base64, ${mod.logo}` : '/images/no_image.webp') : mod.logo;
  $: renderedLogo = actualLogo || `${$siteURL}/images/no_image.webp`;
  $: renderedThumbhash = ('logo_thumbhash' in mod ? mod.logo_thumbhash : undefined) || '2/eFDQIsFmh9h4BreKeAeQqYBxd3d3J4Jw';
  $: author = ('offline' in mod || 'missing' in mod) ? mod.authors[0] : getAuthor(mod);

  $: compatibility = compatibilityStore(modCompatibilityQueryFor(mod));

  $: state = modState(mod.mod_reference, compatibility);

  $: explicitlyDisabled = $state.installed.state && !$state.enabled.state;

  $: installButtonDisplay = ((): ButtonDisplay => {
    if ($state.inProgress) {
      return {
        icon: mdiSync,
        iconHover: mdiSync,
        tooltip: $t('mod-list-item.wait', 'Wait for the current operation to complete.'),
      };
    }
    if ($state.isDependency) {
      return {
        icon: mdiLinkLock,
        iconHover: mdiLinkLock,
        tooltip: $t('mod-list-item.dependency', 'This mod is already installed as a dependency of another mod. It cannot be uninstalled as long as other mods depend on it.'),
      };
    }
    if ($state.installed.queued) {
      return {
        icon: mdiTrayFull,
        iconHover: mdiTrayMinus,
        iconHoverClass: 'text-red-500',
        tooltip: $state.installed.state ?
          $t('mod-list-item.queued-uninstall', 'This mod is queued to be uninstalled. Click to cancel the operation.') :
            $t('mod-list-item.queued-install', 'This mod is queued to be installed. Click to cancel the operation.'),
      };
    }

    let display: ButtonDisplay = {
      icon: mdiDownload,
      iconHover: mdiDownload,
      iconHoverClass: 'text-primary-600',
      tooltip: $t('mod-list-item.install', 'Click to install this mod.'),
    };
    if ($state.installed.state) {
      display = {
        icon: mdiArchiveCheck,
        iconClass: 'text-primary-600',
        iconHover: mdiTrashCan,
        iconHoverClass: 'text-red-500',
        tooltip: $t('mod-list-item.uninstall', 'This mod is installed on this profile. Click to uninstall this mod.'),
      };
    } else if ($compatibility.state !== CompatibilityState.Works) {
      if ($compatibility.state === CompatibilityState.Broken && $compatibility.source === 'version') {
        display.tooltip = $t('mod-list-item.not-installable', 'You can\'t install this mod. Reason:');
      } else {
        display.tooltip = $t('mod-list-item.compatibility-warning', 'There are problems reported with this mod, but you can try to install it anyways. Details:');
      }
      if ($compatibility.note) {
        display.tooltipMarkdown = '<br/>' + $compatibility.note;
      } else {
        // TODO compatibility.note should always be non-null here, what should our fallback text be if it's not?
        display.tooltip += $t('mod-list-item.compatibility-note-none', ' (None specified)');
      }
    }
    if ($state.installed.queued) {
      display.tooltip = $t('mod-list-item.queued', 'This mod is already queued for another operation.');
      delete display.tooltipMarkdown;
    }
    return display;
  })();

  $: enableButtonDisplay = ((): ButtonDisplay => {
    if ($state.enabled.queued) {
      return {
        icon: mdiTrayFull,
        iconHover: mdiTrayMinus,
        iconHoverClass: 'text-red-500',
        tooltip: $state.enabled.state ?
          $t('mod-list-item.disable-queued', 'This mod is queued to be Disabled. Click to cancel the operation.') :
            $t('mod-list-item.enable-queued', 'This mod is queued to be Enabled. Click to cancel the operation.'),
      };
    }

    let display: ButtonDisplay = {
      icon: mdiPause,
      iconHover: mdiPlayCircle,
      iconHoverClass: 'text-primary-600',
      tooltip: $t('mod-list-item.enable', 'Click to enable this mod.'),
    };
    if ($state.enabled.state) {
      display = {
        icon: mdiPlay,
        iconHover: mdiPauseCircle,
        tooltip: $t('mod-list-item.disable', 'This mod is Enabled on this profile. Click to Disable it, which prevents it from loading when you start the game, but still keeps it a part of this profile.'),
      };
    }
    if ($state.installed.queued) {
      display.tooltip = $t('mod-list-item.queued', 'This mod is already queued for another operation.');
    }
    return display;
  })();
  
  $: favorite = favoriteState(mod.mod_reference);

  $: favoriteButtonDisplay = $favorite.state ? 
    {
      icon: mdiStar,
      iconHover: mdiStarMinus,
      iconClass: 'text-warning-500',
      tooltip: $t('mod-list-item.unfavorite', 'Click to remove this mod from your Favorites.'),
    } 
    : {
      icon: mdiStarOutline,
      iconHover: mdiStarPlus,
      iconHoverClass: 'text-warning-500',
      tooltip: $t('mod-list-item.favorite', 'Click to add this mod to your Favorites. Having a mod Favorited is unrelated to whether or not it\'s installed - it\'s a way to keep track of a mod for later regardless of what Profile you have selected.'),
    };
    
  $: popupId = `mods-list-item-popup-${mod.mod_reference}-${index}`;

  $: popupHover = {
    event: 'hover',
    target: popupId,
    middleware: {
      offset: 4,
    },
    placement: 'bottom-start',
  } satisfies PopupSettings;
</script>

<div
  class="my-1 px-0 @md/mods-list:h-[5.5rem] h-[4.25rem] overflow-hidden"
  class:bg-surface-50-900-token={selected}
  class:rounded-lg={selected}
  role="tab"
  tabindex="0"
  on:click={listingClick}
  on:keypress={listingClick}>
  {#if $state.inProgress}
    <div class="relative h-full">
      <ProgressBar
        class="h-full w-full"
        max={1}
        meter="bg-surface-50-900-token"
        track="bg-surface-200-700-token"
        value={$progressPercent}/>
    </div>
  {/if}
  <div class="flex relative h-full space-x-2" class:-top-full={$state.inProgress}>
    <div
      class="logo h-full @md/mods-list:w-[5.5rem] w-[4.25rem]"
      class:grayscale={explicitlyDisabled}
    >
      <Thumbhash
        alt="{mod.name} Logo"
        image={renderedLogo}
        thumbhash={renderedThumbhash}
      />
    </div>
    <div class="flex flex-col h-full grow w-0" class:opacity-30={explicitlyDisabled}>
      <div class="flex items-center" use:popup={popupHover}>
        <span
          class="shrink min-w-[7rem] truncate text-lg font-medium !leading-6"
          class:text-error-600={$compatibility.state === CompatibilityState.Broken}
          class:text-warning-500={$compatibility.state === CompatibilityState.Damaged}>
          {mod.name}
        </span>
        <div class="shrink-0 hidden @lg/mods-list:block truncate w-[7rem] grow">
          <span class="pl-1"><T defaultValue="by" keyName="mod-list-item.by-author" /></span>
          <!-- We could offer keyboard navigation for clicking this, but it's a waste of the user's time while nagivating via keyboard. If they want to search by author, they could enter the mod description pane -->
          <span
            class="text-primary-600 whitespace-nowrap"
            role="button"
            tabindex="-1"
            on:click|stopPropagation={authorClick}
            on:keypress|stopPropagation={authorClick}>{author}</span>
        </div>
      </div>
      <div class="flex flex-col grow h-0 overflow-hidden justify-around flex-wrap">
        <div class="truncate w-full @md/mods-list:text-base text-sm">{'short_description' in mod ? mod.short_description : ''}</div>
        <div class="truncate w-full h-5 hidden @md/mods-list:flex items-center space-x-1">
          <SvgIcon class="w-5 shrink-0" icon={mdiTagMultiple}/>
          {#if !('offline' in mod) && !('missing' in mod) && (mod?.tags?.length ?? -1 > 0 )}
            {#each mod?.tags ?? [] as tag}
              <span>#{tag.name}</span>
            {/each}
          {:else}
            <span><T defaultValue="(none available)" keyName="mod-list-item.no-tags" /></span>
          {/if}
        </div>
        <div class="text-sm w-full">
          <div class="flex h-5 space-x-2">
            {#if !('offline' in mod) && !('missing' in mod)}
              <div class="w-16 flex items-center space-x-1">
                <SvgIcon class="w-4 @md/mods-list:w-5" icon={mdiEye}/>
                <span>{largeNumberFormat(mod.views)}</span>
              </div>
              <div class="w-16 flex items-center space-x-1">
                <SvgIcon class="w-4 @md/mods-list:w-5" icon={mdiDownload}/>
                <span>{largeNumberFormat(mod.downloads)}</span>
              </div>
            {/if}
          </div>
        </div>
      </div>
      {#if $state.inProgress}
        <span class="shrink-0 text-sm">{$progressMessage}</span>
      {/if}
    </div>
    <!-- The purpose of the event handlers here are to prevent navigating to the mod's page when clicking on one of the sub-buttons of the div. Thus, it shouldn't be focusable despite having "interactions" -->
    <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
    <div
      class="pr-2 flex h-full items-center"
      role="separator"
      tabindex="-1"
      on:click|stopPropagation={() => { /* empty */ }}
      on:keypress|stopPropagation={() => { /* empty */ }}>
      <ResponsiveButton
        id="enable-{mod.mod_reference}"
        class="w-8 h-8 @lg/mods-list:mx-1"
        buttonClass="w-full h-full"
        disabled={!$state.enabled.canToggle}
        display={enableButtonDisplay}
        onClickAction={$state.enabled.toggle}
        visible={$state.enabled.available}
      />
      <ResponsiveButton
        id="install-{mod.mod_reference}"
        class="w-8 h-8 @lg/mods-list:mx-1"
        buttonClass="w-full h-full"
        disabled={!$state.installed.canToggle}
        display={installButtonDisplay}
        onClickAction={$state.installed.toggle}
      />
      <ResponsiveButton
        id="favorite-{mod.mod_reference}"
        class="w-8 h-8 @lg/mods-list:mx-1"
        buttonClass="w-full h-full"
        display={favoriteButtonDisplay}
        onClickAction={$favorite.toggle}
      />
    </div>
  </div>
</div>

<Tooltip disabled={!explicitlyDisabled && $compatibility.state === CompatibilityState.Works} {popupId}>
  {#if explicitlyDisabled}
    <T defaultValue="This mod is Disabled. Click the pause icon to Enable it." keyName="mod-list-item.disabled-tooltip" />
  {:else if 'hidden' in mod && mod.hidden && !$state.isDependency}
    <T defaultValue="This mod was hidden by the author." keyName="mod-list-item.hidden" />
  {:else if 'missing' in mod}
    <T defaultValue="This mod is no longer available on ficsit.app. You may want to remove it." keyName="mod-list-item.unavailable" />
  {:else if $compatibility.state !== CompatibilityState.Works}
    <CompatibilityText compatibility={$compatibility} />
  {/if}
</Tooltip>
