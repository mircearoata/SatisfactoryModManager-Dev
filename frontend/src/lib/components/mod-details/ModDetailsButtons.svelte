<script lang="ts">
  import { mdiAlert, mdiDownload, mdiLinkLock, mdiPause, mdiPlay, mdiSync, mdiTrashCan, mdiTrayFull } from '@mdi/js';
  import { getTranslate } from '@tolgee/svelte';

  import type { ButtonDisplay } from './ActionButton.svelte';
  import ActionButton from './ActionButton.svelte';
  import type { PartialMod } from './helpers';

  import CompatibilityText from '$lib/components/CompatibilityText.svelte';
  import SVGIcon from '$lib/components/SVGIcon.svelte';
  import T from '$lib/components/T.svelte';
  import Tooltip from '$lib/components/Tooltip.svelte';
  import { CompatibilityState, type TargetName } from '$lib/generated';
  import { type PopupSettings, popup } from '$lib/skeletonExtensions';
  import { modState } from '$lib/store/modState';
  import { compatibilityStore } from '$lib/utils/modCompatibility';

  export let mod: PartialMod | null | undefined;

  const { t } = getTranslate();

  $: compatibilityQuery = mod ? (('offline' in mod) 
    ? {
      versions: mod.versions?.map((version) => ({
        ...version,
        targets: version.targets.map((target) => ({ targetName: target.target_name as TargetName })),
      })) ?? [],
    } : {
      reported: mod.compatibility ?? undefined,
      versions: mod.versions?.map((version) => ({
        ...version,
        targets: version.targets as { targetName: TargetName }[], // The target should never be null, but the schema allows it to be for some reason
      })) ?? [],
    }
  ) : undefined;

  $: compatibility = compatibilityQuery ? compatibilityStore(compatibilityQuery) : undefined;

  $: state = (mod && compatibility) ? modState(mod.mod_reference, compatibility) : undefined;

  $: installButtonDisplay = ((): ButtonDisplay => {
    if (!state || !$state) {
      return {
        icon: mdiSync,
        buttonClass: 'bg-primary-600',
        tooltip: $t('mod-details.buttons.install', 'Click to install this mod.'),
      };
    }

    const bgClass = $state.installed.state ? 'bg-red-500' : 'bg-primary-600';

    if ($state.inProgress && $state.installed.queued) {
      return {
        icon: mdiSync,
        buttonClass: bgClass,
        iconClass: 'animate-spin [animation-direction:reverse]',
        tooltip: $t('mod-details.buttons.wait', 'Wait for the current operation to complete.'),
      };
    }
    if ($state.isDependency) {
      return {
        icon: mdiLinkLock,
        buttonClass: bgClass,
        tooltip: $t('mod-details.buttons.dependency', 'This mod is already installed as a dependency of another mod. It cannot be uninstalled as long as other mods depend on it.'),
      };
    }
    if ($state.installed.queued) {
      return {
        icon: mdiTrayFull,
        text: $state.installed.state ? undefined : $t('mod-details.buttons.install', 'Install'),
        buttonClass: bgClass,
        tooltip: $state.installed.state ?
          $t('mod-details.buttons.queued-uninstall', 'This mod is queued to be uninstalled. Click to cancel the operation.') :
            $t('mod-details.buttons.queued-install', 'This mod is queued to be installed. Click to cancel the operation.'),
      };
    }

    let display: ButtonDisplay = {
      icon: mdiDownload,
      text: $t('mod-details.buttons.install', 'Install'),
      buttonClass: bgClass,
      tooltip: $t('mod-details.buttons.install-tooltip', 'Click to install this mod.'),
    };
    if ($state.installed.state) {
      display = {
        icon: mdiTrashCan,
        buttonClass: bgClass,
        tooltip: $t('mod-details.buttons.uninstall', 'This mod is installed on this profile. Click to uninstall this mod.'),
      };
    } else if (compatibility && $compatibility && $compatibility.state !== CompatibilityState.Works) {
      if ($compatibility.state === CompatibilityState.Broken && $compatibility.source === 'version') {
        display.tooltip = $t('mod-details.buttons.not-installable', 'You can\'t install this mod. Reason:');
      } else {
        display.tooltip = $t('mod-details.buttons.compatibility-warning', 'There are problems reported with this mod, but you can try to install it anyways. Details:');
      }
      if ($compatibility.note) {
        display.tooltipMarkdown = '<br/>' + $compatibility.note;
      } else {
        // TODO compatibility.note should always be non-null here, what should our fallback text be if it's not?
        display.tooltip += $t('mod-details.buttons.compatibility-note-none', ' (None specified)');
      }
    }
    if ($state.enabled.queued) {
      display.tooltip = $t('mod-details.buttons.queued', 'This mod is already queued for another operation.');
      delete display.tooltipMarkdown;
    }
    return display;
  })();

  $: enableButtonDisplay = ((): ButtonDisplay => {
    if (!state || !$state) {
      return {
        text: $t('mod-details.buttons.enable', 'Enable'),
        icon: mdiSync,
        buttonClass: 'bg-secondary-600',
        tooltip: $t('mod-details.buttons.enable', 'Click to enable this mod.'),
      };
    }

    const bgClass = 'bg-secondary-600';

    if ($state.inProgress && $state.enabled.queued) {
      return {
        text: $state.enabled.state ? $t('mod-details.buttons.disable', 'Disable') : $t('mod-details.buttons.enable', 'Enable'),
        icon: mdiSync,
        buttonClass: bgClass,
        iconClass: 'animate-spin [animation-direction:reverse]',
        tooltip: $t('mod-details.buttons.wait', 'Wait for the current operation to complete.'),
      };
    }

    let display: ButtonDisplay = {
      text: $t('mod-details.buttons.enable', 'Enable'),
      icon: $state.enabled.queued ? mdiTrayFull : mdiPlay,
      buttonClass: bgClass,
      tooltip: $state.enabled.queued 
        ? $t('mod-details.buttons.enable-queued', 'This mod is queued to be Enabled. Click to cancel the operation.') 
        : $t('mod-details.buttons.enable', 'Click to enable this mod.'),
    };
    if ($state.enabled.state) {
      display = {
        text: $t('mod-details.buttons.disable', 'Disable'),
        icon: $state.enabled.queued ? mdiTrayFull : mdiPause,
        buttonClass: bgClass,
        tooltip: $state.enabled.queued 
          ? $t('mod-details.buttons.disable-queued', 'This mod is queued to be Disabled. Click to cancel the operation.')
          : $t('mod-details.buttons.disable', 'This mod is Enabled on this profile. Click to Disable it, which prevents it from loading when you start the game, but still keeps it a part of this profile.'),
      };
    }
    if ($state.installed.queued) {
      display.tooltip = $t('mod-details.buttons.queued', 'This mod is already queued for another operation.');
    }
    return display;
  })();
    
  $: popupId = 'mods-details-buttons-tooltip';

  $: popupHover = {
    event: 'hover',
    target: popupId,
    middleware: {
      offset: 4,
    },
    placement: 'bottom-start',
  } satisfies PopupSettings;

  $: hasCompatibilityWarning = (compatibility && $compatibility && $compatibility.state !== CompatibilityState.Works);
  $: isHidden = (mod && ('hidden' in mod && mod.hidden) && (state && $state && !$state.isDependency));
  $: showTooltip = (hasCompatibilityWarning || isHidden);
  $: warningColor = ($compatibility && $compatibility.state === CompatibilityState.Broken) ? 'text-error-600' : 'text-warning-500';
</script>

<div class="flex space-x-2 h-full">
  {#if showTooltip}
    <div class="h-full" use:popup={popupHover}>
      <SVGIcon class="h-full aspect-square p-2 {warningColor}" icon={mdiAlert} />
    </div>
  {/if}
  <ActionButton
    id="enable"
    disabled={!mod || !state || !$state?.enabled.canToggle}
    display={enableButtonDisplay}
    onclick={state && $state?.enabled.toggle}
    visible={!!mod && !!state && $state?.enabled.available} />
  <ActionButton
    id="install"
    disabled={!mod || !state || !$state?.installed.canToggle}
    display={installButtonDisplay}
    onclick={state && $state?.installed.toggle} />
</div>

<Tooltip disabled={!showTooltip} {popupId}>
  {#if mod && ('hidden' in mod && mod.hidden) && (state && $state && !$state.isDependency)}
    <T defaultValue="This mod was hidden by the author." keyName="mod-details.buttons.hidden" />
  {:else if compatibility && $compatibility && $compatibility.state !== CompatibilityState.Works}
    <CompatibilityText compatibility={$compatibility} />
  {/if}
</Tooltip>
