import type { GetModDetailsQuery } from '$lib/generated';
import type { resolver } from '$wailsjs/go/models';

export interface OfflineMod {
  offline: true;
  mod_reference: string;
  name: string;
  authors: {
    role: 'creator';
    user: {
      username: string;
    }
  }[];
  logo?: string;
  versions?: resolver.ModVersion[];
}

export type PartialMod = NonNullable<GetModDetailsQuery['mod']> | OfflineMod;
