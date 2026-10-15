<script setup lang="ts">
import { Browser } from '@wailsio/runtime'
import { Heart } from '@lucide/vue'
import { toast } from 'vue-sonner'
import PayPalIcon from './icons/PayPalIcon.vue'

interface Props {
  variant?: 'compact' | 'full' | 'pill'
  showHeart?: boolean
  class?: string
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'compact',
  showHeart: false,
  class: '',
})

const PAYPAL_URL = 'https://paypal.me/jedinmathew'

async function openPayPal(e: MouseEvent) {
  e.stopPropagation()
  try {
    if (Browser && typeof Browser.OpenURL === 'function') {
      await Browser.OpenURL(PAYPAL_URL)
    } else {
      window.open(PAYPAL_URL, '_blank')
    }
    toast.success('Opening PayPal', {
      description: 'Thank you for supporting PhotiOS! ✨',
    })
  } catch {
    window.open(PAYPAL_URL, '_blank')
  }
}
</script>

<template>
  <!-- Full variant for Settings Panel -->
  <button
    v-if="props.variant === 'full'"
    type="button"
    class="paypal-support-btn group relative w-full flex items-center justify-between px-3.5 py-2.5 rounded-xl border border-[#0079C1]/30 bg-gradient-to-r from-[#003087]/10 via-[#0079C1]/10 to-[#00457C]/10 hover:from-[#003087]/20 hover:via-[#0079C1]/20 hover:to-[#00457C]/20 hover:border-[#0079C1]/60 text-foreground transition-all duration-200 select-none cursor-pointer overflow-hidden shadow-sm hover:shadow-[0_0_20px_rgba(0,121,193,0.2)] active:scale-[0.98]"
    :class="props.class"
    @click="openPayPal"
  >
    <!-- Subtle shine hover overlay -->
    <div class="absolute inset-0 -translate-x-full group-hover:translate-x-full transition-transform duration-700 ease-out bg-gradient-to-r from-transparent via-white/5 to-transparent pointer-events-none" />

    <div class="flex items-center gap-2.5">
      <div class="size-7 rounded-lg bg-[#0079C1]/15 border border-[#0079C1]/30 flex items-center justify-center shrink-0 shadow-inner group-hover:scale-105 transition-transform duration-200">
        <PayPalIcon :size="15" />
      </div>
      <div class="flex flex-col text-left">
        <div class="flex items-center gap-1.5">
          <span class="text-xs font-semibold tracking-tight text-foreground/95">Support me on PayPal</span>
          <Heart class="size-3 text-rose-400 fill-rose-400/30 group-hover:scale-110 transition-transform duration-200" />
        </div>
        <span class="text-[10px] text-muted-foreground/80 font-mono">paypal.me/jedinmathew</span>
      </div>
    </div>

    <span class="text-[11px] font-medium text-sky-400 group-hover:text-sky-300 flex items-center gap-1 transition-colors">
      Donate
      <span class="group-hover:translate-x-0.5 transition-transform duration-150">→</span>
    </span>
  </button>

  <!-- Compact variant for Main screen toolbar / header -->
  <button
    v-else
    type="button"
    class="paypal-support-compact group relative inline-flex items-center gap-1.5 h-8 px-3 rounded-lg border border-[#0079C1]/30 bg-gradient-to-r from-[#003087]/15 to-[#0079C1]/15 hover:from-[#003087]/25 hover:to-[#0079C1]/25 hover:border-[#0079C1]/60 text-foreground select-none cursor-pointer transition-all duration-150 ease-[cubic-bezier(0.16,1,0.3,1)] active:scale-[0.95] shadow-xs hover:shadow-[0_0_12px_rgba(0,121,193,0.22)]"
    :class="props.class"
    title="Support developer on PayPal (paypal.me/jedinmathew)"
    @click="openPayPal"
  >
    <PayPalIcon :size="14" class="group-hover:scale-110 transition-transform duration-150" />
    <span class="text-xs font-medium text-sky-300 group-hover:text-sky-200 transition-colors">Support me</span>
    <Heart v-if="props.showHeart" class="size-3 text-rose-400 fill-rose-400/30 group-hover:scale-125 transition-transform duration-150" />
  </button>
</template>

<style scoped>
.paypal-support-btn,
.paypal-support-compact {
  -webkit-tap-highlight-color: transparent;
  user-select: none;
  font-family: inherit;
}
</style>
