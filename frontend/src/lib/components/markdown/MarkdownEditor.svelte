<!--
@component

Reusable Markdown editor surface. Chat-specific behavior such as attachments,
typing indicators, autocomplete, drafts, and submit shortcuts belongs in the
caller, not here.
-->
<script module lang="ts">
  import type { TipTapEditorApi } from '$lib/components/composer/TipTapEditor.svelte';

  export type MarkdownEditorApi = TipTapEditorApi;
</script>

<script lang="ts">
  import TipTapEditor from '$lib/components/composer/TipTapEditor.svelte';

  let {
    placeholder,
    editable = true,
    autofocus = false,
    variant = 'compact',
    testid,
    editorClass,
    onUpdate,
    onKeyDown,
    onPaste,
    onReady
  }: {
    placeholder?: string;
    editable?: boolean;
    autofocus?: boolean;
    variant?: 'compact' | 'document';
    testid?: string;
    editorClass?: string;
    onUpdate?: (markdown: string) => void;
    onKeyDown?: (event: KeyboardEvent) => boolean;
    onPaste?: (event: ClipboardEvent) => boolean;
    onReady?: (api: MarkdownEditorApi) => void;
  } = $props();
</script>

<TipTapEditor
  {placeholder}
  {editable}
  {autofocus}
  {variant}
  {testid}
  {editorClass}
  {onUpdate}
  {onKeyDown}
  {onPaste}
  {onReady}
/>
