<script lang="ts">
  import FormField from './FormField.svelte';
  import type { HTMLInputAttributes } from 'svelte/elements';

  let {
    label,
    id,
    testid,
    type = 'text',
    value = $bindable(''),
    placeholder,
    error,
    description,
    required = false,
    disabled = false,
    autocomplete,
    minlength,
    maxlength,
    autofocus = false,
    onkeydown
  }: {
    label: string;
    id?: string;
    testid?: string;
    type?: 'text' | 'email' | 'password' | 'url' | 'tel';
    value?: string;
    placeholder?: string;
    error?: string;
    description?: string;
    required?: boolean;
    disabled?: boolean;
    autocomplete?: HTMLInputAttributes['autocomplete'];
    minlength?: number;
    maxlength?: number;
    autofocus?: boolean;
    onkeydown?: (e: KeyboardEvent) => void;
  } = $props();
</script>

<FormField {label} {id} {error} {description} {required}>
  <!-- svelte-ignore a11y_autofocus -->
  <input
    {id}
    data-testid={testid}
    {type}
    bind:value
    {placeholder}
    {required}
    {disabled}
    {autocomplete}
    {minlength}
    {maxlength}
    {autofocus}
    {onkeydown}
    class="input"
    aria-invalid={error ? 'true' : undefined}
    aria-describedby={error ? `${id}-error` : description ? `${id}-description` : undefined}
  />
</FormField>
