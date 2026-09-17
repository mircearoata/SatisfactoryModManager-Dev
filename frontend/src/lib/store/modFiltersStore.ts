import { derived, get, writable } from 'svelte/store';

import { bindingTwoWayNoExcept } from './wailsStoreBindings';

import { CompatibilityState, type GetModsQuery, TargetName } from '$lib/generated';
import { queuedMods } from '$lib/store/actionQueue';
import { favoriteMods, lockfileMods, manifestMods } from '$lib/store/ficsitCLIStore';
import { type ModCompatibilityQuery, compatibilityContext, getCompatibility } from '$lib/utils/modCompatibility';
import type { resolver } from '$wailsjs/go/models';
import { GetModFiltersFilter, GetModFiltersOrder, SetModFiltersFilter, SetModFiltersOrder } from '$wailsjs/go/settings/settings';

export type OrderByField = 'name' | 'last-updated' | 'popularity' | 'hotness' | 'views' | 'downloads' | 'search';
export type FilterField = 'all' | 'compatible' | 'favorite' | 'queued' | 'installed' | 'dependency' | 'not-installed' | 'enabled' | 'disabled';

export interface OrderBy {
  func: (mod: PartialMod) => unknown;
  hidden?: boolean;
  skipSave?: boolean;
}

export interface Filter {
  func: (mod: PartialMod) => boolean;
}

export const search = writable('');

export const orderByOptions = derived(
  [search],
  ([$search]) => (
    {
      search: { func: () => 0, skipSave: true, hidden: !$search },
      name: { func: (mod: PartialMod) => mod.name.trim().toLowerCase() },
      'last-updated': { func: (mod: PartialMod) => 'last_version_date' in mod ? Date.now() - Date.parse(mod.last_version_date) : 0 },
      popularity: { func: (mod: PartialMod) => 'popularity' in mod ? -mod.popularity : 0 },
      hotness: { func: (mod: PartialMod) => 'hotness' in mod ? -mod.hotness : 0 },
      views: { func: (mod: PartialMod) => 'views' in mod ? -mod.views : 0 },
      downloads: { func: (mod: PartialMod) => 'downloads' in mod ? -mod.downloads : 0 },
    } as Record<OrderByField, OrderBy>
  ),
);

export const visibleOrderByOptions = derived(
  [orderByOptions],
  ([$orderByOptions]) => Object.fromEntries(Object.entries($orderByOptions).filter(([_, o]) => !o.hidden)) as Record<OrderByField, OrderBy>,
);

export const filterOptions = derived(
  [favoriteMods, queuedMods, manifestMods, lockfileMods, compatibilityContext], 
  ([$favoriteMods, $queuedMods, $manifestMods, $lockfileMods, $compatibilityContext]) => (
    {
      all: { func: () => true },
      compatible: { func: (mod: PartialMod) => getCompatibility(modCompatibilityQueryFor(mod), $compatibilityContext).state !== CompatibilityState.Broken },
      favorite: { func: (mod: PartialMod) => $favoriteMods.includes(mod.mod_reference) },
      queued: { func: (mod: PartialMod) => $queuedMods.some((q) => q.mod === mod.mod_reference) },
      installed: { func: (mod: PartialMod) => mod.mod_reference in $manifestMods },
      dependency: { func: (mod: PartialMod) => !(mod.mod_reference in $manifestMods) && mod.mod_reference in $lockfileMods },
      'not-installed': { func: (mod: PartialMod) => !(mod.mod_reference in $manifestMods) },
      enabled: { func: (mod: PartialMod) => $manifestMods[mod.mod_reference]?.enabled ?? mod.mod_reference in $lockfileMods },
      disabled: { func: (mod: PartialMod) => mod.mod_reference in $manifestMods && !(mod.mod_reference in $lockfileMods) },
    } as Record<FilterField, Filter>
  ),
);

let previousSearch = '';
let previousOrderBy: OrderByField | undefined = undefined;
search.subscribe((value) => {
  if (value && !previousSearch) {
    previousOrderBy = get(order);
    order.set('search');
  } else if(!value && previousSearch && previousOrderBy && get(order) === 'search') {
    order.set(previousOrderBy);
  }
  previousSearch = value;
});

export type PartialSMRMod = GetModsQuery['getMods']['mods'][number];
export interface OfflineMod {
  offline: true;
  mod_reference: string;
  name: string;
  logo?: string;
  authors: string[];
  versions?: resolver.ModVersion[];
}
export interface MissingMod {
  missing: true;
  mod_reference: string;
  name: string;
  logo?: string;
  authors: string[];
}
export type PartialMod = PartialSMRMod | OfflineMod | MissingMod;

export const selectedTags = writable<string[]>([]);

export const order = bindingTwoWayNoExcept<OrderByField>('last-updated', 
  { initialGet: GetModFiltersOrder as () => Promise<OrderByField> }, 
  { updateFunction: async (o) => { if (!get(orderByOptions)[o]?.skipSave) { await SetModFiltersOrder(o); } } },
);
export const filter = bindingTwoWayNoExcept<FilterField>('compatible',
  { initialGet: GetModFiltersFilter as () => Promise<FilterField> },
  { updateFunction: SetModFiltersFilter },
);

export function modCompatibilityQueryFor(mod: PartialMod): ModCompatibilityQuery {
  if('offline' in mod) {
    return { 
      versions: mod.versions?.map((version) => ({
        ...version,
        targets: version.targets.map((target) => ({ targetName: target.target_name as TargetName })),
      })) ?? [],
    };
  }
  if('missing' in mod) {
    return { versions: [] };
  }
  return { 
    reported: mod.compatibility ?? undefined,
    versions: mod.versions?.map((version) => ({
      ...version,
      targets: version.targets as { targetName: TargetName }[], // The target should never be null, but the schema allows it to be for some reason
    })) ?? [],
  };
}
