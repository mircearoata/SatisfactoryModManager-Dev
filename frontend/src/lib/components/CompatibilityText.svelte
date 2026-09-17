<script lang="ts">
  import Markdown from '$lib/components/Markdown.svelte';
  import T from '$lib/components/T.svelte';
  import type { CompatibilityWithSource } from '$lib/utils/modCompatibility';

  export let compatibility: CompatibilityWithSource;
</script>

{#if compatibility.source === 'reported'}
  <T defaultValue="This mod has been reported as {'{'}state{'}'} on this game version." keyName="mod-list-item.compatibility-note" params={{ state: compatibility.state }} /><br/>
{/if}
{#if compatibility.note}
  <Markdown class="[&>p]:my-0" markdown={compatibility.note}/>
{:else}
  <T defaultValue="(No further notes provided)" keyName="mod.compatibility-no-notes" />
{/if}
