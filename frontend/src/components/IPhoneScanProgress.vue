<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { Loader2, Search, Check } from '@lucide/vue'

const props = defineProps<{
  phase: string
  total: number
  scanned: number
  unsyncedCount: number
}>()

const emit = defineEmits<{
  cancel: []
}>()

const progressPercent = computed(() => {
  if (props.total === 0) return 0
  return Math.round((props.scanned / props.total) * 100)
})

// Circumference for SVG circle
const radius = 54
const circumference = 2 * Math.PI * radius
const strokeDashoffset = computed(() => {
  return circumference - (progressPercent.value / 100) * circumference
})

const phaseLabel = computed(() => {
  switch (props.phase) {
    case 'discovering': return 'Discovering photos...'
    case 'checking': return 'Comparing with cloud...'
    case 'complete': return 'Scan complete'
    default: return 'Scanning...'
  }
})
</script>

<template>
  <div class="scan-progress-view">
    <!-- Circular progress -->
    <div class="circular-progress-wrapper">
      <svg class="progress-ring" width="140" height="140" viewBox="0 0 140 140">
        <!-- Background circle -->
        <circle
          class="progress-ring-bg"
          cx="70"
          cy="70"
          :r="radius"
          fill="none"
          stroke-width="6"
        />
        <!-- Progress circle -->
        <circle
          class="progress-ring-fill"
          cx="70"
          cy="70"
          :r="radius"
          fill="none"
          stroke-width="6"
          stroke-linecap="round"
          :stroke-dasharray="circumference"
          :stroke-dashoffset="strokeDashoffset"
        />
      </svg>
      
      <!-- Center content -->
      <div class="progress-center">
        <template v-if="phase === 'discovering'">
          <Search :size="24" class="center-icon discovering" />
          <span class="center-count">{{ total }}</span>
          <span class="center-label">found</span>
        </template>
        <template v-else-if="phase === 'checking'">
          <span class="center-percent">{{ progressPercent }}%</span>
          <span class="center-label">checked</span>
        </template>
        <template v-else>
          <Check :size="28" class="center-icon complete" />
          <span class="center-label">Done</span>
        </template>
      </div>
    </div>

    <!-- Phase label -->
    <p class="phase-label">{{ phaseLabel }}</p>

    <!-- Detailed stats -->
    <div class="scan-stats" v-if="phase === 'checking'">
      <div class="scan-stat">
        <span class="scan-stat-value">{{ scanned }}</span>
        <span class="scan-stat-label">/ {{ total }} checked</span>
      </div>
      <div v-if="unsyncedCount > 0" class="scan-stat unsynced">
        <span class="scan-stat-value">{{ unsyncedCount }}</span>
        <span class="scan-stat-label">unsynced</span>
      </div>
    </div>

    <!-- Cancel button -->
    <button class="cancel-scan-button" @click="emit('cancel')">
      Cancel
    </button>
  </div>
</template>

<style scoped>
.scan-progress-view {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  padding: 40px 24px;
  animation: fade-in 0.5s ease-out;
}

@keyframes fade-in {
  from { opacity: 0; transform: translateY(8px); }
  to { opacity: 1; transform: translateY(0); }
}

/* Circular progress */
.circular-progress-wrapper {
  position: relative;
  width: 140px;
  height: 140px;
}

.progress-ring {
  transform: rotate(-90deg);
}

.progress-ring-bg {
  stroke: oklch(1 0 0 / 6%);
}

.progress-ring-fill {
  stroke: oklch(0.6 0.17 162);
  transition: stroke-dashoffset 0.6s cubic-bezier(0.4, 0, 0.2, 1);
  filter: drop-shadow(0 0 6px oklch(0.6 0.17 162 / 40%));
}

.progress-center {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2px;
}

.center-icon {
  margin-bottom: 2px;
}

.center-icon.discovering {
  color: oklch(0.7 0.12 162);
  animation: pulse 1.5s ease-in-out infinite;
}

.center-icon.complete {
  color: oklch(0.7 0.17 162);
  animation: check-pop 0.5s cubic-bezier(0.34, 1.56, 0.64, 1);
}

@keyframes pulse {
  0%, 100% { opacity: 0.6; transform: scale(1); }
  50% { opacity: 1; transform: scale(1.1); }
}

@keyframes check-pop {
  0% { opacity: 0; transform: scale(0.5); }
  100% { opacity: 1; transform: scale(1); }
}

.center-count {
  font-size: 28px;
  font-weight: 700;
  color: oklch(0.95 0 0);
  font-variant-numeric: tabular-nums;
  line-height: 1;
}

.center-percent {
  font-size: 28px;
  font-weight: 700;
  color: oklch(0.95 0 0);
  font-variant-numeric: tabular-nums;
  line-height: 1;
}

.center-label {
  font-size: 11px;
  font-weight: 500;
  color: oklch(0.55 0.02 260);
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

/* Phase label */
.phase-label {
  font-size: 15px;
  font-weight: 600;
  color: oklch(0.85 0 0);
  letter-spacing: -0.01em;
}

/* Stats */
.scan-stats {
  display: flex;
  align-items: center;
  gap: 16px;
}

.scan-stat {
  display: flex;
  align-items: baseline;
  gap: 3px;
}

.scan-stat-value {
  font-size: 16px;
  font-weight: 700;
  color: oklch(0.85 0 0);
  font-variant-numeric: tabular-nums;
}

.scan-stat-label {
  font-size: 12px;
  color: oklch(0.55 0.02 260);
}

.scan-stat.unsynced .scan-stat-value {
  color: oklch(0.75 0.15 55);
}

/* Cancel button */
.cancel-scan-button {
  margin-top: 8px;
  padding: 6px 18px;
  border-radius: 8px;
  border: 1px solid oklch(1 0 0 / 10%);
  background: oklch(1 0 0 / 4%);
  color: oklch(0.65 0.02 260);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: transform 100ms ease-out, background 150ms ease-out;
}

.cancel-scan-button:active {
  transform: scale(0.95);
}

@media (hover: hover) and (pointer: fine) {
  .cancel-scan-button:hover {
    background: oklch(1 0 0 / 8%);
    color: oklch(0.8 0 0);
  }
}
</style>
