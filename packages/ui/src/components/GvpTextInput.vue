<script setup lang="ts">
withDefaults(
  defineProps<{
    modelValue: string
    label?: string
    placeholder?: string
    error?: string
    hint?: string
    inputId?: string
    type?: string
    autocomplete?: string
  }>(),
  { type: 'text' }
)
defineEmits<{ (e: 'update:modelValue', v: string): void }>()
</script>

<template>
  <div class="gvp-field">
    <label v-if="label" class="gvp-field__label" :for="inputId">{{ label }}</label>
    <input
      :id="inputId"
      :value="modelValue"
      :type="type"
      :placeholder="placeholder"
      :autocomplete="autocomplete"
      class="gvp-field__input"
      :class="{ 'gvp-field__input--error': !!error }"
      :aria-invalid="!!error"
      :aria-describedby="error ? `${inputId}-error` : hint ? `${inputId}-hint` : undefined"
      @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
    />
    <p v-if="error" :id="`${inputId}-error`" class="gvp-field__error" role="alert">{{ error }}</p>
    <p v-else-if="hint" :id="`${inputId}-hint`" class="gvp-field__hint">{{ hint }}</p>
  </div>
</template>

<style scoped>
.gvp-field { display: flex; flex-direction: column; gap: 0.375rem; }
.gvp-field__label { font-size: 0.8125rem; font-weight: 600; color: #334155; }
.gvp-field__input {
  font-family: var(--gvp-font-sans);
  font-size: 0.875rem;
  color: #0F172A;
  background: #FFFFFF;
  border: 1px solid #CBD5E1;
  border-radius: 0.375rem;
  height: 2.5rem;
  padding: 0 0.75rem;
  width: 100%;
}
.gvp-field__input::placeholder { color: #94A3B8; }
.gvp-field__input:focus { outline: none; border-color: #0F172A; box-shadow: 0 0 0 1px #0F172A; }
.gvp-field__input--error { border-color: #B91C1C; }
.gvp-field__error { margin: 0; font-size: 0.75rem; color: #B91C1C; }
.gvp-field__hint { margin: 0; font-size: 0.75rem; color: #64748B; }
</style>
