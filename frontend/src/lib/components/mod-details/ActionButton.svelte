<script context="module" lang="ts">
  export interface ButtonDisplay {
    text?: string;
    icon: string;
    buttonClass?: string;
    iconClass?: string;
    tooltip: string;
    tooltipMarkdown?: string;
  }
</script>

<script lang="ts">
  import Markdown from '$lib/components/Markdown.svelte';
  import SVGIcon from '$lib/components/SVGIcon.svelte';
  import Tooltip from '$lib/components/Tooltip.svelte';
  import { type PopupSettings, popup } from '$lib/skeletonExtensions';

  export let display: ButtonDisplay;
  export let disabled = false;
  export let onclick: () => void = () => {};
  export let id: string;
  export let buttonClass = '';
  export let visible = true;
    
  $: popupId = `mods-details-button-hover-${id}`;

  $: popupHover = {
    event: 'hover',
    target: popupId,
    middleware: {
      offset: 4,
    },
    placement: 'bottom-start',
  } satisfies PopupSettings;
</script>

{#if visible}
  <button
    class="btn h-full !p-4 {buttonClass} {display.buttonClass ?? ''}"
    {disabled}
    on:click={onclick}
    use:popup={popupHover}>
    {#if display.text}
      <span class="mr-auto">{display.text}</span>
    {/if}
    <SVGIcon class="h-5 w-5 shrink-0 {display.iconClass ?? ''}" icon={display.icon} />
  </button>
  <Tooltip {popupId}>
    <span>
      {display.tooltip}
      {#if display.tooltipMarkdown}
        <Markdown inline markdown={display.tooltipMarkdown}/>
      {/if}
    </span>
  </Tooltip>
{/if}
