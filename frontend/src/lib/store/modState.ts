import { type Readable, derived } from 'svelte/store';

import { addQueuedModAction, queuedMods, removeQueuedModAction } from './actionQueue';
import { canInstallMods, favoriteMods, lockfileMods, manifestMods, progress } from './ficsitCLIStore';
import { error } from './generalStore';

import { CompatibilityState } from '$lib/generated';
import { type CompatibilityWithSource } from '$lib/utils/modCompatibility';
import { DisableMod, EnableMod, InstallMod, RemoveMod } from '$wailsjs/go/ficsitcli/ficsitCLI';
import { ficsitcli } from '$wailsjs/go/models';
import { FavoriteMod, UnFavoriteMod } from '$wailsjs/go/settings/settings';

export function favoriteState(modReference: string) {
  return derived([favoriteMods], ([$favoriteMods]) => {
    const isFavorite = $favoriteMods.includes(modReference);
    return {
      state: isFavorite,
      toggle: async () => {
        if(!isFavorite) {
          await FavoriteMod(modReference);
        } else {
          await UnFavoriteMod(modReference);
        }
      },
    };
  });
}

export function modState(modReference: string, compatibility: Readable<CompatibilityWithSource>) {
  return derived([manifestMods, lockfileMods, canInstallMods, queuedMods, progress, compatibility], ([$manifestMods, $lockfileMods, $canInstallMods, $queuedMods, $progress, $compatibility]) => {
    const compatibilityInstallable = $compatibility.state !== CompatibilityState.Broken || $compatibility.source !== 'version';
    
    const modActions = [
      ficsitcli.Action.INSTALL,
      ficsitcli.Action.UNINSTALL,
      ficsitcli.Action.ENABLE,
      ficsitcli.Action.DISABLE,
    ];

    const isInstalled = modReference in $manifestMods;
    const isEnabled = $manifestMods[modReference]?.enabled ?? modReference in $lockfileMods;
    const isDependency = !isInstalled && isEnabled;
    
    const inProgress = $progress && $progress.item.name === modReference && modActions.includes($progress.action);
    
    const queueEntry = $queuedMods.find((q) => q.mod === modReference);
    const queuedInstall = queueEntry && (queueEntry.action === 'install' || queueEntry.action === 'remove');
    const queuedEnable = queueEntry && (queueEntry.action === 'enable' || queueEntry.action === 'disable');

    const canPerformAction = $canInstallMods && !isDependency && (compatibilityInstallable || isInstalled) && !inProgress;

    const enableAvailable = isInstalled && !isDependency;

    return {
      inProgress,
      isDependency,
      installed: {
        state: isInstalled,
        queued: queuedInstall,
        canToggle: canPerformAction && !queuedEnable,
        toggle: () => {
          const action = isInstalled ? async () => RemoveMod(modReference).catch((e) => error.set(e)) : async () => InstallMod(modReference).catch((e) => error.set(e));
          const actionName = isInstalled ? 'remove' : 'install';
          if(queuedInstall) {
            removeQueuedModAction(modReference);
            return;
          }
          return addQueuedModAction(
            modReference,
            actionName,
            action,
          );
        },
      },
      enabled: {
        available: enableAvailable,
        state: isEnabled,
        queued: queuedEnable,
        canToggle: canPerformAction && !queuedInstall && enableAvailable,
        toggle: () => {
          const action = isEnabled ? async () => DisableMod(modReference).catch((e) => error.set(e)) : async () => EnableMod(modReference).catch((e) => error.set(e));
          const actionName = isEnabled ? 'disable' : 'enable';
          if(queuedEnable) {
            removeQueuedModAction(modReference);
            return;
          }
          return addQueuedModAction(
            modReference,
            actionName,
            action,
          );
        },
      },
    };
  });
}
