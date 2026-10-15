<script setup lang="ts">
import { computed } from 'vue'
import { Smartphone, Scan, X, Loader2 } from '@lucide/vue'

const props = defineProps<{
  deviceName: string
  isScanning: boolean
  scanProgress?: {
    phase: string
    total: number
    scanned: number
    unsyncedCount: number
  }
}>()

const emit = defineEmits<{
  scan: []
  dismiss: []
  cancelScan: []
}>()

const progressPercent = computed(() => {
  if (!props.scanProgress || props.scanProgress.total === 0) return 0
  return Math.round((props.scanProgress.scanned / props.scanProgress.total) * 100)
})

const statusText = computed(() => {
  if (!props.scanProgress) return ''
  if (props.scanProgress.phase === 'discovering') {
    return `Discovering photos... ${props.scanProgress.total} found`
  }
  if (props.scanProgress.phase === 'checking') {
    return `Checking ${props.scanProgress.scanned}/${props.scanProgress.total} • ${props.scanProgress.unsyncedCount} unsynced`
  }
  return ''
})
</script>

<template>
  <div class="iphone-banner" :class="{ scanning: isScanning }">
    <!-- Connection pulse indicator -->
    <div class="banner-glow" />
    
    <div class="banner-content">
      <div class="device-info">
        <div class="device-icon-wrapper">
          <div class="pulse-ring" />
          <div class="device-icon">
            <Smartphone :size="18" />
          </div>
        </div>
        <div class="device-details">
          <span class="device-name">{{ deviceName }}</span>
          <span v-if="!isScanning" class="device-status">Connected • Ready to scan</span>
          <span v-else class="device-status scanning-status">{{ statusText }}</span>
        </div>
      </div>

      <div class="banner-actions">
        <template v-if="!isScanning">
          <button class="scan-button" @click="emit('scan')">
            <Scan :size="14" />
            <span>Scan Photos</span>
          </button>
          <button class="dismiss-button" @click="emit('dismiss')" aria-label="Dismiss">
            <X :size="14" />
          </button>
        </template>
        <template v-else>
          <!-- Scan progress bar -->
          <div class="scan-progress-container">
            <div class="scan-progress-bar">
              <div 
                class="scan-progress-fill"
                :style="{ width: progressPercent + '%' }"
              />
            </div>
            <span class="scan-progress-text">{{ progressPercent }}%</span>
          </div>
          <button class="dismiss-button" @click="emit('cancelScan')" aria-label="Cancel scan">
            <X :size="14" />
          </button>
        </template>
      </div>
    </div>
  </div>
</template>

<style scoped>
.iphone-banner {
  position: relative;
  width: 100%;
  overflow: hidden;
  border-radius: 14px;
  background: linear-gradient(
    135deg,
    oklch(0.22 0.01 260 / 80%) 0%,
    oklch(0.18 0.015 280 / 90%) 100%
  );
  border: 1px solid oklch(1 0 0 / 8%);
  backdrop-filter: blur(20px) saturate(180%);
  -webkit-backdrop-filter: blur(20px) saturate(180%);
  animation: banner-enter 0.5s cubic-bezier(0.34, 1.56, 0.64, 1) forwards;
  transform-origin: top center;
}

@keyframes banner-enter {
  from {
    opacity: 0;
    transform: translateY(-8px) scale(0.97);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

.banner-glow {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 1px;
  background: linear-gradient(
    90deg,
    transparent 0%,
    oklch(0.7 0.17 162 / 60%) 50%,
    transparent 100%
  );
}

.banner-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 14px;
  gap: 12px;
}

.device-info {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.device-icon-wrapper {
  position: relative;
  flex-shrink: 0;
}

.pulse-ring {
  position: absolute;
  inset: -4px;
  border-radius: 50%;
  border: 1.5px solid oklch(0.7 0.17 162 / 40%);
  animation: pulse-ring-anim 2.5s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}

@keyframes pulse-ring-anim {
  0%, 100% {
    opacity: 0;
    transform: scale(0.85);
  }
  50% {
    opacity: 1;
    transform: scale(1);
  }
}

.device-icon {
  width: 34px;
  height: 34px;
  border-radius: 10px;
  background: linear-gradient(135deg, oklch(0.55 0.15 162) 0%, oklch(0.45 0.12 180) 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  box-shadow: 0 2px 8px oklch(0.5 0.15 162 / 30%);
}

.device-details {
  display: flex;
  flex-direction: column;
  gap: 1px;
  min-width: 0;
}

.device-name {
  font-size: 13px;
  font-weight: 600;
  color: oklch(0.95 0 0);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  letter-spacing: -0.01em;
}

.device-status {
  font-size: 11px;
  color: oklch(0.65 0.02 260);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.scanning-status {
  color: oklch(0.7 0.12 162);
}

.banner-actions {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}

.scan-button {
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 6px 12px;
  border-radius: 8px;
  border: none;
  background: linear-gradient(135deg, oklch(0.6 0.17 162) 0%, oklch(0.5 0.14 175) 100%);
  color: white;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  white-space: nowrap;
  transition: transform 150ms cubic-bezier(0.34, 1.56, 0.64, 1),
              opacity 150ms ease-out;
  letter-spacing: -0.01em;
}

.scan-button:active {
  transform: scale(0.95);
}

@media (hover: hover) and (pointer: fine) {
  .scan-button:hover {
    opacity: 0.9;
    transform: scale(1.02);
  }
}

.dismiss-button {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 8px;
  border: none;
  background: oklch(1 0 0 / 6%);
  color: oklch(0.6 0.02 260);
  cursor: pointer;
  transition: transform 100ms ease-out, background 150ms ease-out;
}

.dismiss-button:active {
  transform: scale(0.9);
}

@media (hover: hover) and (pointer: fine) {
  .dismiss-button:hover {
    background: oklch(1 0 0 / 12%);
    color: oklch(0.8 0 0);
  }
}

.scan-progress-container {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 100px;
}

.scan-progress-bar {
  flex: 1;
  height: 4px;
  border-radius: 2px;
  background: oklch(1 0 0 / 10%);
  overflow: hidden;
}

.scan-progress-fill {
  height: 100%;
  border-radius: 2px;
  background: linear-gradient(90deg, oklch(0.6 0.17 162), oklch(0.7 0.15 145));
  transition: width 300ms cubic-bezier(0.4, 0, 0.2, 1);
}

.scan-progress-text {
  font-size: 11px;
  font-weight: 600;
  color: oklch(0.7 0.12 162);
  font-variant-numeric: tabular-nums;
  min-width: 30px;
  text-align: right;
}
</style>
