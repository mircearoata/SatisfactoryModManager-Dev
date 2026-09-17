import { coerce, satisfies } from 'semver';
import { type Readable, derived } from 'svelte/store';

import {
  type Compatibility,
  CompatibilityState,
  TargetName,
  type Version,
  type VersionTarget,
} from '$lib/generated';
import {
  installsMetadata,
  selectedInstallMetadata,
  selectedProfile,
  selectedProfileTargets,
} from '$lib/store/ficsitCLIStore';
import { common, ficsitcli } from '$wailsjs/go/models';

export interface ModCompatibilityQuery {
  reported?: ReportedCompatibility;
  versions: ModVersion[];
}

export interface ReportedCompatibility {
  EA: Compatibility;
  EXP: Compatibility;
}

export interface CompatibilityWithSource extends Compatibility {
  source: 'reported' | 'version';
}

const clientTargets = [TargetName.Windows];
const serverTargets = [TargetName.LinuxServer, TargetName.WindowsServer];

interface CompatibilityContext {
  installsMetadata: Record<string, ficsitcli.installationMetadata>;
  selectedInstallMetadata: ficsitcli.installationMetadata;
  selectedProfile: string | null;
  selectedProfileTargets: Record<string, string[]>;
}

export const compatibilityContext = derived([selectedInstallMetadata, installsMetadata, selectedProfile, selectedProfileTargets], ([$selectedInstallMetadata, $installsMetadata, $selectedProfile, $selectedProfileTargets]) => (
  {
    installsMetadata: $installsMetadata,
    selectedInstallMetadata: $selectedInstallMetadata,
    selectedProfile: $selectedProfile,
    selectedProfileTargets: $selectedProfileTargets,
  } as CompatibilityContext
));

export function compatibilityStore(query: ModCompatibilityQuery[]): Readable<CompatibilityWithSource[]>;
export function compatibilityStore(query: ModCompatibilityQuery): Readable<CompatibilityWithSource>;
export function compatibilityStore(query: ModCompatibilityQuery | ModCompatibilityQuery[]): Readable<CompatibilityWithSource | CompatibilityWithSource[]> {
  return derived([compatibilityContext], ([$compatibilityContext]) => {
    if (Array.isArray(query)) {
      return query.map((mod) => getCompatibility(mod, $compatibilityContext));
    }
    return getCompatibility(query, $compatibilityContext);
  });
}

export function getCompatibility(mod: ModCompatibilityQuery, compatibilityContext: CompatibilityContext): CompatibilityWithSource {
  const installInfo = compatibilityContext.selectedInstallMetadata.info;
  if(!installInfo) {
    return { state: CompatibilityState.Broken, note: 'No game selected.', source: 'version' };
  }

  const reportedCompatibility = mod.reported ? getReportedCompatibility(mod.reported, installInfo.branch) : undefined;
  if(reportedCompatibility && reportedCompatibility.state !== CompatibilityState.Works) {
    return { ...reportedCompatibility, source: 'reported' };
  }

  if(!mod.versions || mod.versions.length === 0) {
    return { state: CompatibilityState.Broken, note: 'Mod has no versions uploaded.', source: 'version' };
  }

  const compatibleModVersions = getVersionCompatibleVersions(mod.versions, compatibilityContext);
  if (compatibleModVersions.length === 0) {
    return { state: CompatibilityState.Broken, note: 'This mod is incompatible with your game version.', source: 'version' };
  }

  const targetCompatibility = getTargetCompatibility(mod.versions, compatibleModVersions, compatibilityContext);
  if (targetCompatibility.state !== CompatibilityState.Works) {
    return { ...targetCompatibility, source: 'version' };
  }

  if (reportedCompatibility) {
    return { ...reportedCompatibility, source: 'reported' };
  }

  return { state: CompatibilityState.Works, source: 'version' };
}

function friendlyInstallName(install: common.Installation) {
  const installType = install.type === common.InstallType.WINDOWS ? 'Game' : 'Server';
  return `${install.launcher} (${installType})`;
}

function getTargetCompatibility(modVersions: ModVersion[], versionCompatibleVersions: ModVersion[], compatibilityContext: CompatibilityContext): Compatibility {
  const requiredTargets = Object.entries(compatibilityContext.selectedProfileTargets);
  if (requiredTargets.length === 0) {
    return { state: CompatibilityState.Works };
  }

  const clientRequiredTargets = requiredTargets.filter(([target]) => clientTargets.includes(target as TargetName)).map(([target]) => target as TargetName);
  const serverRequiredTargets = requiredTargets.filter(([target]) => serverTargets.includes(target as TargetName)).map(([target]) => target as TargetName);

  const versionCompatibility = modVersions.map((version) => ({ version, missingTargets: checkTargetCompatibility(version, clientRequiredTargets, serverRequiredTargets) }));
  const hasEverBeenCompatible = versionCompatibility.some((comp) => comp.missingTargets.length === 0);

  // The versions should be sorted by newest first
  const latestVersionCompatibility = versionCompatibility.filter(({ version }) => versionCompatibleVersions.includes(version))[0];

  if (latestVersionCompatibility.missingTargets.length === 0) {
    return { state: CompatibilityState.Works };
  }

  const missingTargets = requiredTargets.filter(([target]) => latestVersionCompatibility.missingTargets.includes(target as TargetName));

  const requiredTargetsWithInstall = missingTargets
    .map(([target, installs]) =>
      `${target} required by: ${installs.map((install) => friendlyInstallName(compatibilityContext.installsMetadata[install].info!)).join(',')}`)
    .join('\n\n');

  if (hasEverBeenCompatible) {
    return { state: CompatibilityState.Damaged, note: `An older version of this mod supports all your installs using profile \`${compatibilityContext.selectedProfile}\`, but the latest version does not\n\n${requiredTargetsWithInstall}` };
  }
  return { state: CompatibilityState.Broken, note: `This mod does not support one or more of your installs using profile \`${compatibilityContext.selectedProfile}\`\n\n${requiredTargetsWithInstall}` };
}

function checkTargetCompatibility(
  version: Pick<Version, 'required_on_remote'> & { targets: (Pick<VersionTarget, 'targetName'> | null)[] },
  clientRequiredTargets: TargetName[],
  serverRequiredTargets: TargetName[]): string[] {
  const unsupportedClientTargets = clientRequiredTargets.filter((target) =>
    !version.targets.some((verTarget) => verTarget && verTarget.targetName === target));
  const unsupportedServerTargets = serverRequiredTargets.filter((target) =>
    !version.targets.some((verTarget) => verTarget && verTarget.targetName === target));

  if (version.required_on_remote) {
    return [...unsupportedClientTargets, ...unsupportedServerTargets];
  }

  const hasAllClient = clientRequiredTargets.length > 0 && unsupportedClientTargets.length === 0;
  const hasAllServer = serverRequiredTargets.length > 0 && unsupportedServerTargets.length === 0;

  if (hasAllClient || hasAllServer) {
    return [];
  }
  return [...unsupportedClientTargets, ...unsupportedServerTargets];
}

function getReportedCompatibility(reportedCompatibility: ReportedCompatibility, gameBranch: common.GameBranch): Compatibility | undefined {
  switch(gameBranch) {
    case common.GameBranch.STABLE:
      return reportedCompatibility.EA;
    case common.GameBranch.EXPERIMENTAL:
      return reportedCompatibility.EXP;
  }
  return undefined;
}

interface ModVersion {
  version: string;
  game_version: string;
  required_on_remote: boolean;
  targets: { targetName: TargetName }[];
}

function getVersionCompatibleVersions(modVersions: ModVersion[], compatibilityContext: CompatibilityContext): ModVersion[] {
  const installsToCheck = Object.values(compatibilityContext.selectedProfileTargets).flat();
  return modVersions.filter((version) =>
    version.game_version === '' || installsToCheck.some((install) => satisfies(gameVersionToSemver(compatibilityContext.installsMetadata[install].info?.version ?? 0), version.game_version)),
  );
}

function gameVersionToSemver(version: number): string {
  return coerce(version)!.format();
}
