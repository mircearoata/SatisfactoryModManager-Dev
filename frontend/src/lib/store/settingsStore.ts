import { GetStartView, SetStartView } from '$wailsjs/go/bindings/Settings';
import { writableBindingSync } from './wailsStoreBindings';

export type View = 'compact' | 'expanded';

export const startView = writableBindingSync<View | null>(null, { 
  initialGet: GetStartView,
  updateFunction: SetStartView
});