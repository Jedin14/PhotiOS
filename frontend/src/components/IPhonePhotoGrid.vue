<script setup lang="ts">
import { computed, ref, watch, onMounted, onUnmounted } from 'vue'
import {
  Check,
  CheckCheck,
  Upload,
  Image as ImageIcon,
  X,
  ArrowLeft,
  FolderPlus,
  Cloud,
  Play,
  Maximize2,
  ChevronLeft,
  ChevronRight,
  Folder,
  Layers,
  Sparkles
} from '@lucide/vue'
import Button from './ui/button/Button.vue'
import Input from './ui/input/Input.vue'
import Label from './ui/label/Label.vue'
import SupportButton from './SupportButton.vue'
import * as IPhoneService from '../../bindings/app/core/iphoneservice'

export interface IPhonePhoto {
  path: string
  fileName: string
  size: number
  extension: string
  thumbnail?: string
  isSynced: boolean
}

export interface ScanResult {
  photos: IPhonePhoto[]
  totalOnPhone: number
  unsyncedCount: number
  syncedCount: number
}

export interface UploadOptions {
  paths: string[]
  albumMode: 'none' | 'name' | 'auto'
  albumName?: string
}

const props = defineProps<{
  scanResult: ScanResult
  isUploading: boolean
}>()

const emit = defineEmits<{
  upload: [options: UploadOptions]
  uploadSelected: [paths: string[]]
  uploadAll: []
  back: []
}>()

// Selected photos
const selectedPaths = ref<Set<string>>(new Set())
const isAllSelected = computed(() => {
  return props.scanResult.photos.length > 0 &&
    selectedPaths.value.size === props.scanResult.photos.length
})
const selectedCount = computed(() => selectedPaths.value.size)

// Lazy thumbnails cache
const lazyThumbnails = ref<Record<string, string>>({})

async function loadMissingThumbnail(path: string) {
  if (lazyThumbnails.value[path]) return
  try {
    const thumb = await IPhoneService.GetThumbnail(path)
    if (thumb) {
      lazyThumbnails.value[path] = thumb
    }
  } catch (err) {
    console.debug('Failed to get thumbnail for', path, err)
  }
}

// Lightbox preview state
const previewIndex = ref<number | null>(null)
const highResPreviews = ref<Record<string, string>>({})

const previewPhoto = computed(() => {
  if (previewIndex.value === null || !props.scanResult.photos[previewIndex.value]) {
    return null
  }
  return props.scanResult.photos[previewIndex.value]
})

async function loadPreview(path: string) {
  if (highResPreviews.value[path]) return
  try {
    const preview = await IPhoneService.GetPreview(path)
    if (preview) {
      highResPreviews.value[path] = preview
    }
  } catch (err) {
    console.debug('Failed to get high-res preview for', path, err)
  }
}

function openPreview(index: number) {
  previewIndex.value = index
  const p = props.scanResult.photos[index]
  if (p) {
    if (!p.thumbnail && !lazyThumbnails.value[p.path]) {
      loadMissingThumbnail(p.path)
    }
    loadPreview(p.path)
  }
}

function closePreview() {
  previewIndex.value = null
}

function nextPreview() {
  if (previewIndex.value !== null && previewIndex.value < props.scanResult.photos.length - 1) {
    openPreview(previewIndex.value + 1)
  }
}

function prevPreview() {
  if (previewIndex.value !== null && previewIndex.value > 0) {
    openPreview(previewIndex.value - 1)
  }
}

// Album modal state
const showAlbumModal = ref(false)
type AlbumMode = 'new' | 'existing' | 'auto'
const albumTab = ref<AlbumMode>('new')
const newAlbumName = ref('')
const existingAlbumKeyOrName = ref('')

// Compute suggested album names based on current date
const defaultSuggestions = computed(() => {
  const now = new Date()
  const monthNames = [
    'January', 'February', 'March', 'April', 'May', 'June',
    'July', 'August', 'September', 'October', 'November', 'December'
  ]
  const currentMonth = `${monthNames[now.getMonth()]} ${now.getFullYear()}`
  const prevDate = new Date(now.getFullYear(), now.getMonth() - 1, 1)
  const prevMonth = `${monthNames[prevDate.getMonth()]} ${prevDate.getFullYear()}`
  return ['iPhone Photos', currentMonth, prevMonth]
})

function selectSuggestion(name: string) {
  newAlbumName.value = name
}

function togglePhoto(path: string) {
  const newSet = new Set(selectedPaths.value)
  if (newSet.has(path)) {
    newSet.delete(path)
  } else {
    newSet.add(path)
  }
  selectedPaths.value = newSet
}

function selectAll() {
  if (isAllSelected.value) {
    selectedPaths.value = new Set()
  } else {
    selectedPaths.value = new Set(props.scanResult.photos.map(p => p.path))
  }
}

function isVideo(photo: IPhonePhoto): boolean {
  const ext = (photo.extension || '').toLowerCase()
  return ext === '.mov' || ext === '.mp4' || ext === '.m4v'
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

// Upload handlers
function handleUploadOnly() {
  const paths = Array.from(selectedPaths.value)
  emit('upload', {
    paths,
    albumMode: 'none',
  })
  emit('uploadSelected', paths)
}

function openAlbumModal() {
  if (!newAlbumName.value && defaultSuggestions.value.length > 1) {
    newAlbumName.value = defaultSuggestions.value[1] // Default to current month
  }
  showAlbumModal.value = true
}

function closeAlbumModal() {
  showAlbumModal.value = false
}

function handleConfirmAlbumUpload() {
  const paths = Array.from(selectedPaths.value)
  if (albumTab.value === 'auto') {
    emit('upload', {
      paths,
      albumMode: 'auto',
    })
  } else if (albumTab.value === 'new') {
    emit('upload', {
      paths,
      albumMode: 'name',
      albumName: newAlbumName.value.trim() || 'iPhone Photos',
    })
  } else {
    emit('upload', {
      paths,
      albumMode: 'name',
      albumName: existingAlbumKeyOrName.value.trim(),
    })
  }
  emit('uploadSelected', paths)
  showAlbumModal.value = false
}

// Keyboard shortcuts for lightbox & modal
function handleKeyDown(e: KeyboardEvent) {
  if (showAlbumModal.value) {
    if (e.key === 'Escape') closeAlbumModal()
    return
  }
  if (previewIndex.value !== null) {
    if (e.key === 'Escape') closePreview()
    else if (e.key === 'ArrowRight') nextPreview()
    else if (e.key === 'ArrowLeft') prevPreview()
    else if (e.key === ' ') {
      e.preventDefault()
      if (previewPhoto.value) togglePhoto(previewPhoto.value.path)
    }
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleKeyDown)
  // Check for any missing thumbnails and load them
  for (const photo of props.scanResult.photos) {
    if (!photo.thumbnail) {
      loadMissingThumbnail(photo.path)
    }
  }
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeyDown)
})

// Auto-select all on mount
watch(() => props.scanResult.photos, (photos) => {
  if (photos.length > 0 && selectedPaths.value.size === 0) {
    selectedPaths.value = new Set(photos.map(p => p.path))
  }
  // Check for missing thumbnails
  for (const p of photos) {
    if (!p.thumbnail && !lazyThumbnails.value[p.path]) {
      loadMissingThumbnail(p.path)
    }
  }
}, { immediate: true })
</script>

<template>
  <div class="photo-grid-container">
    <!-- Header -->
    <div class="grid-header">
      <button class="back-button" title="Back" @click="emit('back')">
        <ArrowLeft :size="16" />
      </button>
      <div class="header-info">
        <div class="header-title-row">
          <h2 class="header-title">Unsynced iPhone Media</h2>
          <span class="brand-pill">
            <Sparkles :size="10" /> PhotiOS
          </span>
        </div>
        <p class="header-subtitle">
          {{ scanResult.unsyncedCount }} of {{ scanResult.totalOnPhone }} photos need uploading
        </p>
      </div>
      <SupportButton variant="compact" class="grid-support-btn" />
    </div>

    <!-- Stats bar -->
    <div class="stats-bar">
      <div class="stat">
        <span class="stat-value">{{ scanResult.totalOnPhone }}</span>
        <span class="stat-label">On Phone</span>
      </div>
      <div class="stat-divider" />
      <div class="stat">
        <span class="stat-value synced">{{ scanResult.syncedCount }}</span>
        <span class="stat-label">Synced</span>
      </div>
      <div class="stat-divider" />
      <div class="stat">
        <span class="stat-value unsynced">{{ scanResult.unsyncedCount }}</span>
        <span class="stat-label">Unsynced</span>
      </div>
    </div>

    <!-- Selection controls -->
    <div class="selection-bar">
      <button class="select-all-button" @click="selectAll">
        <CheckCheck :size="14" />
        <span>{{ isAllSelected ? 'Deselect All' : 'Select All' }}</span>
      </button>
      <div v-if="selectedCount > 0" class="selection-count">
        <span class="count-badge">{{ selectedCount }}</span>
        <span class="count-label">selected</span>
      </div>
    </div>

    <!-- Photo grid -->
    <div class="photo-grid-scroll">
      <div class="photo-grid">
        <div
          v-for="(photo, index) in scanResult.photos"
          :key="photo.path"
          class="photo-cell"
          :class="{ selected: selectedPaths.has(photo.path) }"
          :style="{ '--enter-delay': Math.min(index * 25, 400) + 'ms' }"
          @click="togglePhoto(photo.path)"
          @dblclick.stop="openPreview(index)"
        >
          <div class="photo-thumbnail">
            <img 
              v-if="photo.thumbnail || lazyThumbnails[photo.path]"
              :src="photo.thumbnail || lazyThumbnails[photo.path]" 
              :alt="photo.fileName"
              loading="lazy"
            />
            <div v-else class="photo-placeholder">
              <ImageIcon :size="20" />
            </div>

            <!-- Video badge -->
            <div v-if="isVideo(photo)" class="video-indicator">
              <Play :size="10" fill="white" />
              <span>MOV</span>
            </div>

            <!-- Preview zoom icon button -->
            <button 
              class="preview-trigger-btn"
              title="Preview photo"
              @click.stop="openPreview(index)"
            >
              <Maximize2 :size="12" />
            </button>
          </div>
          
          <!-- Selection overlay -->
          <div class="selection-overlay" :class="{ active: selectedPaths.has(photo.path) }">
            <div class="check-circle">
              <Check :size="12" />
            </div>
          </div>

          <!-- File info -->
          <div class="photo-info">
            <span class="photo-name">{{ photo.fileName }}</span>
            <span class="photo-size">{{ formatBytes(photo.size) }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Upload action bar -->
    <div class="action-bar" :class="{ visible: selectedCount > 0 }">
      <div class="action-bar-content">
        <div class="action-info">
          <Upload :size="15" />
          <span class="selected-text">{{ selectedCount }} selected</span>
        </div>
        <div class="action-buttons">
          <!-- Button 1: Upload Only (Direct) -->
          <Button
            variant="secondary"
            class="upload-btn upload-only-btn"
            :disabled="isUploading || selectedCount === 0"
            title="Upload directly to Google Photos library"
            @click="handleUploadOnly"
          >
            <Cloud :size="14" class="mr-1.5" />
            Upload Only
          </Button>

          <!-- Button 2: Upload to Album (Selection) -->
          <Button
            class="upload-btn upload-album-btn"
            :disabled="isUploading || selectedCount === 0"
            title="Choose or create an album"
            @click="openAlbumModal"
          >
            <FolderPlus :size="14" class="mr-1.5" />
            To Album...
          </Button>
        </div>
      </div>
    </div>

    <!-- Lightbox Preview Modal -->
    <div v-if="previewPhoto" class="lightbox-overlay" @click="closePreview">
      <div class="lightbox-dialog" @click.stop>
        <div class="lightbox-header">
          <div class="lightbox-title-area">
            <span class="lightbox-filename">{{ previewPhoto.fileName }}</span>
            <span class="lightbox-meta">
              {{ formatBytes(previewPhoto.size) }} • {{ previewPhoto.extension.toUpperCase().replace('.', '') }}
            </span>
          </div>
          <button class="lightbox-close" title="Close (Esc)" @click="closePreview">
            <X :size="18" />
          </button>
        </div>

        <div class="lightbox-media-container">
          <img 
            v-if="highResPreviews[previewPhoto.path] || previewPhoto.thumbnail || lazyThumbnails[previewPhoto.path]"
            :src="highResPreviews[previewPhoto.path] || previewPhoto.thumbnail || lazyThumbnails[previewPhoto.path]"
            :alt="previewPhoto.fileName"
            class="lightbox-img"
          />
          <div v-else class="lightbox-loading">
            <ImageIcon :size="36" class="animate-pulse text-muted-foreground/40" />
            <span class="loading-subtext">Generating preview...</span>
          </div>
        </div>

        <div class="lightbox-footer">
          <button 
            class="lightbox-nav-btn"
            :disabled="previewIndex === 0"
            title="Previous (Left Arrow)"
            @click="prevPreview"
          >
            <ChevronLeft :size="20" />
          </button>

          <button 
            class="lightbox-select-btn"
            :class="{ active: selectedPaths.has(previewPhoto.path) }"
            @click="togglePhoto(previewPhoto.path)"
          >
            <div class="check-circle-sm">
              <Check v-if="selectedPaths.has(previewPhoto.path)" :size="12" />
            </div>
            <span>{{ selectedPaths.has(previewPhoto.path) ? 'Selected for Upload' : 'Click to Select' }}</span>
          </button>

          <button 
            class="lightbox-nav-btn"
            :disabled="previewIndex === scanResult.photos.length - 1"
            title="Next (Right Arrow)"
            @click="nextPreview"
          >
            <ChevronRight :size="20" />
          </button>
        </div>
      </div>
    </div>

    <!-- Upload to Album Sheet/Modal -->
    <div v-if="showAlbumModal" class="album-modal-overlay" @click="closeAlbumModal">
      <div class="album-modal-card" @click.stop>
        <div class="album-modal-header">
          <div class="album-modal-badge">
            <FolderPlus :size="18" />
          </div>
          <div class="album-modal-titles">
            <h3 class="album-modal-title">Upload to Album</h3>
            <p class="album-modal-subtitle">
              Add {{ selectedCount }} item{{ selectedCount !== 1 ? 's' : '' }} to Google Photos
            </p>
          </div>
          <button class="album-modal-close" @click="closeAlbumModal">
            <X :size="16" />
          </button>
        </div>

        <!-- Segmented Mode Control -->
        <div class="segmented-control">
          <button 
            class="segment-pill" 
            :class="{ active: albumTab === 'new' }"
            @click="albumTab = 'new'"
          >
            <Sparkles :size="13" />
            <span>New Album</span>
          </button>
          <button 
            class="segment-pill" 
            :class="{ active: albumTab === 'existing' }"
            @click="albumTab = 'existing'"
          >
            <Folder :size="13" />
            <span>Existing</span>
          </button>
          <button 
            class="segment-pill" 
            :class="{ active: albumTab === 'auto' }"
            @click="albumTab = 'auto'"
          >
            <Layers :size="13" />
            <span>Auto by Month</span>
          </button>
        </div>

        <!-- Tab 1: New Album -->
        <div v-if="albumTab === 'new'" class="album-tab-content">
          <div class="input-group">
            <Label for="new-album-input" class="input-label">Album Title</Label>
            <Input 
              id="new-album-input"
              v-model="newAlbumName"
              placeholder="e.g. Vacation, October 2026"
              autofocus
              @keydown.enter="handleConfirmAlbumUpload"
            />
          </div>
          
          <div class="suggestions-row">
            <span class="suggestions-label">Suggestions:</span>
            <button 
              v-for="sug in defaultSuggestions"
              :key="sug"
              class="sug-chip"
              :class="{ selected: newAlbumName === sug }"
              @click="selectSuggestion(sug)"
            >
              {{ sug }}
            </button>
          </div>
        </div>

        <!-- Tab 2: Existing Album -->
        <div v-else-if="albumTab === 'existing'" class="album-tab-content">
          <div class="input-group">
            <Label for="existing-album-input" class="input-label">Existing Album Name or Key</Label>
            <Input 
              id="existing-album-input"
              v-model="existingAlbumKeyOrName"
              placeholder="Album title or AF1Qip... key"
              autofocus
              @keydown.enter="existingAlbumKeyOrName.trim() && handleConfirmAlbumUpload()"
            />
          </div>
          <p class="tab-note">
            Enter the exact title of an existing album, or paste its Google Photos album share key.
          </p>
        </div>

        <!-- Tab 3: Auto by Month -->
        <div v-else class="album-tab-content">
          <div class="auto-album-card">
            <div class="auto-album-icon">
              <Layers :size="20" />
            </div>
            <div class="auto-album-text">
              <h4>Automatic Folder Organization</h4>
              <p>
                Photos will be automatically placed into albums named after your iPhone's camera folders (e.g. <code>202609_a</code>, <code>202610_a</code>).
              </p>
            </div>
          </div>
        </div>

        <!-- Modal Footer Actions -->
        <div class="album-modal-footer">
          <Button variant="ghost" class="cancel-btn" @click="closeAlbumModal">
            Cancel
          </Button>
          <Button 
            class="confirm-album-btn"
            :disabled="albumTab === 'existing' && !existingAlbumKeyOrName.trim()"
            @click="handleConfirmAlbumUpload"
          >
            <Upload :size="14" class="mr-1.5" />
            Upload to Album ({{ selectedCount }})
          </Button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.photo-grid-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  width: 100%;
  overflow: hidden;
  position: relative;
  user-select: none;
}

/* Header */
.grid-header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 16px 8px;
}

.back-button {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 10px;
  border: none;
  background: oklch(1 0 0 / 6%);
  color: oklch(0.8 0 0);
  cursor: pointer;
  transition: transform 120ms cubic-bezier(0.16, 1, 0.3, 1), background 150ms;
  flex-shrink: 0;
}

.back-button:active {
  transform: scale(0.92);
}

@media (hover: hover) and (pointer: fine) {
  .back-button:hover {
    background: oklch(1 0 0 / 12%);
  }
}

.header-info {
  flex: 1;
  min-width: 0;
}

.grid-support-btn {
  margin-left: auto;
  flex-shrink: 0;
}

.header-title-row {
  display: flex;
  align-items: center;
  gap: 6px;
}

.header-title {
  font-size: 16px;
  font-weight: 700;
  letter-spacing: -0.02em;
  color: oklch(0.95 0 0);
  line-height: 1.2;
}

.brand-pill {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  padding: 1px 6px;
  border-radius: 10px;
  font-size: 10px;
  font-weight: 600;
  background: oklch(0.6 0.17 162 / 15%);
  color: oklch(0.8 0.12 162);
  border: 1px solid oklch(0.6 0.17 162 / 25%);
}

.header-subtitle {
  font-size: 11px;
  color: oklch(0.6 0.02 260);
  margin-top: 1px;
}

/* Stats */
.stats-bar {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  padding: 8px 16px;
  margin: 4px 16px;
  border-radius: 12px;
  background: oklch(1 0 0 / 4%);
  border: 1px solid oklch(1 0 0 / 6%);
}

.stat {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1px;
}

.stat-value {
  font-size: 15px;
  font-weight: 700;
  color: oklch(0.9 0 0);
  font-variant-numeric: tabular-nums;
}

.stat-value.synced {
  color: oklch(0.7 0.17 162);
}

.stat-value.unsynced {
  color: oklch(0.78 0.16 65);
}

.stat-label {
  font-size: 10px;
  font-weight: 500;
  color: oklch(0.55 0.02 260);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.stat-divider {
  width: 1px;
  height: 22px;
  background: oklch(1 0 0 / 8%);
}

/* Selection bar */
.selection-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 16px;
}

.select-all-button {
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 5px 10px;
  border-radius: 8px;
  border: 1px solid oklch(1 0 0 / 10%);
  background: oklch(1 0 0 / 4%);
  color: oklch(0.8 0 0);
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: transform 120ms cubic-bezier(0.16, 1, 0.3, 1), background 150ms ease-out;
}

.select-all-button:active {
  transform: scale(0.95);
}

@media (hover: hover) and (pointer: fine) {
  .select-all-button:hover {
    background: oklch(1 0 0 / 8%);
  }
}

.selection-count {
  display: flex;
  align-items: center;
  gap: 5px;
}

.count-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 20px;
  height: 20px;
  padding: 0 6px;
  border-radius: 10px;
  background: oklch(0.6 0.17 162);
  color: white;
  font-size: 11px;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
  animation: badge-pop 0.35s cubic-bezier(0.34, 1.56, 0.64, 1);
}

@keyframes badge-pop {
  0% { transform: scale(0.5); opacity: 0; }
  100% { transform: scale(1); opacity: 1; }
}

.count-label {
  font-size: 12px;
  color: oklch(0.6 0.02 260);
}

/* Photo grid scroll area */
.photo-grid-scroll {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 0 12px;
  scrollbar-width: thin;
}

.photo-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 4px;
  padding-bottom: 84px;
}

/* Photo cell */
.photo-cell {
  position: relative;
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  animation: cell-enter 0.35s cubic-bezier(0.34, 1.56, 0.64, 1) var(--enter-delay, 0ms) both;
  transition: transform 140ms cubic-bezier(0.34, 1.56, 0.64, 1);
}

.photo-cell:active {
  transform: scale(0.95);
}

@keyframes cell-enter {
  from {
    opacity: 0;
    transform: scale(0.88);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}

.photo-cell.selected {
  outline: 2px solid oklch(0.6 0.17 162);
  outline-offset: -2px;
  border-radius: 8px;
}

.photo-thumbnail {
  aspect-ratio: 1;
  background: oklch(0.18 0.01 260);
  overflow: hidden;
  position: relative;
}

.photo-thumbnail img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.photo-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: oklch(0.4 0.02 260);
  background: linear-gradient(135deg, oklch(0.16 0.01 260), oklch(0.2 0.01 260));
}

.video-indicator {
  position: absolute;
  bottom: 5px;
  left: 5px;
  display: flex;
  align-items: center;
  gap: 3px;
  padding: 2px 5px;
  border-radius: 4px;
  background: oklch(0 0 0 / 65%);
  backdrop-filter: blur(4px);
  color: white;
  font-size: 9px;
  font-weight: 700;
  letter-spacing: 0.02em;
}

.preview-trigger-btn {
  position: absolute;
  top: 5px;
  left: 5px;
  width: 22px;
  height: 22px;
  border-radius: 6px;
  border: none;
  background: oklch(0 0 0 / 40%);
  backdrop-filter: blur(4px);
  color: oklch(0.85 0 0);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  opacity: 0;
  transition: opacity 150ms, transform 100ms;
}

.preview-trigger-btn:active {
  transform: scale(0.9);
}

.photo-cell:hover .preview-trigger-btn {
  opacity: 1;
}

/* Selection overlay */
.selection-overlay {
  position: absolute;
  top: 5px;
  right: 5px;
  pointer-events: none;
}

.check-circle {
  width: 22px;
  height: 22px;
  border-radius: 50%;
  border: 2px solid oklch(1 0 0 / 50%);
  background: oklch(0 0 0 / 30%);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  color: transparent;
  transition: all 180ms cubic-bezier(0.34, 1.56, 0.64, 1);
}

.selection-overlay.active .check-circle {
  background: oklch(0.6 0.17 162);
  border-color: oklch(0.6 0.17 162);
  color: white;
  transform: scale(1.05);
}

/* Photo info */
.photo-info {
  padding: 4px 6px;
  display: flex;
  flex-direction: column;
  gap: 0;
  background: oklch(0.14 0.01 260);
}

.photo-name {
  font-size: 10px;
  font-weight: 500;
  color: oklch(0.85 0 0);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.photo-size {
  font-size: 9px;
  color: oklch(0.5 0.02 260);
  font-variant-numeric: tabular-nums;
}

/* Action bar */
.action-bar {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  transform: translateY(100%);
  opacity: 0;
  transition: transform 0.35s cubic-bezier(0.34, 1.56, 0.64, 1),
              opacity 0.25s ease-out;
  pointer-events: none;
  z-index: 20;
}

.action-bar.visible {
  transform: translateY(0);
  opacity: 1;
  pointer-events: auto;
}

.action-bar-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 14px;
  background: linear-gradient(
    to top,
    oklch(0.12 0.01 260) 0%,
    oklch(0.12 0.01 260 / 96%) 70%,
    oklch(0.12 0.01 260 / 0%) 100%
  );
  padding-top: 26px;
  gap: 8px;
}

.action-info {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  font-weight: 500;
  color: oklch(0.75 0 0);
  white-space: nowrap;
}

.selected-text {
  font-variant-numeric: tabular-nums;
}

.action-buttons {
  display: flex;
  align-items: center;
  gap: 8px;
}

.upload-btn {
  cursor: pointer;
  font-weight: 600 !important;
  font-size: 12px !important;
  height: 34px !important;
  padding: 0 12px !important;
  border-radius: 9px !important;
  letter-spacing: -0.01em;
  transition: transform 120ms cubic-bezier(0.16, 1, 0.3, 1) !important;
}

.upload-btn:active {
  transform: scale(0.95);
}

.upload-only-btn {
  background: oklch(1 0 0 / 10%) !important;
  color: oklch(0.9 0 0) !important;
  border: 1px solid oklch(1 0 0 / 12%) !important;
}

.upload-album-btn {
  background: oklch(0.6 0.17 162) !important;
  color: white !important;
}

/* Lightbox Modal */
.lightbox-overlay {
  position: absolute;
  inset: 0;
  background: oklch(0 0 0 / 85%);
  backdrop-filter: blur(16px) saturate(180%);
  z-index: 50;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 12px;
  animation: fade-enter 0.2s ease-out;
}

.lightbox-dialog {
  display: flex;
  flex-direction: column;
  width: 100%;
  max-width: 440px;
  height: 90%;
  max-height: 560px;
  background: oklch(0.16 0.01 260);
  border: 1px solid oklch(1 0 0 / 10%);
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 20px 40px oklch(0 0 0 / 60%);
}

.lightbox-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 14px;
  border-bottom: 1px solid oklch(1 0 0 / 8%);
}

.lightbox-title-area {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.lightbox-filename {
  font-size: 13px;
  font-weight: 600;
  color: oklch(0.95 0 0);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.lightbox-meta {
  font-size: 11px;
  color: oklch(0.55 0.02 260);
}

.lightbox-close {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 8px;
  border: none;
  background: oklch(1 0 0 / 8%);
  color: oklch(0.8 0 0);
  cursor: pointer;
  transition: transform 120ms cubic-bezier(0.16, 1, 0.3, 1), background 150ms;
}

.lightbox-close:active {
  transform: scale(0.92);
}

.lightbox-close:hover {
  background: oklch(1 0 0 / 14%);
}

.lightbox-media-container {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background: oklch(0.1 0.01 260);
  overflow: hidden;
  padding: 8px;
}

.lightbox-img {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
  border-radius: 8px;
}

.lightbox-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.loading-subtext {
  font-size: 11px;
  color: oklch(0.55 0.02 260);
}

.lightbox-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  border-top: 1px solid oklch(1 0 0 / 8%);
  background: oklch(0.14 0.01 260);
}

.lightbox-nav-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 8px;
  border: 1px solid oklch(1 0 0 / 10%);
  background: oklch(1 0 0 / 5%);
  color: oklch(0.85 0 0);
  cursor: pointer;
  transition: transform 120ms cubic-bezier(0.16, 1, 0.3, 1);
}

.lightbox-nav-btn:active {
  transform: scale(0.92);
}

.lightbox-nav-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.lightbox-select-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 14px;
  border-radius: 20px;
  border: 1px solid oklch(1 0 0 / 15%);
  background: oklch(1 0 0 / 6%);
  color: oklch(0.8 0 0);
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 150ms cubic-bezier(0.16, 1, 0.3, 1);
}

.lightbox-select-btn:active {
  transform: scale(0.96);
}

.lightbox-select-btn.active {
  background: oklch(0.6 0.17 162 / 20%);
  border-color: oklch(0.6 0.17 162);
  color: oklch(0.9 0.1 162);
}

.check-circle-sm {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  border: 1.5px solid currentColor;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* Album Selection Modal Sheet */
.album-modal-overlay {
  position: absolute;
  inset: 0;
  background: oklch(0 0 0 / 70%);
  backdrop-filter: blur(12px) saturate(180%);
  z-index: 60;
  display: flex;
  align-items: flex-end;
  justify-content: center;
  animation: fade-enter 0.2s ease-out;
}

.album-modal-card {
  width: 100%;
  background: oklch(0.18 0.01 260);
  border-top: 1px solid oklch(1 0 0 / 12%);
  border-radius: 20px 20px 0 0;
  padding: 18px 18px 24px;
  box-shadow: 0 -10px 30px oklch(0 0 0 / 50%);
  animation: slide-up 0.28s cubic-bezier(0.16, 1, 0.3, 1);
  display: flex;
  flex-direction: column;
  gap: 14px;
}

@keyframes slide-up {
  from { transform: translateY(100%); }
  to { transform: translateY(0); }
}

@keyframes fade-enter {
  from { opacity: 0; }
  to { opacity: 1; }
}

.album-modal-header {
  display: flex;
  align-items: center;
  gap: 12px;
}

.album-modal-badge {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  background: oklch(0.6 0.17 162 / 15%);
  color: oklch(0.7 0.17 162);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.album-modal-titles {
  flex: 1;
  min-width: 0;
}

.album-modal-title {
  font-size: 15px;
  font-weight: 700;
  color: oklch(0.95 0 0);
  letter-spacing: -0.01em;
}

.album-modal-subtitle {
  font-size: 12px;
  color: oklch(0.6 0.02 260);
}

.album-modal-close {
  width: 28px;
  height: 28px;
  border-radius: 8px;
  border: none;
  background: oklch(1 0 0 / 6%);
  color: oklch(0.7 0 0);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: transform 120ms cubic-bezier(0.16, 1, 0.3, 1);
}

.album-modal-close:active {
  transform: scale(0.92);
}

/* Segmented Control */
.segmented-control {
  display: flex;
  background: oklch(0.13 0.01 260);
  padding: 3px;
  border-radius: 10px;
  gap: 2px;
}

.segment-pill {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 5px;
  padding: 6px 8px;
  border-radius: 7px;
  border: none;
  background: transparent;
  color: oklch(0.65 0 0);
  font-size: 11px;
  font-weight: 500;
  cursor: pointer;
  transition: all 150ms cubic-bezier(0.16, 1, 0.3, 1);
}

.segment-pill.active {
  background: oklch(0.24 0.01 260);
  color: oklch(0.95 0 0);
  font-weight: 600;
  box-shadow: 0 1px 4px oklch(0 0 0 / 30%);
}

.album-tab-content {
  display: flex;
  flex-direction: column;
  gap: 10px;
  min-height: 80px;
}

.input-group {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.input-label {
  font-size: 11px;
  font-weight: 500;
  color: oklch(0.7 0 0);
}

.suggestions-row {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.suggestions-label {
  font-size: 11px;
  color: oklch(0.5 0.02 260);
}

.sug-chip {
  padding: 3px 8px;
  border-radius: 12px;
  border: 1px solid oklch(1 0 0 / 12%);
  background: oklch(1 0 0 / 5%);
  color: oklch(0.8 0 0);
  font-size: 11px;
  cursor: pointer;
  transition: all 120ms cubic-bezier(0.16, 1, 0.3, 1);
}

.sug-chip:hover {
  background: oklch(1 0 0 / 10%);
}

.sug-chip.selected {
  background: oklch(0.6 0.17 162 / 20%);
  border-color: oklch(0.6 0.17 162);
  color: oklch(0.9 0.1 162);
}

.tab-note {
  font-size: 11px;
  color: oklch(0.55 0.02 260);
  line-height: 1.4;
}

.auto-album-card {
  display: flex;
  gap: 10px;
  padding: 10px 12px;
  border-radius: 10px;
  background: oklch(1 0 0 / 4%);
  border: 1px solid oklch(1 0 0 / 8%);
}

.auto-album-icon {
  color: oklch(0.7 0.17 162);
  margin-top: 2px;
}

.auto-album-text h4 {
  font-size: 12px;
  font-weight: 600;
  color: oklch(0.9 0 0);
  margin-bottom: 2px;
}

.auto-album-text p {
  font-size: 11px;
  color: oklch(0.55 0.02 260);
  line-height: 1.35;
}

.auto-album-text code {
  background: oklch(1 0 0 / 8%);
  padding: 1px 4px;
  border-radius: 4px;
  font-family: monospace;
}

/* Modal Footer */
.album-modal-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
  padding-top: 4px;
}

.cancel-btn {
  font-size: 12px !important;
  color: oklch(0.7 0 0) !important;
}

.confirm-album-btn {
  background: oklch(0.6 0.17 162) !important;
  color: white !important;
  font-size: 12px !important;
  font-weight: 600 !important;
  border-radius: 8px !important;
  transition: transform 120ms cubic-bezier(0.16, 1, 0.3, 1) !important;
}

.confirm-album-btn:active {
  transform: scale(0.96);
}
</style>
